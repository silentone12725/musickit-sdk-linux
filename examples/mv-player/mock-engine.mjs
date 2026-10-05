// A stand-in for the engine's /vseg endpoints, so the client in vseg-player.js can be
// tried (and its seek handling exercised) without an Apple account. It serves a synthetic
// H.264 test pattern made by ffmpeg, in both flavours the real engine produces:
//
//   FFmpeg path  timestamps restart at 0 after every seek; the client offsets by `t`
//   direct path  fragments keep a raw timeline 10 s ahead of the playlist's, and the
//                response states tsOffset (-10) instead
//
//   node mock-engine.mjs [port]            default 20125, plain http
//   MOCK_DIRECT=0       never offer the direct path
//   MOCK_BAD_DIRECT=1   offer it but report a wrong tsOffset, so the client's probe
//                       rejects it and falls back
import http from 'node:http';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

const port = Number(process.argv[2] ?? 20125);
const FRAG_SEC = 2, RAW_BASE = 10, SECONDS = 24;

// ── synthetic stream ──────────────────────────────────────────────────────────
const out = join(mkdtempSync(join(tmpdir(), 'mock-mv-')), 'mv.mp4');
const ff = spawnSync('ffmpeg', ['-hide_banner', '-loglevel', 'error', '-f', 'lavfi',
  '-i', 'testsrc2=size=640x360:rate=24', '-t', String(SECONDS), '-c:v', 'libx264', '-profile:v', 'high',
  '-level', '3.1', '-pix_fmt', 'yuv420p', '-g', String(FRAG_SEC * 24), '-keyint_min', String(FRAG_SEC * 24),
  '-sc_threshold', '0', '-movflags', 'frag_keyframe+empty_moov+default_base_moof', '-f', 'mp4', out]);
if (ff.status !== 0) { console.error('ffmpeg is required:', ff.stderr?.toString() || ff.error); process.exit(1); }

const file = readFileSync(out);
const boxes = [];
for (let o = 0; o < file.length;) { const size = file.readUInt32BE(o); boxes.push({ type: file.toString('latin1', o + 4, o + 8), o, size }); o += size; }
const moov = boxes.find((b) => b.type === 'moov');
const init = file.subarray(0, moov.o + moov.size);
const frags = [];
for (let i = 0; i < boxes.length; i++) {
  if (boxes[i].type === 'moof') frags.push(file.subarray(boxes[i].o, boxes[i + 1].o + boxes[i + 1].size));
}
const mdhd = init.indexOf('mdhd');
const timescale = init.readUInt32BE(mdhd + 4 + (init[mdhd + 4] === 0 ? 12 : 20));
const avcC = init.indexOf('avcC');
const codecs = 'avc1.' + [5, 6, 7].map((k) => init[avcC + k].toString(16).padStart(2, '0')).join('');
console.log(`synthetic stream: ${frags.length} fragments of ${FRAG_SEC}s, timescale ${timescale}, ${codecs}`);

/** A copy of fragment i with its tfdt moved by deltaSec. */
function shifted(i, deltaSec) {
  const f = Buffer.from(frags[i]);
  const at = f.indexOf('tfdt');
  const delta = Math.round(deltaSec * timescale);
  if (f[at + 4] === 1) f.writeBigUInt64BE(f.readBigUInt64BE(at + 8) + BigInt(delta), at + 8);
  else f.writeUInt32BE(f.readUInt32BE(at + 8) + delta, at + 8);
  return f;
}

// ── engine API ────────────────────────────────────────────────────────────────
const direct = process.env.MOCK_DIRECT !== '0';
const badDirect = process.env.MOCK_BAD_DIRECT === '1';
const state = { kind: 'base', start: 0, sbDirect: false }; // what /seg serves, and the init the player holds

const send = (res, code, body, type = 'application/json') => {
  res.writeHead(code, { 'Content-Type': type, 'Access-Control-Allow-Origin': '*', 'Access-Control-Allow-Headers': 'Content-Type, Range', 'Access-Control-Allow-Methods': 'GET, POST, DELETE, OPTIONS' });
  res.end(body);
};
const json = (res, obj) => send(res, 200, JSON.stringify(obj));

http.createServer((req, res) => {
  const url = new URL(req.url, 'http://x'), path = url.pathname;
  if (req.method === 'OPTIONS') return send(res, 204, '');
  if (req.method === 'POST' && path === '/api/v1/playback') return json(res, { id: 'mock', type: 'mv', durationMs: SECONDS * 1000, capabilities: { video: true } });
  if (req.method === 'DELETE') return send(res, 204, '');

  const m = path.match(/^\/api\/v1\/playback\/[^/]+\/vseg\/(manifest|init|seek|seg\/(\d+))$/);
  if (!m) return send(res, 404, 'not found', 'text/plain');
  switch (m[1].split('/')[0]) {
    case 'manifest': return json(res, { codecs, timescale, frags: [], done: true, err: false });
    case 'init': return send(res, 200, init, 'video/mp4');
    case 'seek': {
      const t = Number(url.searchParams.get('t'));
      const f = Math.min(frags.length - 1, Math.max(0, Math.floor(t / FRAG_SEC)));
      const wantDirect = direct && url.searchParams.get('direct') !== '0';
      state.kind = wantDirect ? 'direct' : 'ffmpeg';
      state.start = f;
      const reinit = state.sbDirect !== wantDirect;
      state.sbDirect = wantDirect;
      console.log(`seek t=${t} → ${state.kind} from fragment ${f}${reinit ? ' (reinit)' : ''}`);
      const resp = { n: 0, t: f * FRAG_SEC, reinit };
      if (wantDirect) Object.assign(resp, { direct: true, tsOffset: badDirect ? 0 : -RAW_BASE });
      return json(res, resp);
    }
    default: { // seg/{n}
      const i = state.start + Number(m[2]);
      if (i >= frags.length) return send(res, 404, 'end', 'text/plain');
      // direct: raw timeline (+10 s). FFmpeg flavour: restarts at 0 after a seek (base starts at 0 anyway).
      const body = state.kind === 'direct' ? shifted(i, RAW_BASE) : shifted(i, -state.start * FRAG_SEC);
      return send(res, 200, body, 'video/mp4');
    }
  }
}).listen(port, '127.0.0.1', () => console.log(`mock engine on http://127.0.0.1:${port}  (direct=${direct}${badDirect ? ', bad tsOffset' : ''})`));
