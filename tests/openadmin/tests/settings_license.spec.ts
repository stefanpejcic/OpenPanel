import { test, expect } from '@playwright/test';

// NOTE: verify-only on submitting a key -- saving an invalid/test license key would
// require an OpenAdmin restart to take effect and could downgrade Enterprise features
// on a real license; we don't have a real key to round-trip safely.

test('license page loads with key field and actions', async ({ page }) => {
  await page.goto('/license');
  await expect(page).toHaveURL(/license);

  await expect(page.locator('#license-key')).toBeVisible();
  await expect(page.locator('#save_license_btn')).toBeVisible();
  await expect(page.getByRole('link', { name: 'my.openpanel.com' })).toHaveAttribute('href', /my\.openpanel\.com/);

  console.log('license page loaded with key field and actions');
});

test('verify/downgrade links are gated on having a key set', async ({ page }) => {
  await page.goto('/license');

  const keyValue = await page.locator('#license-key').inputValue();
  const verifyLink = page.locator('a.btn-verify');
  const downgradeLink = page.locator('a.btn-downgrade');

  if (keyValue) {
    await expect(verifyLink).toBeVisible();
    await expect(downgradeLink).toBeVisible();
    console.log('license key present, verify/downgrade links visible');
  } else {
    await expect(verifyLink).toBeHidden();
    await expect(downgradeLink).toBeHidden();
    console.log('no license key set, verify/downgrade links correctly hidden');
  }
});

test('generate support report link is present', async ({ page }) => {
  await page.goto('/license');

  await expect(page.locator('#tour-generate-report-btn')).toHaveAttribute('href', '/support/report');
  console.log('generate support report link present');
});
