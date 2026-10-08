// runs shell commands in a service's web terminal (Containers > Terminal) and screenshots the result
//
// usage: PANEL_URL=https://host:2083 node panel-terminal.mjs <service> <commands.sh> [out.png] [--full] [--width 1100] [--height 900]
//   <commands.sh>  one command per line, each waits for the prompt to come back (up to 10 minutes)
//   out.png        tabs bar down to the last line of output, or the whole window with --full
//   --width        viewport width; with the sidebar collapsed, 1500 fits about 170 columns
// the terminal output is also printed, so it can be checked before it's used in a post
import { chromium } from 'playwright';
import { readFileSync, mkdirSync } from 'node:fs';
import { dirname } from 'node:path';
import { parseArgs } from 'node:util';
import { BASE_URL, STATE_PATH } from './shots.mjs';

const { values: opt, positionals } = parseArgs({
  allowPositionals: true,
  options: { full: { type: 'boolean' }, width: { type: 'string', default: '1100' }, height: { type: 'string', default: '900' } },
});
const [service, cmdFile, out] = positionals;
if (!service || !cmdFile) {
  console.error('usage: node panel-terminal.mjs <service> <commands.sh> [out.png] [--full] [--width px] [--height px]');
  process.exit(1);
}

const browser = await chromium.launch();
const ctx = await browser.newContext({
  viewport: { width: +opt.width, height: +opt.height },
  deviceScaleFactor: 2,
  ignoreHTTPSErrors: true,
  storageState: STATE_PATH,
  locale: 'en-US',
});
const page = await ctx.newPage();
await page.addInitScript(() => localStorage.setItem('color-theme', 'light'));
await page.goto(`${BASE_URL}/containers/terminal/${service}?ctx=service`);
if (new URL(page.url()).pathname.startsWith('/login')) throw new Error('session expired, run `node login.mjs` again');
await page.waitForSelector('.xterm-rows');
await page.waitForTimeout(3000);

// collapse the sidebar so the terminal gets the full width, then let xterm refit
await page.locator('header button').first().click();
await page.waitForTimeout(800);
await page.evaluate(() => window.dispatchEvent(new Event('resize')));
await page.waitForTimeout(800);

const text = () => page.evaluate(() => document.querySelector('.xterm-rows').innerText);
await page.locator('.xterm').first().click();
for (const line of readFileSync(cmdFile, 'utf8').split('\n').filter(Boolean)) {
  await page.keyboard.type(line + '\n', { delay: 2 });
  await page.waitForTimeout(800);
  // done when the last line is a bare prompt again
  for (const start = Date.now(); Date.now() - start < 600_000; await page.waitForTimeout(1000)) {
    const rows = (await text()).replace(/\s+$/, '').split('\n');
    if (/^[#$]$/.test(rows[rows.length - 1].trim())) break;
  }
}
console.log(await text());

if (out) {
  mkdirSync(dirname(out), { recursive: true });
  await page.mouse.move(0, 0);
  const clip = opt.full
    ? undefined
    : await page.evaluate(() => {
        const rows = [...document.querySelectorAll('.xterm-rows > div')].filter(d => d.textContent.trim());
        const last = rows[rows.length - 1].getBoundingClientRect();
        const main = document.querySelector('main').getBoundingClientRect();
        const nav = document.querySelector('main nav, main [role=tablist]');
        const top = nav ? nav.getBoundingClientRect().top : main.top;
        return { x: main.x, y: top, width: main.width, height: last.bottom + 16 - top };
      });
  await page.screenshot({ path: out, clip, animations: 'disabled', caret: 'hide' });
  console.log(`ok   ${out}`);
}
await browser.close();
