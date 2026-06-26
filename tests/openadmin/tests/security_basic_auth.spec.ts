import { test, expect } from '@playwright/test';

test('basic auth settings page loads with expected fields', async ({ page }) => {
  await page.goto('/security/basic_auth');
  await expect(page).toHaveURL(/security\/basic_auth/);

  await expect(page.locator('#basic_auth')).toBeVisible();
  await expect(page.locator('[name="basic_auth_username"]')).toBeVisible();
  await expect(page.locator('#basic_auth_password')).toBeVisible();
  await expect(page.locator('#generatePassword')).toBeVisible();
  await expect(page.locator('#togglePassword')).toBeVisible();
  await expect(page.locator('#tour-save-basic-auth-btn')).toBeVisible();

  console.log('basic auth page loaded with expected fields');
});

test('generate password button fills password field', async ({ page }) => {
  await page.goto('/security/basic_auth');

  const passwordField = page.locator('#basic_auth_password');
  await passwordField.fill('');
  await page.locator('#generatePassword').click();

  const value = await passwordField.inputValue();
  expect(value.length).toBe(16);
  await expect(passwordField).toHaveAttribute('type', 'text');

  console.log('generate password populated a 16-char password and revealed it');
});

test('toggle password visibility button switches input type', async ({ page }) => {
  await page.goto('/security/basic_auth');

  const passwordField = page.locator('#basic_auth_password');
  await expect(passwordField).toHaveAttribute('type', 'password');

  await page.locator('#togglePassword').click();
  await expect(passwordField).toHaveAttribute('type', 'text');

  await page.locator('#togglePassword').click();
  await expect(passwordField).toHaveAttribute('type', 'password');

  console.log('toggle password visibility working');
});

test('save basic auth settings persists username (kept disabled)', async ({ page }) => {
  // NOTE: intentionally never sets basic_auth select to "yes" here -- enabling it would
  // require basic-auth credentials on every subsequent request and could lock the
  // Playwright session (and CI) out of OpenAdmin since the storageState has no such header.
  await page.goto('/security/basic_auth');

  await expect(page.locator('#basic_auth')).toHaveValue('no');

  const usernameField = page.locator('[name="basic_auth_username"]');
  const testUsername = `testbasicauth${Date.now() % 100000}`;
  await usernameField.fill(testUsername);
  await page.locator('#basic_auth_password').fill('TestPassword123');

  await page.locator('#tour-save-basic-auth-btn').click();
  await expect(page.getByText('Basic_auth settings for OpenAdmin edited successfully.')).toBeVisible();

  await page.reload();
  await expect(page.locator('[name="basic_auth_username"]')).toHaveValue(testUsername);
  await expect(page.locator('#basic_auth')).toHaveValue('no');

  console.log('basic auth settings saved and verified while staying disabled');
});
