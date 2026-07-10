import { test, expect } from '@playwright/test';

test('logout', async ({ page }) => {
  await page.goto('/dashboard');

  await page.locator('#user-btn-info').click();
  await page.getByRole('link', { name: 'Sign out' }).click();
  await expect(page).toHaveURL(/.*login/);

  // validate
  await page.goto('/dashboard');
  await expect(page).toHaveURL(/.*login/);

  console.log('logout terminated session');
});
