import { mkdir } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = fileURLToPath(new URL('.', import.meta.url));

// collects console errors and uncaught exceptions; failed requests are kept apart because gallery pages call JSON endpoints that don't exist here
export function watch(page) {
  const errors = [];
  const failedRequests = [];
  page.on('console', (m) => { if (m.type() === 'error' && !/Failed to load resource/.test(m.text())) errors.push(m.text()); });
  page.on('pageerror', (e) => errors.push(String(e)));
  page.on('response', (r) => { if (r.status() >= 400) failedRequests.push(`${r.status()} ${new URL(r.url()).pathname}`); });
  return { errors, failedRequests };
}

// widest element that sticks out past the viewport, or null when the page fits
export async function overflow(page) {
  return page.evaluate(() => {
    const vw = document.documentElement.clientWidth;
    if (document.documentElement.scrollWidth <= vw + 1) return null;
    let worst = null;
    for (const el of document.querySelectorAll('body *')) {
      const r = el.getBoundingClientRect();
      if (r.width && r.right > vw + 1 && (!worst || r.right > worst.right)) {
        const id = el.id ? `#${el.id}` : '';
        const cls = typeof el.className === 'string' && el.className ? `.${el.className.trim().split(/\s+/).slice(0, 3).join('.')}` : '';
        worst = { right: Math.round(r.right), selector: `${el.tagName.toLowerCase()}${id}${cls}` };
      }
    }
    return { scrollWidth: document.documentElement.scrollWidth, viewport: vw, worst };
  });
}

export async function shoot(page, testInfo, group, name) {
  const file = resolve(here, '..', 'out', testInfo.project.name, group, `${name}.png`);
  await mkdir(dirname(file), { recursive: true });
  await page.screenshot({ path: file, fullPage: true, animations: 'disabled' });
  await testInfo.attach(name, { path: file, contentType: 'image/png' });
}
