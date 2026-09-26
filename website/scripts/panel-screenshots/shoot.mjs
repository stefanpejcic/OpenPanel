// usage: node shoot.mjs [--admin] [page-key ...]   e.g. node shoot.mjs mysql/databases
import { chromium } from 'playwright';
import { existsSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
// --admin switches to the OpenAdmin demo (port 2087) and its own manifest
const admin = process.argv.includes('--admin');
const { BASE_URL, STATE_PATH, OUT_DIR, VIEWPORT_WIDTH, pages, globalPrepare } = await import(admin ? './admin-shots.mjs' : './shots.mjs');

if (!existsSync(STATE_PATH)) {
  console.error(`No session found, run \`node login.mjs${admin ? ' --admin' : ''}\` first.`);
  process.exit(1);
}

// --dark shoots the same pages in dark mode, saved next to the light ones with a _dark suffix
const dark = process.argv.includes('--dark');
const args = process.argv.slice(2).filter(a => a !== '--admin' && a !== '--dark');
const keys = args.length ? args : Object.keys(pages);
const browser = await chromium.launch();
const contextOptions = {
  // narrow enough that the content column is close to the docs column width, so UI text isn't shrunk much
  viewport: { width: VIEWPORT_WIDTH, height: 900 },
  deviceScaleFactor: 2,
  ignoreHTTPSErrors: true,
  colorScheme: dark ? 'dark' : 'light',
  locale: 'en-US',
};
const ctx = await browser.newContext({ ...contextOptions, storageState: STATE_PATH });
// logged-out context for the login and password reset pages
const anonCtx = await browser.newContext(contextOptions);

// never let a shot change anything on the demo server
for (const c of [ctx, anonCtx]) await c.route('**/*', route => (route.request().method() === 'GET' ? route.continue() : route.abort()));

// main column from the top down to the lowest visible leaf element, so min-height containers don't add empty space
async function contentBox(page, from = 'main') {
  return page.evaluate(from => {
    const main = document.querySelector('main');
    const start = document.querySelector(from) || main;
    const m = main.getBoundingClientRect();
    let bottom = 0;
    for (const el of main.querySelectorAll('*')) {
      if (el.children.length && !['BUTTON', 'SELECT', 'TEXTAREA', 'TABLE', 'svg', 'CANVAS', 'IMG'].includes(el.tagName)) continue;
      const r = el.getBoundingClientRect();
      const cs = getComputedStyle(el);
      if (!r.width || !r.height || cs.visibility === 'hidden' || cs.display === 'none' || cs.position === 'fixed') continue;
      bottom = Math.max(bottom, r.bottom);
    }
    const top = start.getBoundingClientRect().top;
    return { x: m.x, y: top, width: m.width, height: Math.min(bottom + 32, m.bottom) - top };
  }, from);
}

// union of the from/to element boxes plus padding, clamped to the content area
async function cropBox(page, crop) {
  if (crop === 'content') return contentBox(page);
  if (crop.content) {
    const box = await contentBox(page, crop.content);
    if (crop.maxHeight) box.height = Math.min(box.height, crop.maxHeight);
    return box;
  }
  const { from, to, pad = 0, fromTop = false, clamp = true, height, right = false } = crop;
  const boxes = [];
  for (const sel of [from, to].filter(Boolean)) {
    const box = await page.locator(sel).first().boundingBox();
    if (!box) throw new Error(`nothing visible for selector: ${sel}`);
    boxes.push(box);
  }
  // pages outside the panel layout (login, password reset) have no main, so fall back to the whole page
  const main = (await page.locator('main').first().count()) ? await page.locator('main').first().boundingBox() : { x: 0, y: 0, width: VIEWPORT_WIDTH, height: 1e5 };
  // fromTop: start at the top of main, and only the `to` element sets the bottom
  if (fromTop) boxes.splice(0, 1, { x: main.x, y: main.y, width: main.width, height: 0 });
  // clamp: false for content drawn outside main, like the GoAccess report
  const x1 = Math.max(clamp ? main.x : 0, Math.min(...boxes.map(b => b.x)) - pad);
  const y1 = Math.max(clamp ? main.y : 0, Math.min(...boxes.map(b => b.y)) - pad);
  // right: run to the right edge of main, for elements whose text is narrower than their card
  const x2 = right ? main.x + main.width : Math.min(main.x + main.width, Math.max(...boxes.map(b => b.x + b.width)) + pad);
  const y2 = height ? y1 + height : Math.max(...boxes.map(b => b.y + b.height)) + pad;
  return { x: x1, y: y1, width: x2 - x1, height: y2 - y1 };
}

let failed = 0;
for (const key of keys) {
  const def = pages[key];
  if (!def) {
    console.error(`unknown page: ${key}`);
    failed++;
    continue;
  }
  for (const shot of def.shots) {
    const page = await (def.anonymous ? anonCtx : ctx).newPage();
    try {
      // init runs before any page script, e.g. to set the theme in localStorage
      // every shot starts in light mode unless its init() says otherwise
      await page.addInitScript(theme => {
        localStorage.setItem('color-theme', theme);
        // a tour started by one shot must not carry over into the next
        for (const k of Object.keys(localStorage)) if (/tour/i.test(k)) localStorage.removeItem(k);
      }, dark ? 'dark' : 'light');
      if (def.init) await page.addInitScript(def.init);
      await page.goto(BASE_URL + (shot.url || def.url), { waitUntil: 'domcontentloaded' });
      if (!def.anonymous && new URL(page.url()).pathname.startsWith('/login')) throw new Error(`session expired, run \`node login.mjs${admin ? ' --admin' : ''}\` again`);
      await page.waitForLoadState('load');
      await page.waitForTimeout(800);
      // the page must be tall enough that crops below the fold still render
      // shots with fixed-height UI (drawers, editors) keep a set viewport instead of growing to the page height
      const height = shot.viewportHeight || Math.max(900, await page.evaluate(() => document.documentElement.scrollHeight));
      await page.setViewportSize({ width: shot.viewportWidth || VIEWPORT_WIDTH, height });
      // shot setup first, some pages look rows up by their real names while loading data
      // park the mouse first so a hover set up in prepare() survives
      await page.mouse.move(0, 0);
      // before: page-wide setup the shot's own steps depend on, like dropping rows that aren't fixtures
      if (def.before) await def.before(page);
      if (shot.before) await shot.before(page);
      if (shot.prepare) await shot.prepare(page);
      if (def.prepare) await def.prepare(page);
      await globalPrepare(page);
      // as: a second entry for the same docs page, like one shot on another server, writes the same files
      const out = join(OUT_DIR, `${def.as || key}-${shot.name}${dark ? '_dark' : ''}.png`);
      mkdirSync(dirname(out), { recursive: true });
      const clip = shot.crop ? await cropBox(page, shot.crop) : undefined;
      await page.screenshot({ path: out, clip, animations: 'disabled', caret: 'hide' });
      console.log(`ok   ${out}`);
    } catch (err) {
      failed++;
      console.error(`FAIL ${key}-${shot.name}: ${err.message}`);
    } finally {
      await page.close();
    }
  }
}
await browser.close();
process.exit(failed ? 1 : 0);
