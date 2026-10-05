// Node 20+ client for a running engine:  node client.mjs [port]
// Allow the engine's self-signed loopback certificate for this script only.
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

const port = process.argv[2] ?? '20025';
const base = `https://127.0.0.1:${port}/api/v1`;
const get = async (path) => (await fetch(base + path)).json();

console.log('status      ', await get('/status'));
console.log('capabilities', await get('/capabilities'));

// Live engine events (DRM state, playback, export progress)
const res = await fetch(base + '/events');
const decoder = new TextDecoder();
let seen = 0;
for await (const chunk of res.body) {
  process.stdout.write(decoder.decode(chunk));
  if (++seen >= 3) break;
}
