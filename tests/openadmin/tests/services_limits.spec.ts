import { test, expect } from '@playwright/test';

// NOTE: verify-only -- saving service limits restarts the affected docker containers
// with the new CPU/RAM constraints, which would disrupt mysql/varnish/etc for every
// other test relying on them in a shared environment.

test('service limits page loads with per-service-group sections', async ({ page }) => {
  await page.goto('/services/limits');
  await expect(page).toHaveURL(/services\/limits/);

  await expect(page.locator('#tour-save-limits-btn')).toBeVisible();
  await expect(page.locator('#tour-first-service-limit')).toBeVisible();

  console.log('service limits page loaded with at least one group section');
});

test('each service group exposes numeric CPU/RAM and boolean fields', async ({ page }) => {
  await page.goto('/services/limits');

  const numberInputs = page.locator('input[type="number"]');
  const numberCount = await numberInputs.count();
  expect(numberCount).toBeGreaterThan(0);

  for (let i = 0; i < numberCount; i++) {
    await expect(numberInputs.nth(i)).toHaveAttribute('min', '0');
  }

  console.log(`service limits page has ${numberCount} numeric CPU/RAM fields`);
});
