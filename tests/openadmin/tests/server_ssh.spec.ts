import { test, expect } from '@playwright/test';

// NOTE: we verify the SSH config UI but never submit the basic/advanced/keys forms --
// changing the SSH port, root login, password/pubkey auth, or the raw sshd_config could
// lock out real SSH access to the host with no UI-based way to revert.

test('ssh page loads with status indicator and basic tab fields', async ({ page }) => {
  await page.goto('/server/ssh');
  await expect(page).toHaveURL(/server\/ssh/);

  await expect(page.getByText('SSH service status:')).toBeVisible();
  await expect(page.locator('#port')).toBeVisible();
  await expect(page.locator('#permit_root_login')).toBeVisible();
  await expect(page.locator('#password_auth')).toBeVisible();
  await expect(page.locator('#pubkey_auth')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Save', exact: true })).toBeVisible();

  console.log('ssh basic tab loaded with expected fields');
});

test('switching to advanced tab lazy-loads full sshd config', async ({ page }) => {
  await page.goto('/server/ssh');

  await page.getByRole('button', { name: 'Advanced' }).click();

  const textarea = page.locator('#full_config');
  await expect(textarea).toBeVisible();
  await expect(textarea).not.toHaveValue('', { timeout: 10_000 });

  await expect(page.getByRole('button', { name: 'Save Configuration' })).toBeVisible();
});

test('keys tab is only shown when pubkey auth is enabled', async ({ page }) => {
  await page.goto('/server/ssh');

  const pubkeyAuth = await page.locator('#pubkey_auth').inputValue();
  const keysTab = page.getByRole('button', { name: 'Authorized Keys' });

  if (pubkeyAuth === 'yes') {
    await expect(keysTab).toBeVisible();
    await keysTab.click();
    await expect(page).toHaveURL(/#keys/);
    await expect(page.getByText('Add New SSH Key')).toBeVisible();
    await expect(page.locator('textarea[name="new_key"]')).toBeVisible();
    console.log('pubkey auth enabled, keys tab shows authorized keys management');
  } else {
    await expect(keysTab).toHaveCount(0);
    console.log('pubkey auth disabled, keys tab correctly hidden');
  }
});

test('tab state persists via URL hash on reload', async ({ page }) => {
  await page.goto('/server/ssh#advanced');
  await expect(page.locator('#full_config')).toBeVisible();

  console.log('advanced tab restored from URL hash');
});
