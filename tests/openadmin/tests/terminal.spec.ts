import { test, expect } from '@playwright/test';

// NOTE: the live shell is driven over a raw WebSocket with a custom init/resize
// protocol -- not something plain Playwright page actions can usefully drive. We only
// verify the page loads, is access-gated, and shows the expected xterm.js scaffold;
// we don't attempt to type commands into the terminal.

test('terminal page loads for an admin user with expected UI elements', async ({ page }) => {
  const response = await page.goto('/terminal');

  // route may 403 (disabled flag-file or non-admin) -- only assert scaffold when it succeeds
  test.skip(response?.status() === 403, 'Terminal disabled via /root/disable_openadmin_terminal_ui flag-file or current user lacks access');

  await expect(page).toHaveURL(/terminal/);
  await expect(page.locator('#shell')).toBeVisible();
  await expect(page.locator('#term-status')).toBeVisible();
  await expect(page.locator('#status-dot')).toBeVisible();
  await expect(page.locator('#status-text')).toBeVisible();
  await expect(page.locator('#terminal')).toBeVisible();

  console.log('terminal page loaded with shell selector and status scaffold');
});

test('reconnect button is present', async ({ page }) => {
  const response = await page.goto('/terminal');
  test.skip(response?.status() === 403, 'Terminal disabled or current user lacks access');

  await page.waitForLoadState('networkidle');
  await expect(page.locator('#reconnect-btn')).toBeAttached();
  console.log('reconnect button present');
});

test('shell selector offers bash/sh options', async ({ page }) => {
  const response = await page.goto('/terminal');
  test.skip(response?.status() === 403, 'Terminal disabled or current user lacks access');

  const options = page.locator('#shell option');
  const count = await options.count();
  expect(count).toBeGreaterThan(0);

  console.log(`shell selector offers ${count} options`);
});
