// renders a plain-text terminal transcript as a screenshot: node terminal-shot.mjs <input.txt> <output.png> [title]
// markup in the text: [g]green[/] [y]yellow[/] [r]red[/] [p]prompt[/]
import { chromium } from 'playwright';
import { readFileSync, mkdirSync } from 'fs';
import { dirname } from 'path';

const [input, output, title = 'root@server: ~'] = process.argv.slice(2);
if (!input || !output) { console.error('usage: node terminal-shot.mjs <input.txt> <output.png> [title]'); process.exit(1); }

const esc = s => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
const colors = { g: '#4ade80', y: '#facc15', r: '#f87171', p: '#60a5fa' };
const body = esc(readFileSync(input, 'utf8').replace(/\n$/, ''))
  .replace(/\[([gyrp])\]/g, (_, c) => `<span style="color:${colors[c]}">`)
  .replace(/\[\/\]/g, '</span>');

const html = `<!doctype html><html><head><meta charset="utf-8"><style>
  body { margin: 0; padding: 24px; background: #e5e7eb; }
  .win { background: #0b1020; border-radius: 10px; overflow: hidden; box-shadow: 0 8px 24px rgba(0,0,0,.25); }
  .bar { display: flex; align-items: center; gap: 8px; padding: 10px 14px; background: #1f2937; color: #9ca3af; font: 13px system-ui, sans-serif; }
  .dot { width: 12px; height: 12px; border-radius: 50%; }
  .t { flex: 1; text-align: center; margin-right: 52px; }
  pre { margin: 0; padding: 16px 18px; color: #e5e7eb; font: 12.5px/1.45 'DejaVu Sans Mono', Menlo, Consolas, monospace; white-space: pre-wrap; word-break: break-word; }
</style></head><body><div class="win"><div class="bar">
  <span class="dot" style="background:#ef4444"></span><span class="dot" style="background:#f59e0b"></span><span class="dot" style="background:#22c55e"></span>
  <span class="t">${esc(title)}</span></div><pre>${body}</pre></div></body></html>`;

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1100, height: 600 }, deviceScaleFactor: 2 });
await page.setContent(html);
mkdirSync(dirname(output), { recursive: true });
await page.locator('.win').screenshot({ path: output });
await browser.close();
console.log(`ok   ${output}`);
