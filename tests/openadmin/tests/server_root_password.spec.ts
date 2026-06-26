import { test, expect } from '@playwright/test';

test('root password page loads with expected fields', async ({ page }) => {
  await page.goto('/server/root-password');

  const passwordField = page.locator('#password');

  await expect(passwordField).toBeVisible();
  await expect(passwordField).toHaveAttribute('name', 'password');
  await expect(passwordField).toHaveAttribute('minlength', '6');
  await expect(passwordField).toHaveAttribute('maxlength', '30');
  await expect(passwordField).toHaveAttribute('required', '');

  await expect(page.getByRole('button', { name: 'Change Password' })).toBeVisible();
});

test('password field rejects apostrophe via pattern attribute', async ({ page }) => {
  await page.goto('/server/root-password');

  const passwordField = page.locator('#password');
  await expect(passwordField).toHaveAttribute('pattern', "[^']{6,30}");

  console.log('password field pattern correctly excludes apostrophes');
});
