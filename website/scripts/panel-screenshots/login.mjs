// signs in with the credentials prefilled on the demo login form and saves the session for shoot.mjs
// if a captcha widget is present it falls back to a visible browser so you can solve it by hand
import { chromium } from 'playwright';
import { mkdirSync } from 'node:fs';
// --admin logs in to the OpenAdmin demo (port 2087) instead
const { BASE_URL, STATE_PATH } = await import(process.argv.includes('--admin') ? './admin-shots.mjs' : './shots.mjs');

async function login(headless) {
  const browser = await chromium.launch({ headless });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, ignoreHTTPSErrors: true });
  const page = await ctx.newPage();
  await page.goto(`${BASE_URL}/login`);
  const hasCaptcha = (await page.locator('.cf-turnstile, [name="cf-turnstile-response"], .g-recaptcha').count()) > 0;
  if (headless && hasCaptcha) {
    await browser.close();
    return false;
  }
  if (hasCaptcha) console.log('Captcha detected, solve it and click Sign In...');
  else await page.click('button[type=submit]');
  await page.waitForURL(u => !u.pathname.startsWith('/login'), { timeout: hasCaptcha ? 600_000 : 30_000 });
  await ctx.storageState({ path: STATE_PATH });
  console.log(`Logged in, session saved to ${STATE_PATH}`);
  await browser.close();
  return true;
}

mkdirSync('.auth', { recursive: true });
if (!(await login(true))) await login(false);
