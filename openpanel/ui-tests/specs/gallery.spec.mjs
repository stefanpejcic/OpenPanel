import { test, expect } from '@playwright/test';
import { existsSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { watch, overflow, shoot } from './checks.mjs';

const dir = fileURLToPath(new URL('../gallery/', import.meta.url));
const index = existsSync(`${dir}index.jsonl`)
  ? readFileSync(`${dir}index.jsonl`, 'utf8').trim().split('\n').filter(Boolean).map((l) => JSON.parse(l))
  : [];
// one render per template unless GALLERY_ALL=1, the render tests often cover the same page several times
const seen = new Set();
// identical renders share a file name, so GALLERY_ALL still dedupes by file
const entries = index.filter((e) => { const k = process.env.GALLERY_ALL ? e.file : e.template; return !seen.has(k) && seen.add(k); });
// templates already moved to the new design; only these must pass the layout checks, the rest are screenshot-only until they migrate
const migrated = new Set(readFileSync(fileURLToPath(new URL('../migrated.txt', import.meta.url)), 'utf8').split('\n').map((l) => l.trim()).filter((l) => l && !l.startsWith('#')));

test.skip(entries.length === 0, 'no gallery yet, run npm run gallery first');

for (const e of entries) {
  const strict = migrated.has(e.template);
  test(`gallery ${e.template} ${e.file.split('--')[1].replace('.html', '')}`, async ({ page }, testInfo) => {
    testInfo.annotations.push({ type: 'route', description: e.path || '(none)' });
    const watched = watch(page);
    await page.goto(`/gallery/${e.file}`);
    await page.waitForLoadState('load');
    const over = await overflow(page);
    if (strict) {
      // the gallery has no backend, so a page that logs its failed fetch or websocket isn't broken; anything else still fails
      const offline = watched.errors.filter((err) => (watched.failedRequests.length && /Unexpected end of JSON input|is not valid JSON|Unexpected response/.test(err)) || /^WebSocket connection to .* failed/.test(err));
      for (const err of offline) testInfo.annotations.push({ type: 'no backend', description: err.slice(0, 300) });
      expect(watched.errors.filter((err) => !offline.includes(err)), 'console errors').toEqual([]);
      expect(over, 'page scrolls sideways').toBeNull();
    } else {
      if (over) testInfo.annotations.push({ type: 'overflow', description: `${over.scrollWidth}px wide at ${over.viewport}px, widest: ${over.worst?.selector}` });
      for (const err of watched.errors) testInfo.annotations.push({ type: 'console', description: err.slice(0, 300) });
    }
    await shoot(page, testInfo, 'gallery', e.file.replace('.html', ''));
  });
}
