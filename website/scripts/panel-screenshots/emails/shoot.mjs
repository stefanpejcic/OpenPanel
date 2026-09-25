// renders the notification emails with render.go and screenshots each one into the account docs images
// usage: node emails/shoot.mjs [path to openadmin/internal/webtemplates]
import { chromium } from 'playwright';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, readdirSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const templates = process.argv[2] || join(here, '../../../../../openadmin/internal/webtemplates');
const tmp = mkdtempSync(join(tmpdir(), 'op-emails-'));
const out = join(here, '../../../static/img/openpanel-screenshots/account');
const flags = join(here, '../../../../openpanel/static/flags');
console.log(execFileSync('go', ['run', join(here, 'render.go'), templates, tmp, flags], { encoding: 'utf8' }).trim());

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 720, height: 600 }, deviceScaleFactor: 2 });
for (const f of readdirSync(tmp).filter(f => f.endsWith('.html'))) {
  await page.goto('file://' + join(tmp, f));
  const file = join(out, 'notifications-email-' + f.replace('.html', '.png'));
  await page.locator('#mail').screenshot({ path: file });
  console.log('ok  ', file);
}
await browser.close();
