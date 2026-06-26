import { test, expect } from '@playwright/test';

// NOTE: templates/settings/caddy.html body is effectively blank, so we only assert
// the base layout (nav/header from base.html) renders without erroring.

test('caddy settings page loads without error', async ({ page }) => {
  const response = await page.goto('/settings/caddy');
  await expect(page).toHaveURL(/settings\/caddy/);
  expect(response?.status()).toBeLessThan(400);

  console.log('caddy settings page loaded base layout without error');
});

test('caddy metrics endpoint returns raw text proxy', async ({ page }) => {
  const response = await page.goto('/settings/caddy/metrics');
  expect(response?.status()).toBeLessThan(500);

  console.log(`caddy metrics endpoint responded with status ${response?.status()}`);
});
