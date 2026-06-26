import { test, expect } from '@playwright/test';

test('locales page loads with table', async ({ page }) => {
  await page.goto('/settings/locales');
  await expect(page).toHaveURL(/settings\/locales/);
  await expect(page.locator('#tour-locales-table')).toBeVisible();

  console.log('locales page loaded');
});

test('search filters the locales table', async ({ page }) => {
  await page.goto('/settings/locales');

  const rows = page.locator('#tour-locales-table tbody tr');
  const count = await rows.count();
  test.skip(count === 0, 'No locales available on this environment');

  const firstLocale = (await rows.first().locator('td').nth(1).innerText()).trim();
  await page.locator('input[x-model="searchQuery"]').fill(firstLocale);
  await page.waitForTimeout(150);

  await expect(rows.filter({ hasText: firstLocale })).toBeVisible();
  console.log(`search filtered locales table to "${firstLocale}"`);
});

test('set a different installed locale as default, then revert', async ({ page }) => {
  await page.goto('/settings/locales');

  const setDefaultButtons = page.getByRole('button', { name: 'Set as Default' });
  const count = await setDefaultButtons.count();
  test.skip(count === 0, 'Only one (or zero) installed locale available');

  const rows = page.locator('#tour-locales-table tbody tr');

  const defaultRow = rows.filter({ hasText: 'Default' }).first();
  const originalLocale = (await defaultRow.locator('td').nth(1).innerText()).trim();

  const firstNonDefaultButton = setDefaultButtons.first();

  await firstNonDefaultButton.click();

  // wait for UI to reflect change instead of assuming immediate DOM stability
  await expect(rows.filter({ hasText: 'Default' })).not.toHaveText(originalLocale);

  // now find the new "Set as Default" button for the original row
  const restoreButton = rows
    .filter({ hasText: originalLocale })
    .getByRole('button', { name: 'Set as Default' });

  await expect(restoreButton).toBeVisible();
  await restoreButton.click();

  await expect(
    rows.filter({ hasText: originalLocale }).getByText('Default')
  ).toBeVisible();

  console.log(`switched default locale away from "${originalLocale}" and back`);
});

test('install button present for non-installed locales (not clicked)', async ({ page }) => {
  // NOTE: verify-only -- installing a locale is a real, possibly slow package operation;
  // we don't want every test run to install/download locale files on the host.
  await page.goto('/settings/locales');

  const installButtons = page.getByRole('button', { name: /Install/ });
  const count = await installButtons.count();
  test.skip(count === 0, 'All locales already installed on this environment');

  await expect(installButtons.first()).toBeVisible();
  console.log(`found ${count} not-yet-installed locales with an Install action`);
});
