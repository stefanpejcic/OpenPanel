import { test, expect } from '@playwright/test';

const domain = 'wp.tests.openpanel.org';


test('traffic stats page loads', async ({ page }) => {
  await page.goto('/domains/stats');
  await expect(page).toHaveURL(/domains\/stats/);
  await expect(page.locator('body')).toContainText(/stats|traffic|goaccess|analytics|visits/i);
  console.log('traffic stats page accessible');
});


test('traffic stats shows domain list', async ({ page }) => {
  await page.goto('/domains/stats');

  const hasDomainLinks = await page.locator('a[href*="stats"]').first().isVisible({ timeout: 5000 }).catch(() => false);
  const hasTable = await page.locator('table').isVisible({ timeout: 3000 }).catch(() => false);
  const hasDomainText = await page.locator('body').textContent().then(t => t?.includes(domain) ?? false);

  expect(hasDomainLinks || hasTable || hasDomainText).toBe(true);
  console.log('traffic stats shows domain entries');
});


//test('per-domain stats page loads', async ({ page }) => {
//  await page.goto(`/domains/stats/${domain}`);
//  await expect(page).toHaveURL(new RegExp(`domains/stats/${domain.replace('.', '\\.')}`));
//  await expect(page.locator('body')).toContainText(/stats|traffic|goaccess|analytics|visits|no data/i, { timeout: 30000 });
//  console.log(`per-domain stats for ${domain} accessible`);
//});


//test('per-domain stats shows report or no-data message', async ({ page }) => {
//  test.setTimeout(60000);
//  await page.goto(`/domains/stats/${domain}`);

//  // GoAccess renders an HTML report or a "no data" message
//  const hasReport = await page.locator('#init, #panels, .panel, #total_reqs').isVisible({ timeout: 5000 }).catch(() => false);
//  const hasNoData = await page.locator('body').textContent().then(t => /no data|no log|no traffic|empty/i.test(t ?? ''));
//  const hasIframe = await page.locator('iframe').isVisible({ timeout: 3000 }).catch(() => false);

//  expect(hasReport || hasNoData || hasIframe).toBe(true);
//  console.log(`per-domain stats: report=${hasReport}, noData=${hasNoData}, iframe=${hasIframe}`);
//});
