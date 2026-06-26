import { test, expect } from '@playwright/test';

test('waf settings page loads with status select and active/total badge', async ({ page }) => {
  await page.goto('/security/waf');
  await expect(page).toHaveURL(/security\/waf/);

  await expect(page.locator('#status')).toBeVisible();
  const sets = page.locator('div:has(span:text("Active:"))');
  await expect(sets.locator('span').nth(0)).toHaveText('Active:');
  await expect(sets.locator('span').nth(1)).toHaveText(/\d+\s*\/\s*\d+/);
  await expect(page.locator('#tour-waf-manage-rules')).toBeVisible();

  console.log('waf settings page loaded');
});

test('manage rules link navigates to waf rules page', async ({ page }) => {
  await page.goto('/security/waf');
  await page.locator('#tour-waf-manage-rules').click();
  await expect(page).toHaveURL(/security\/waf\/rules/);
  await expect(page.locator('#waf_sets')).toBeVisible();

  console.log('navigated to waf rules page');
});

test('search filters waf rule sets table', async ({ page }) => {
  await page.goto('/security/waf/rules');

  const rows = page.locator('#waf_sets tbody tr');
  const count = await rows.count();
  test.skip(count === 0, 'No WAF rule sets detected on this environment');

  const firstName = (await rows.first().locator('td').first().innerText()).trim();
  await page.locator('input[x-model="searchQuery"]').fill(firstName);
  await page.waitForTimeout(150);

  await expect(rows.filter({ hasText: firstName })).toBeVisible();

  console.log(`search filtered waf rule sets to "${firstName}"`);
});

test('view rule set opens raw rules in new tab', async ({ page, context }) => {
  await page.goto('/security/waf/rules');

  const rows = page.locator('#waf_sets tbody tr');
  const count = await rows.count();
  test.skip(count === 0, 'No WAF rule sets detected on this environment');

  const [newTab] = await Promise.all([
    context.waitForEvent('page'),
    rows.first().getByRole('link', { name: 'View' }).click(),
  ]);
  await newTab.waitForLoadState();
  await expect(newTab).toHaveURL(/security\/waf\/view-rules\?edit=/);

  console.log('view rules opened raw rules in new tab');
});

test('toggle a rule set enable/disable and revert', async ({ page }) => {
  await page.goto('/security/waf/rules');

  const rows = page.locator('#waf_sets tbody tr');
  const count = await rows.count();
  test.skip(count === 0, 'No WAF rule sets detected on this environment');

  const row = rows.first();

  const ruleName = await row.locator('td').first().innerText();

  const button = row.getByRole('button');
  const initialButtonText = (await button.innerText()).trim();

  await button.click();
  await expect(page.getByText(/Rules set (enabled|disabled)\. Restart Caddy to apply changes\./)).toBeVisible();

  const sameRow = page.locator('#waf_sets tbody tr').filter({ hasText: ruleName });

  const sameButton = sameRow.getByRole('button');

  await expect(sameButton).not.toHaveText(initialButtonText);

  await sameButton.click();
  await expect(page.getByText(/Rules set (enabled|disabled)\. Restart Caddy to apply changes\./)).toBeVisible();

  console.log(`toggled rule set "${ruleName}" and reverted`);
});
