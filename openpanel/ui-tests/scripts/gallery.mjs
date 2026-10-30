// renders every page the Go render tests cover into ui-tests/gallery (needs Go, or run the same command in the golang image)
import { spawnSync } from 'node:child_process';
import { rmSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const out = fileURLToPath(new URL('../gallery/', import.meta.url));
const repo = fileURLToPath(new URL('../../', import.meta.url));
rmSync(out, { recursive: true, force: true });
const r = spawnSync('go', ['test', '-tags', 'gallery', '-count=1', './internal/...'], {
  cwd: repo, stdio: 'inherit', env: { ...process.env, OPENPANEL_GALLERY_DIR: out },
});
process.exit(r.status ?? 1);
