import { test, expect } from '@playwright/test';

// NOTE: verify-only -- changing the default node IP/SSH key path affects where future
// user accounts get provisioned; submitting with test data could silently break new
// account creation, so we just confirm the form is wired correctly.

test('default node page loads with expected fields', async ({ page }) => {
  await page.goto('/server/node');
  await expect(page).toHaveURL(/server\/node/);

  await expect(page.locator('#default_node')).toBeVisible();
  await expect(page.locator('#default_ssh_key_path')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Save settings' })).toBeVisible();

  console.log('default node page loaded with expected fields, form not submitted');
});

test('default node fields are required', async ({ page }) => {
  await page.goto('/server/node');

  await expect(page.locator('#default_node')).toHaveAttribute('required', '');
  await expect(page.locator('#default_ssh_key_path')).toHaveAttribute('required', '');

  console.log('default node fields correctly required');
});
