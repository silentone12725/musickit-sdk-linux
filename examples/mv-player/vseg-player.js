// A minimal MSE client for the engine's segmented music-video endpoints
// (/api/v1/playback/{id}/vseg/*). It shows the whole protocol:
//
//   start   POST /playback {capabilities:{video:true}} → session id
//   setup   GET  /vseg/manifest → codecs;  GET /vseg/init → append;  GET /vseg/seg/{n} → append
//   seek    GET  /vseg/seek?t=…  → {n, t, direct?, tsOffset?, reinit?}
//   end     /vseg/seg/{n} answers 404 past the last fragment → endOfStream()
//
// Video only: the soundtrack is a separate stream (/playback/{id}/audio) that a real
// player keeps in sync with the video clock.

const BUFFER_AHEAD = 12; // seconds to buffer ahead of the playhead

// Whether Chrome accepted the engine's fragment-level ("direct") output, per codec
// string. A rejected append kills a media element for good, so the first direct seek is
// tried on a throwaway MediaSource before the real one sees it.
const directVerdicts = {};

export class VsegPlayer {
  /**
   * @param {object} o
   * @param {string} o.engine  e.g. "https://127.0.0.1:20025"
   * @param {HTMLVideoElement} o.video
   * @param {(msg: string) => void} [o.log]
   */
  constructor({ engine, video, log = console.log }) {
    this.engine = engine.replace(/\/$/, '');
    this.video = video;
    this.log = log;
    this.id = null;
    this.sb = null;
    this.codecs = '';
    this.gen = 0; // bumped to cancel the running fetch loop
    this.abort = new AbortController();
    this.seeking = false;
  }

  get #api() { return `${this.engine}/api/v1`; }
  get #vseg() { return `${this.#api}/playback/${this.id}/vseg`; }

  async start(assetId, { storefront = '', mvMaxHeight = 0 } = {}) {
    const res = await fetch(`${this.#api}/playback`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ assetId, storefront, mvMaxHeight, capabilities: { video: true } }),
    });
    if (!res.ok) throw new Error(`POST /playback → HTTP ${res.status}: ${await res.text()}`);
    const session = await res.json();
    this.id = session.id;
    this.log(`session ${this.id}`);

    const ms = new MediaSource();
    this.ms = ms;
    this.video.src = URL.createObjectURL(ms);
    await new Promise((r) => ms.addEventListener('sourceopen', r, { once: true }));
    // Without the full duration the browser clamps a seek to the end of what is buffered.
    if (session.durationMs > 0) ms.duration = session.durationMs / 1000;

    this.codecs = await this.#codecs();
    this.sb = ms.addSourceBuffer(`video/mp4; codecs="${this.codecs}"`);
    this.sb.addEventListener('error', () => this.log('SourceBuffer error'));
    await this.#append(await this.#bytes(`${this.#vseg}/init`));
    this.video.addEventListener('seeking', () => this.#onSeeking());
    this.video.addEventListener('seeked', () => { this.seeking = false; });
    this.#fetchLoop(0);
  }

  async stop() {
    this.gen++;
    this.abort.abort();
    if (!this.id) return;
    await fetch(`${this.#vseg}`, { method: 'DELETE' }).catch(() => {});
    await fetch(`${this.#api}/playback/${this.id}`, { method: 'DELETE' }).catch(() => {});
    this.id = null;
  }

  // ── plumbing ────────────────────────────────────────────────────────────────

  async #codecs() {
    for (let i = 0; i < 10; i++) { // the producer may still be starting
      const m = await fetch(`${this.#vseg}/manifest`).then((r) => r.json()).catch(() => null);
      if (m?.codecs) return m.codecs;
      await new Promise((r) => setTimeout(r, 300));
    }
    throw new Error('no codec string from /vseg/manifest');
  }

  async #bytes(url, signal) {
    const r = await fetch(url, { signal });
    if (!r.ok) throw Object.assign(new Error(`${url} → HTTP ${r.status}`), { status: r.status });
    return r.arrayBuffer();
  }

  #idle(sb = this.sb) {
    return new Promise((res) => {
      if (!sb.updating) return res();
      sb.addEventListener('updateend', res, { once: true });
    });
  }

  async #append(buf) {
    await this.#idle();
    this.sb.appendBuffer(buf);
    await this.#idle();
  }

  async #fetchLoop(startN) {
    const gen = ++this.gen;
    let n = startN;
    try {
      while (gen === this.gen) {
        const b = this.sb.buffered;
        if (b.length && b.end(b.length - 1) - this.video.currentTime > BUFFER_AHEAD) {
          await new Promise((r) => setTimeout(r, 400)); // far enough ahead; wait
          continue;
        }
        let data;
        try {
          data = await this.#bytes(`${this.#vseg}/seg/${n}`);
        } catch (e) {
          if (e.status === 404) { // past the last fragment
            if (this.ms.readyState === 'open') this.ms.endOfStream();
            this.log('end of stream');
            return;
          }
          throw e;
        }
        if (gen !== this.gen) return;
        await this.#append(data);
        n++;
      }
    } catch (e) {
      if (e.name !== 'AbortError') this.log(`fetch loop stopped: ${e.message}`);
    }
  }

  // ── seeking ─────────────────────────────────────────────────────────────────

  async #onSeeking() {
    if (this.seeking) return; // remove() below makes the browser fire 'seeking' again
    const t = this.video.currentTime;
    const b = this.sb.buffered;
    for (let i = 0; i < b.length; i++) if (b.start(i) <= t && t < b.end(i)) return; // browser handles it
    this.seeking = true;
    this.gen++; // stop fetching from the old position
    this.log(`seek to ${t.toFixed(2)}s`);

    const seekUrl = (noDirect) => `${this.#vseg}/seek?t=${t}${noDirect ? '&direct=0' : ''}`;
    const ask = (noDirect) => fetch(seekUrl(noDirect)).then((r) => r.json());

    let resp = await ask(directVerdicts[this.codecs] === false);

    // First fragment-level seek: have Chrome vet it before the real SourceBuffer does.
    if (resp.direct && directVerdicts[this.codecs] === undefined) {
      const ok = await this.#probeDirect(resp);
      directVerdicts[this.codecs] = ok;
      if (!ok) {
        this.log('direct output rejected — using the FFmpeg path from now on');
        resp = await ask(true);
      }
    }
    this.log(`seek → ${JSON.stringify(resp)}`);

    // Drop whatever is buffered; the new data replaces it.
    await this.#idle();
    const b2 = this.sb.buffered;
    if (b2.length) {
      this.sb.remove(0, b2.end(b2.length - 1) + 0.001);
      await this.#idle();
    }

    // Where do the new fragments sit on the player's timeline?
    //   direct: they keep the stream's raw timeline; the engine states the offset (negative).
    //   n === 0 (FFmpeg seek producer): timestamps restart at 0, so offset by t.
    //   n  > 0 (served from the from-0 producer): already absolute, no offset.
    this.sb.timestampOffset = typeof resp.tsOffset === 'number' ? resp.tsOffset : (resp.n === 0 ? resp.t : 0);

    // The engine switches between the FFmpeg init segment and the original one;
    // the SourceBuffer needs the matching init before any fragment.
    if (resp.reinit) {
      await this.#append(await this.#bytes(`${this.#vseg}/init`));
      this.log(`re-initialised (${resp.direct ? 'direct' : 'ffmpeg'} init)`);
    }
    this.#fetchLoop(resp.n);
  }

  /** Append init + first fragment to a throwaway MediaSource and check placement. */
  async #probeDirect(resp) {
    let url, vid;
    try {
      const [init, seg] = await Promise.all([
        this.#bytes(`${this.#vseg}/init`), this.#bytes(`${this.#vseg}/seg/0`)]);
      const ms = new MediaSource();
      vid = document.createElement('video');
      url = URL.createObjectURL(ms);
      vid.src = url;
      await new Promise((res, rej) => {
        ms.addEventListener('sourceopen', res, { once: true });
        setTimeout(() => rej(new Error('sourceopen timeout')), 3000);
      });
      const sb = ms.addSourceBuffer(`video/mp4; codecs="${this.codecs}"`);
      sb.timestampOffset = resp.tsOffset;
      const append = (buf) => new Promise((res, rej) => {
        sb.addEventListener('error', () => rej(new Error('append error')), { once: true });
        sb.addEventListener('updateend', res, { once: true });
        sb.appendBuffer(buf);
      });
      await append(init);
      await append(seg);
      if (!sb.buffered.length) throw new Error('nothing buffered');
      const start = sb.buffered.start(0);
      if (Math.abs(start - resp.t) > 0.5) throw new Error(`placed at ${start.toFixed(2)}s, engine said ${resp.t.toFixed(2)}s`);
      this.log(`direct probe ok (buffered ${start.toFixed(2)}s)`);
      return true;
    } catch (e) {
      this.log(`direct probe failed: ${e.message}`);
      return false;
    } finally {
      try { vid?.removeAttribute('src'); vid?.load(); } catch {}
      if (url) URL.revokeObjectURL(url);
    }
  }
}
