import { test, expect } from '@playwright/test';

test('update proxy and test restart needed msg', async ({ page }) => {
  // Go to settings page
  await page.goto('/settings/general');
  await expect(page).toHaveURL(/\/settings\/general/);

  // Update setting
  const input = page.getByRole('textbox', { name: /openpanel/i });
  await input.fill('newlink');

  await page.getByRole('button', { name: /save settings/i }).click();

  // Wait for UI updates after save
  await expect(page.getByText(/settings updated/i)).toBeVisible();
  await expect(page.getByText(/newlink/i)).toBeVisible();
  await expect(page.getByText(/2 services need restart/i)).toBeVisible();

  // Navigate to services needing restart
  await page.getByRole('link', { name: /services need restart/i }).click();
  await expect(page).toHaveURL(/\/services/);

  // Restart first service
  await page.getByRole('button', { name: /restart openpanel/i }).click();
  await expect(page.getByText(/1 service needs restart/i)).toBeVisible();

  // Restart second service
  await page.getByRole('button', { name: /restart admin/i }).click();

  // Expect failure message (can be improved later with network wait)
  await expect(page.getByText(/failed to restart/i)).toBeVisible();

  // Reload services page to ensure final state
  await page.goto('/services/');
  await expect(page).toHaveURL(/\/services/);

  // Ensure no restart-needed messages remain
  await expect(page.getByText(/need(s)? restart/i)).toHaveCount(0);
});
