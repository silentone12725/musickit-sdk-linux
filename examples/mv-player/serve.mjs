// Serves this directory on http://127.0.0.1:8080 (or the port given as the first
// argument). The engine accepts browser requests only from https://music.apple.com or a
// loopback origin, so the page must be served from 127.0.0.1 / localhost.
//   node serve.mjs [port]
import http from 'node:http';
import { readFile } from 'node:fs/promises';
import { extname, join, normalize, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = resolve(fileURLToPath(new URL('.', import.meta.url)));
const types = { '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css' };
const port = Number(process.argv[2] ?? 8080);

http.createServer(async (req, res) => {
  const path = new URL(req.url, 'http://x').pathname;
  const file = resolve(join(root, normalize(path === '/' ? '/index.html' : path)));
  if (!file.startsWith(root + '/') || !types[extname(file)]) { res.writeHead(404).end('not found'); return; }
  try {
    res.writeHead(200, { 'Content-Type': types[extname(file)] }).end(await readFile(file));
  } catch { res.writeHead(404).end('not found'); }
}).listen(port, '127.0.0.1', () => console.log(`http://127.0.0.1:${port}/`));
