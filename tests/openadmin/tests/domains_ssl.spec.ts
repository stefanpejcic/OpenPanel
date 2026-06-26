import { test, expect } from '@playwright/test';

async function firstDomainSslUrl(page) {
  await page.goto('/domains');
  const link = page.locator('a[href^="/domains/ssl/"]').first();
  const count = await link.count();
  if (count === 0) return null;
  return link.getAttribute('href');
}

test('ssl management page loads for a domain', async ({ page }) => {
  const href = await firstDomainSslUrl(page);
  test.skip(!href, 'No domains available to inspect SSL for');

  await page.goto(href!);
  await expect(page.getByText('Manage SSL for')).toBeVisible();
  await expect(page.getByText('Status')).toBeVisible();
  await expect(page.getByRole('link', { name: 'Back to Domains' })).toHaveAttribute('href', '/domains');

  console.log(`ssl management page loaded for ${href}`);
});

test('ssl status shows AutoSSL, Custom SSL, or Unknown branch', async ({ page }) => {
  // NOTE: verify-only -- "Switch back to AutoSSL" / "Try enabling AutoSSL" trigger a real
  // certificate issuance/renewal flow; we don't want to fire that from every test run.
  const href = await firstDomainSslUrl(page);
  test.skip(!href, 'No domains available to inspect SSL for');

  await page.goto(href!);

  const autoSsl = page.getByText('Auto SSL');
  const customSsl = page.getByText('Custom SSL');
  const unknown = page.getByText('Unknown');

  if (await autoSsl.isVisible().catch(() => false)) {
    console.log('ssl status: AutoSSL currently active');
  } else if (await customSsl.isVisible().catch(() => false)) {
    await expect(page.getByRole('button', { name: 'Switch back to AutoSSL' })).toBeVisible();
    console.log('ssl status: Custom SSL currently active, switch-back button present (not clicked)');
  } else {
    await expect(unknown).toBeVisible();
    await expect(page.getByRole('button', { name: 'Try enabling AutoSSL' })).toBeVisible();
    console.log('ssl status: Unknown, enable-autossl button present (not clicked)');
  }
});

test('custom certificate upload form is present', async ({ page }) => {
  const href = await firstDomainSslUrl(page);
  test.skip(!href, 'No domains available to inspect SSL for');

  await page.goto(href!);

  const publicPath = page.locator('input[name="public_path"]');
  const hasCustomForm = (await publicPath.count()) > 0;
  test.skip(!hasCustomForm, 'Custom SSL upload form not rendered on this environment/state');

  await expect(publicPath).toBeVisible();
  await expect(page.locator('input[name="private_path"]')).toBeVisible();

  console.log('custom certificate upload form present (not submitted)');
});
