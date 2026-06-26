import { test, expect } from '@playwright/test';

// NOTE: verify-only on save -- modules gate real OpenPanel features; saving with a
// random toggle flipped could disable something other tests or users depend on.

test('modules page loads with grid, filters, and save button', async ({ page }) => {
  await page.goto('/settings/modules');
  await expect(page).toHaveURL(/settings\/modules/);

  await expect(page.locator('#tour-modules-grid')).toBeVisible();
  await expect(page.locator('#tour-modules-filters')).toBeVisible();
  await expect(page.locator('#tour-bulk-modules')).toBeVisible();
  await expect(page.locator('#tour-save-modules-btn')).toBeVisible();

  console.log('modules page loaded with grid and controls');
});

test('search filters the modules grid', async ({ page }) => {
  await page.goto('/settings/modules');

  const cards = page.locator('#tour-modules-grid [data-row]');
  const count = await cards.count();
  test.skip(count === 0, 'No modules registered on this environment');

  const firstName = (await cards.first().locator('h4').innerText()).trim();
  await page.locator('#tour-modules-filters input[x-model="searchQuery"]').fill(firstName);
  await page.waitForTimeout(150);

  await expect(cards.filter({ hasText: firstName })).toBeVisible();
  console.log(`search filtered modules grid to "${firstName}"`);
});

test('category filter buttons toggle active state', async ({ page }) => {
  await page.goto('/settings/modules');

  const communityBtn = page.getByRole('button', { name: 'Community' });
  await communityBtn.click();
  await expect(communityBtn).toHaveClass(/bg-indigo-600/);

  await communityBtn.click();
  await expect(communityBtn).not.toHaveClass(/bg-indigo-600/);

  console.log('category filter button toggled active state');
});

test('a module toggle switch responds to click without saving', async ({ page }) => {
  await page.goto('/settings/modules');

  const cards = page.locator('#tour-modules-grid [data-row]');
  const count = await cards.count();
  test.skip(count === 0, 'No modules registered on this environment');

  const checkbox = cards.first().locator('input[type="checkbox"]');
  const lockedCard = await cards.first().getAttribute('data-locked');
  test.skip(lockedCard !== null, 'First module is a locked plugin card with no toggle');

  await expect(checkbox).toBeAttached();

  const initial = await checkbox.isChecked();

  // toggle ON
  await checkbox.check({ force: true });
  expect(await checkbox.isChecked()).toBe(!initial);

  // toggle OFF (restore original state)
  await checkbox.check({ force: true });
  expect(await checkbox.isChecked()).toBe(initial);

  console.log('module toggle switch responded to toggles (not saved)');
});

test('enable all / disable all bulk buttons affect toggle states', async ({ page }) => {
  await page.goto('/settings/modules');

  const checkboxes = page.locator('#tour-modules-grid [data-row]:not([data-locked]) input[type="checkbox"]');
  const count = await checkboxes.count();
  test.skip(count === 0, 'No unlocked modules to bulk toggle');

  await page.getByRole('button', { name: 'Disable All' }).click();
  await expect(checkboxes.first()).not.toBeChecked();

  await page.getByRole('button', { name: 'Enable All' }).click();
  await expect(checkboxes.first()).toBeChecked();

  console.log('bulk enable/disable buttons updated toggle states (not saved)');
});
