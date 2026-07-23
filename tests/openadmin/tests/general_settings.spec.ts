import { test, expect } from '@playwright/test';

test('update proxy and test restart needed msg', async ({ page }) => {
  await page.goto('/settings/general');
  await expect(page).toHaveURL(/\/settings\/general/);

  // Update setting
  const redirectInput = page.getByRole('textbox', { name: /openpanel/i });
  await redirectInput.fill('newlink');

  await page.getByRole('button', { name: /save settings/i }).click();

  // Toast — scope it so it can't match sidebar/nav text
  await expect(page.getByRole('alert')).toContainText(/settings updated/i);

  // Input value, not text content
  await expect(redirectInput).toHaveValue('newlink');

  // Restart banner is a link; assert on the role, tolerate 1 or 2
  const restartLink = page.getByRole('link', { name: /services? needs? restart/i });
  await expect(restartLink).toBeVisible();

  await restartLink.click();
  await expect(page).toHaveURL(/\/services/);

  // Restart first service
  await page.getByRole('button', { name: /restart openpanel/i }).click();
  await expect(page.getByRole('link', { name: /1 service needs restart/i })).toBeVisible();

  // Restart second service
  await page.getByRole('button', { name: /restart admin/i }).click();
  await expect(page.getByRole('alert')).toContainText(/failed to restart/i);

  // Final state
  await page.goto('/services/');
  await expect(page).toHaveURL(/\/services/);
  await expect(page.getByRole('link', { name: /needs? restart/i })).toHaveCount(0);
});
