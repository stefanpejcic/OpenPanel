import { test, expect } from '@playwright/test';

// NOTE: this spec intentionally never submits the enable/disable forms.
// Enabling 2FA on the admin account used for the Playwright session would lock the
// suite out on the next run (the storageState has no TOTP secret to compute a code),
// and disabling it requires a password we don't want to assume. We only verify the
// UI renders correctly for whichever state the account is currently in.

test('2fa page loads and shows the correct branch for current state', async ({ page }) => {
  await page.goto('/security/2fa');
  await expect(page).toHaveURL(/security\/2fa/);

  const container = page.locator('#tour-2fa');
  await expect(container).toBeVisible();

  const enabledNotice = page.getByText('Two-factor authentication is enabled.');
  const isEnabled = await enabledNotice.isVisible().catch(() => false);

  if (isEnabled) {
    await expect(page.locator('#password')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Disable two-factor authentication' })).toBeVisible();
    console.log('2fa page: account currently has 2FA enabled, disable form present');
  } else {
    await expect(page.locator('#code')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Enable two-factor authentication' })).toBeVisible();
    // QR code or manual secret should be present to set up an authenticator app
    const qrVisible = await page.locator('img[alt="2FA QR code"]').isVisible().catch(() => false);
    const secretVisible = await page.locator('code.font-mono').isVisible().catch(() => false);
    expect(qrVisible || secretVisible).toBeTruthy();
    console.log('2fa page: account currently has 2FA disabled, setup form present');
  }
});

test('2fa code input only accepts 6 digits', async ({ page }) => {
  await page.goto('/security/2fa');

  const codeInput = page.locator('#code');
  const isVisible = await codeInput.isVisible().catch(() => false);
  test.skip(!isVisible, '2FA already enabled on this account, no enable form to test');

  await expect(codeInput).toHaveAttribute('maxlength', '6');
  await expect(codeInput).toHaveAttribute('pattern', '[0-9]*');

  console.log('2fa code input has correct constraints');
});
