import { test, expect } from '@playwright/test';

// NOTE: actually registering a passkey requires a WebAuthn virtual authenticator
// (Chrome DevTools Protocol), which standard Playwright `page` actions can't drive.
// We verify the surrounding UI instead: listing, rename, remove (without confirming),
// and the add-passkey section gating logic.

test('passkeys page loads', async ({ page }) => {
  await page.goto('/security/passkeys');
  await expect(page).toHaveURL(/security\/passkeys/);
  await expect(page.getByText('Your passkeys')).toBeVisible();
  await expect(page.getByText('Add a passkey')).toBeVisible();

  console.log('passkeys page loaded');
});

test('shows either domain warning or add-passkey form depending on access method', async ({ page }) => {
  await page.goto('/security/passkeys');

  const domainWarning = page.locator('#tour-passkeys-needs-domain');
  const addBtn = page.locator('#add-passkey-btn');

  const warningVisible = await domainWarning.isVisible().catch(() => false);
  if (warningVisible) {
    await expect(domainWarning).toContainText('Passkeys require OpenAdmin to be accessed over a domain name');
    console.log('passkeys: accessed via IP, domain warning shown as expected');
  } else {
    await expect(addBtn).toBeVisible();
    await expect(page.locator('#passkey-name')).toBeVisible();
    console.log('passkeys: webauthn-capable, add-passkey form shown');
  }
});

test('existing passkeys list shows rename and remove forms', async ({ page }) => {
  await page.goto('/security/passkeys');

  const noPasskeysNotice = page.getByText('No passkeys registered yet.');
  const hasNone = await noPasskeysNotice.isVisible().catch(() => false);
  test.skip(hasNone, 'No passkeys registered on this account to inspect');

  const firstRow = page.locator('form[action*="passkeys/rename"]').first().locator('xpath=ancestor::div[contains(@class,"flex items-center justify-between")][1]');
  await expect(firstRow.getByRole('button', { name: 'Rename' })).toBeVisible();
  await expect(firstRow.getByRole('button', { name: 'Remove' })).toBeVisible();

  console.log('passkey row has rename and remove actions');
});

test('remove passkey form has a confirm dialog and is not submitted', async ({ page }) => {
  await page.goto('/security/passkeys');

  const removeForm = page.locator('form[action*="passkeys/delete"]').first();
  const exists = await removeForm.count();
  test.skip(exists === 0, 'No passkeys registered on this account to inspect');

  await expect(removeForm).toHaveAttribute('onsubmit', /confirm\(/);

  // dismiss the confirm dialog so the passkey is NOT actually removed
  page.once('dialog', dialog => dialog.dismiss());
  await removeForm.getByRole('button', { name: 'Remove' }).click();

  await expect(page).toHaveURL(/security\/passkeys/);
  console.log('remove passkey confirm dialog dismissed, passkey not removed');
});
