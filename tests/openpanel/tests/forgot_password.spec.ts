import { test, expect } from '@playwright/test';

// TODO: test email!
test.use({ storageState: { cookies: [], origins: [] } });


test('reset password page loads', async ({ page }) => {
  await page.goto('/reset_password');
  await expect(page).toHaveURL(/reset_password/);
  await expect(page.locator('body')).toContainText(/reset|forgot|password/i);
  console.log('reset password page accessible');
});


test('submit empty email shows error', async ({ page }) => {
  await page.goto('/reset_password');

  const emailInput = page.locator('input[type="email"], input[name="email"]');
  await expect(emailInput).toBeVisible();

  await page.getByRole('button', { name: /reset|send|submit/i }).click();

  // browser validation or server-side error
  const isInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid);
  const hasError = await page.locator('body').textContent().then(t => /error|required|invalid/i.test(t ?? ''));

  expect(isInvalid || hasError).toBe(true);
  console.log('empty email correctly blocked');
});


test('submit invalid email format shows error', async ({ page }) => {
  await page.goto('/reset_password');

  await page.locator('input[type="email"], input[name="email"]').fill('not-an-email');
  await page.getByRole('button', { name: /reset|send|submit/i }).click();

  const emailInput = page.locator('input[type="email"], input[name="email"]');
  const isInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid);
  const hasError = await page.locator('body').textContent().then(t => /error|invalid/i.test(t ?? ''));

  expect(isInvalid || hasError).toBe(true);
  console.log('invalid email format correctly blocked');
});


test('submit valid email shows confirmation', async ({ page }) => {
  await page.goto('/reset_password');

  await page.locator('input[type="email"], input[name="email"]').fill('nonexistent@openpanel.com');
  await page.getByRole('button', { name: /reset|send|submit/i }).click();

  // should show "link sent" or similar, even for non-existent emails (security best practice)
  await expect(page.locator('body')).toContainText(/email.*sent|check.*email|link.*sent|if.*account.*exists/i, { timeout: 10000 });
  console.log('valid email submission shows confirmation');
});


test('invalid reset token shows error', async ({ page }) => {
  await page.goto('/reset_password/this-is-an-invalid-token-12345');
  await expect(page.locator('body')).toContainText(/invalid|expired|not found|error/i, { timeout: 10000 });
  console.log('invalid token correctly rejected');
});
