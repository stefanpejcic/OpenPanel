import { test, expect } from '@playwright/test';

test('php settings page loads with options textarea and ini accordions', async ({ page }) => {
  await page.goto('/settings/php');
  await expect(page).toHaveURL(/settings\/php/);

  await expect(page.locator('textarea[name="options"]')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Save Options' })).toBeVisible();
  await expect(page.locator('#tour-first-ini-toggle')).toBeVisible();

  console.log('php settings page loaded');
});

test('edit and save available php options, then revert', async ({ page }) => {
  await page.goto('/settings/php');

  const textarea = page.locator('textarea[name="options"]');
  const original = await textarea.inputValue();
  const comment = `\n; test-comment-${Date.now()}`;

  await textarea.fill(original + comment);
  await page.getByRole('button', { name: 'Save Options' }).click();
  await page.waitForLoadState();
  await expect(page.locator('textarea[name="options"]')).toHaveValue(new RegExp(comment.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')));

  // revert
  await page.locator('textarea[name="options"]').fill(original);
  await page.getByRole('button', { name: 'Save Options' }).click();
  await page.waitForLoadState();
  await expect(page.locator('textarea[name="options"]')).toHaveValue(original);

  console.log('php options edited, saved, and reverted');
});

test('first php.ini accordion expands and collapses', async ({ page }) => {
  await page.goto('/settings/php');

  const toggle = page.locator('#tour-first-ini-toggle');
  const content = page.locator('#tour-first-ini-content');

  await expect(content).toBeHidden();
  await toggle.click();
  await expect(content).toBeVisible();
  await expect(content.locator('textarea')).not.toHaveValue('');

  await toggle.click();
  await expect(content).toBeHidden();

  console.log('first php.ini accordion expand/collapse working');
});

test('edit, save, and revert the first php.ini file', async ({ page }) => {
  await page.goto('/settings/php');

  await page.locator('#tour-first-ini-toggle').click();
  const textarea = page.locator('#tour-first-ini-content textarea');
  await expect(textarea).toBeVisible();

  const original = await textarea.inputValue();
  const comment = `\n; test-comment-${Date.now()}`;

  await textarea.fill(original + comment);
  await page.locator('#tour-first-ini-buttons').getByRole('button', { name: 'Save' }).click();
  await page.waitForLoadState();

  await page.locator('#tour-first-ini-toggle').click();
  await expect(page.locator('#tour-first-ini-content textarea')).toHaveValue(new RegExp(comment.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')));

  // revert
  await page.locator('#tour-first-ini-content textarea').fill(original);
  await page.locator('#tour-first-ini-buttons').getByRole('button', { name: 'Save' }).click();
  await page.waitForLoadState();

  console.log('first php.ini file edited, saved, and reverted');
});

test('restore default fetches ini content from GitHub', async ({ page }) => {
  await page.goto('/settings/php');

  await page.locator('#tour-first-ini-toggle').click();
  const textarea = page.locator('#tour-first-ini-content textarea');
  const original = await textarea.inputValue();

  await textarea.fill('placeholder-to-be-replaced');
  await page.locator('#tour-first-ini-buttons').getByRole('button', { name: 'Restore Default' }).click();

  await expect(textarea).not.toHaveValue('placeholder-to-be-replaced', { timeout: 15_000 });

  // revert without saving the fetched defaults
  await textarea.fill(original);
  console.log('restore default fetched php.ini content from GitHub');
});
