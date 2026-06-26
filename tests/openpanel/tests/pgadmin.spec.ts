import { test, expect } from '@playwright/test';

// Covers modules/pgadmin.py — /postgresql/pgadmin (settings) and /pgadmin/ (auto-login redirect)
// Note: postgresql.spec.ts already does a light smoke check of /postgresql/pgadmin; this file
// owns the dedicated pgadmin module coverage (status toggle, settings edit, auto-login redirect).

async function openPgadminSettings(page: any) {
  await page.goto('/postgresql/pgadmin');
  await expect(page).toHaveURL(/postgresql\/pgadmin/);
}

test('pgadmin settings page loads', async ({ page }) => {
  await openPgadminSettings(page);
  await expect(page.getByRole('heading', { name: /pgAdmin/i })).toBeVisible();
  console.log('pgadmin settings page accessible');
});

test('enable and disable pgadmin service', async ({ page }) => {
  test.setTimeout(60_000);

  await openPgadminSettings(page);

  const enableBtn = page.locator('button', { hasText: 'Click to Enable' });
  const disableBtn = page.locator('button', { hasText: 'Click to Disable' });

  const isEnabled = await disableBtn.isVisible().catch(() => false);

  if (!isEnabled) {
    await enableBtn.click();
    await expect(page.locator('text=pgAdmin service is now enabled')).toBeVisible();
    await openPgadminSettings(page);
    await expect(disableBtn).toBeVisible();
    console.log('pgadmin enabled');
  } else {
    console.log('pgadmin already enabled, skipping enable step');
  }

  await disableBtn.click();
  await expect(page.locator('text=pgAdmin service is now disabled')).toBeVisible();
  await openPgadminSettings(page);
  await expect(enableBtn).toBeVisible();
  console.log('pgadmin disabled');

  // restore enabled state for the auto-login test below
  await enableBtn.click();
  await expect(page.locator('text=pgAdmin service is now enabled')).toBeVisible();
});

test('update pgadmin email setting', async ({ page }) => {
  await openPgadminSettings(page);

  const newEmail = `pgadmin-${Date.now()}@tests.openpanel.org`;

  const emailCard = page.locator('div.relative.rounded-lg.border').filter({ has: page.locator('input[name="email"]') });
  await emailCard.locator('span.cursor-pointer').click({ force: true });
  await emailCard.locator('input[name="email"]').fill(newEmail);
  await emailCard.getByRole('button', { name: 'Save' }).click();

  await openPgadminSettings(page);
  await expect(page.locator('body')).toContainText(newEmail);
  console.log('pgadmin email setting updated');
});

test('pgadmin auto-login redirect', async ({ page }) => {
  await openPgadminSettings(page);

  // ensure service is enabled so the redirect can produce a usable link
  const enableBtn = page.locator('button', { hasText: 'Click to Enable' });
  if (await enableBtn.isVisible().catch(() => false)) {
    await enableBtn.click();
    await expect(page.locator('text=pgAdmin service is now enabled')).toBeVisible();
  }

  await page.goto('/pgadmin/?output=json');
  const body = await page.locator('body').textContent();
  expect(body).toContain('pgadmin_url');
  console.log('pgadmin auto-login redirect is working');
});
