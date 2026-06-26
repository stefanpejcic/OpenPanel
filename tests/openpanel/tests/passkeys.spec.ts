import { test, expect } from '@playwright/test';

// Covers modules/account/passkeys.py — /account/passkeys (list/register/delete) and the
// /login/passkey/* WebAuthn endpoints. Uses Chromium's CDP virtual authenticator so the
// browser can satisfy navigator.credentials calls without real hardware.

async function addVirtualAuthenticator(page: any) {
  const client = await page.context().newCDPSession(page);
  await client.send('WebAuthn.enable');
  const { authenticatorId } = await client.send('WebAuthn.addVirtualAuthenticator', {
    options: {
      protocol: 'ctap2',
      transport: 'internal',
      hasResidentKey: true,
      hasUserVerification: true,
      isUserVerified: true,
      automaticPresenceSimulation: true,
    },
  });
  return { client, authenticatorId };
}

test('passkeys settings page loads', async ({ page }) => {
  await page.goto('/account/passkeys');
  await expect(page).toHaveURL(/account\/passkeys/);
  await expect(page.getByRole('heading', { name: /Passkeys/i })).toBeVisible();
  await expect(page.locator('#add-passkey-btn')).toBeVisible();
  console.log('passkeys settings page accessible');
});

test('register a new passkey', async ({ page }) => {
  test.setTimeout(60_000);

  await addVirtualAuthenticator(page);
  await page.goto('/account/passkeys');

  const passkeyName = `Test Passkey ${Date.now()}`;
  page.once('dialog', dialog => dialog.accept(passkeyName));

  await page.locator('#add-passkey-btn').click();
  await page.waitForURL(/account\/passkeys/, { timeout: 30000 });

  const row = page.locator('tr', { hasText: passkeyName });
  await expect(row).toBeVisible();
  console.log('passkey registered:', passkeyName);
});

test('remove a passkey', async ({ page }) => {
  test.setTimeout(60_000);

  await addVirtualAuthenticator(page);
  await page.goto('/account/passkeys');

  const passkeyName = `Removable Passkey ${Date.now()}`;
  page.once('dialog', dialog => dialog.accept(passkeyName));
  await page.locator('#add-passkey-btn').click();
  await page.waitForURL(/account\/passkeys/, { timeout: 30000 });

  const row = page.locator('tr', { hasText: passkeyName });
  await expect(row).toBeVisible();

  page.once('dialog', dialog => dialog.accept());
  await row.getByRole('button', { name: 'Remove' }).click();

  await expect(page.locator('tr', { hasText: passkeyName })).toHaveCount(0);
  console.log('passkey removed:', passkeyName);
});
