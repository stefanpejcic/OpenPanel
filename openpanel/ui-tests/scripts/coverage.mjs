// every page template the Go code loads must show up in the gallery, or be listed in coverage-allow.txt
import { readFileSync, readdirSync, existsSync, statSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const repo = fileURLToPath(new URL('../../', import.meta.url));
const here = fileURLToPath(new URL('../', import.meta.url));
const walk = (d) => readdirSync(d).flatMap((f) => (statSync(join(d, f)).isDirectory() ? walk(join(d, f)) : f.endsWith('.go') && !f.endsWith('_test.go') ? [join(d, f)] : []));

const loaded = new Set();
for (const f of walk(join(repo, 'internal'))) {
  for (const call of readFileSync(f, 'utf8').matchAll(/[lL]oadPage\(([^)]*)\)/g)) {
    const files = [...call[1].matchAll(/"([^"]+\.html)"/g)].map((m) => m[1]).filter((t) => t !== 'base.html' && !t.startsWith('partials/'));
    // same rule as internal/web/gallery_on.go: a module's own _shared.html is a partial, not the page
    const page = files.find((t) => !t.split('/').pop().startsWith('_')) ?? (files.length === 1 && call[0].includes('"') && !/append\(/.test(call[0]) ? files[0] : null);
    if (page) loaded.add(page);
  }
}
const indexFile = join(here, 'gallery', 'index.jsonl');
if (!existsSync(indexFile)) { console.error('no gallery yet, run npm run gallery first'); process.exit(1); }
const captured = new Set(readFileSync(indexFile, 'utf8').trim().split('\n').filter(Boolean).map((l) => JSON.parse(l).template));
const allowFile = join(here, 'coverage-allow.txt');
const allowed = new Set(existsSync(allowFile) ? readFileSync(allowFile, 'utf8').split('\n').map((l) => l.trim()).filter((l) => l && !l.startsWith('#')) : []);

const missing = [...loaded].filter((t) => !captured.has(t)).sort();
const fresh = missing.filter((t) => !allowed.has(t));
const fixed = [...allowed].filter((t) => !missing.includes(t));
console.log(`${loaded.size} page templates, ${loaded.size - missing.length} in the gallery, ${missing.length} without a render test`);
if (process.argv.includes('--write-allow')) {
  writeFileSync(allowFile, '# page templates without a render test yet; remove a line once its test exists\n' + missing.join('\n') + '\n');
  console.log(`wrote ${missing.length} entries to coverage-allow.txt`);
  process.exit(0);
}
if (fixed.length) console.log(`now covered, remove from coverage-allow.txt:\n  ${fixed.join('\n  ')}`);
if (fresh.length) {
  console.error(`no render test for:\n  ${fresh.join('\n  ')}\nadd one (see internal/modules/*/render_test.go) so the page lands in the gallery`);
  process.exit(1);
}
