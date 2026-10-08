// renders a 1500x800 blog featured image in the style of the existing covers in static/img/blog:
// navy background, OpenPanel logo, big uppercase title and a screenshot in a slanted panel on the right
//
// usage: node blog-cover.mjs --title "SLOW QUERY|ANALYSIS" --shot <screenshot.png> --out <cover.png>
//          [--scale 1] [--x 0] [--y 0]
//   --title  lines separated by |, uppercased; a line too wide for the left side is shrunk to fit
//   --shot   any screenshot, e.g. one taken with shoot.mjs; it fills the panel from its top-left corner
//   --scale  how much to scale the screenshot, at least enough to fill the panel (the default)
//   --x/--y  where in the screenshot the panel starts, in screenshot pixels
import { chromium } from 'playwright';
import { readFileSync, mkdirSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { parseArgs } from 'node:util';

const { values: opt } = parseArgs({
  options: { title: { type: 'string' }, shot: { type: 'string' }, out: { type: 'string' }, scale: { type: 'string' }, x: { type: 'string', default: '0' }, y: { type: 'string', default: '0' } },
});
if (!opt.title || !opt.shot || !opt.out) {
  console.error('usage: node blog-cover.mjs --title "LINE ONE|LINE TWO" --shot <screenshot.png> --out <cover.png> [--scale n] [--x px] [--y px]');
  process.exit(1);
}

const here = dirname(fileURLToPath(import.meta.url));
const dataUri = (path, type) => `data:${type};base64,${readFileSync(path).toString('base64')}`;
const esc = s => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

// measured from the existing covers
const NAVY = '#051d40';
const BLUE = '#145da0';
const PANEL = { left: 835, top: 48, width: 585, height: 700, slant: 160 };
const TITLE = { center: 580, maxWidth: 720, capTop: 299, capHeight: 85, pitch: 135 };
// diamond centers, top right cluster and bottom left grid
const DIAMONDS = [[922, 23], [1013, 23], [1104, 23], [967, 69], [922, 115], [27, 682], [119, 682], [210, 682], [73, 728], [165, 728], [27, 773], [119, 773], [210, 773]];

const lines = opt.title.split('|').map(l => l.trim().toUpperCase());
const html = `<!doctype html><html><head><meta charset="utf-8"><style>
  @font-face { font-family: Spartan; src: url(${dataUri(`${here}/blog-cover/LeagueSpartan.ttf`, 'font/ttf')}); font-weight: 100 900; }
  * { margin: 0; padding: 0; }
  body { width: 1500px; height: 800px; background: ${NAVY}; position: relative; overflow: hidden; }
  .stripe { position: absolute; left: 1315px; top: 0; width: 150px; height: 800px; background: ${BLUE}; }
  .d { position: absolute; width: 15px; height: 15px; background: ${BLUE}; transform: translate(-50%, -50%) rotate(45deg); }
  .panel { position: absolute; left: ${PANEL.left}px; top: ${PANEL.top}px; width: ${PANEL.width}px; height: ${PANEL.height}px; overflow: hidden;
    clip-path: polygon(${PANEL.slant}px 0, 100% 0, 100% 100%, 0 100%); background: #fff; }
  .panel img { position: absolute; left: 0; top: 0; transform-origin: 0 0; }
  .logo { position: absolute; left: 160px; top: 85px; }
  .title { position: absolute; left: 0; width: ${TITLE.center * 2}px; text-align: center; font: 700 120px/${TITLE.pitch}px Spartan; color: #fffbfb; white-space: nowrap; }
  .title div { display: block; }
  .title span { display: inline-block; }
</style></head><body>
  <div class="stripe"></div>
  ${DIAMONDS.map(([x, y]) => `<div class="d" style="left:${x}px;top:${y}px"></div>`).join('')}
  <div class="panel"><img src="${dataUri(resolve(opt.shot), 'image/png')}"></div>
  <img class="logo" src="${dataUri(`${here}/blog-cover/openpanel-logo-white.png`, 'image/png')}">
  <div class="title">${lines.map(l => `<div><span>${esc(l)}</span></div>`).join('')}</div>
</body></html>`;

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1500, height: 800 } });
await page.setContent(html);
await page.evaluate(() => document.fonts.ready);
await page.evaluate(({ PANEL, TITLE, scale, x, y }) => {
  // screenshot: never smaller than the panel and never shifted past its edges, so no gaps show
  const img = document.querySelector('.panel img');
  const cover = Math.max(PANEL.width / img.naturalWidth, PANEL.height / img.naturalHeight);
  const s = Math.max(scale ? +scale : cover, cover);
  const sx = Math.min(Math.max(x, 0), img.naturalWidth - PANEL.width / s);
  const sy = Math.min(Math.max(y, 0), img.naturalHeight - PANEL.height / s);
  img.style.transform = `scale(${s}) translate(${-sx}px, ${-sy}px)`;

  // title: size the font so capitals are TITLE.capHeight tall, shrink lines that don't fit
  const title = document.querySelector('.title');
  const probe = document.createElement('canvas').getContext('2d');
  probe.font = '700 120px Spartan';
  const m = probe.measureText('H');
  const capRatio = m.actualBoundingBoxAscent / 120;
  const size = TITLE.capHeight / capRatio;
  title.style.fontSize = `${size}px`;
  for (const span of title.querySelectorAll('span')) {
    const w = span.getBoundingClientRect().width;
    if (w > TITLE.maxWidth) span.style.fontSize = `${(size * TITLE.maxWidth) / w}px`;
  }
  // put the top of the first line's capitals at TITLE.capTop (the 5px corrects for what the font metrics report)
  const first = title.querySelector('div');
  const half = (TITLE.pitch - size) / 2;
  const ascent = (probe.measureText('H').fontBoundingBoxAscent / 120) * size;
  title.style.top = `${TITLE.capTop - 5 - (first.offsetTop + half + ascent - capRatio * size)}px`;
}, { PANEL, TITLE, scale: opt.scale, x: +opt.x, y: +opt.y });

mkdirSync(dirname(resolve(opt.out)), { recursive: true });
await page.screenshot({ path: opt.out });
await browser.close();
console.log(`ok   ${opt.out}`);
