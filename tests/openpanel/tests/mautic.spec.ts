import { test, expect } from '@playwright/test';

// Covers modules/mautic.py — /mautic, /mautic/install, /website/mautic_info/<domain>,
// /mautic/backup/run/<domain>, /mautic/remove

const domain = 'mautic.tests.openpanel.org';

test.setTimeout(10 * 60 * 1000);

test.beforeAll(async ({ browser }) => {
  const page = await browser.newPage();
  await page.goto('/domains/new');
  await page.getByRole('textbox', { name: 'Domain*' }).fill(domain);
  await page.getByRole('button', { name: 'Add Domain' }).click();
  await expect(page.getByText(`Domain name ${domain} added successfully`)).toBeVisible();
  await page.close();
});

test('mautic install page loads', async ({ page }) => {
  await page.goto('/mautic/install');
  await expect(page).toHaveURL(/mautic\/install/);
  await expect(page.getByRole('heading', { name: /Install Mautic/i })).toBeVisible();
  await expect(page.locator('#domain_id')).toBeVisible();
  console.log('mautic install page accessible');
});

test('install mautic', async ({ page }) => {
  await page.goto('/mautic/install');
  await expect(page).toHaveURL(/mautic\/install/);

  await page.fill('#website_name', 'My Mautic Site');
  await page.locator('#domain_id').selectOption({ label: domain });

  // wait for username/password autofill from page JS
  await expect(page.locator('#admin_username')).not.toHaveValue('');
  await expect(page.locator('#admin_password')).not.toHaveValue('');
  await expect(page.locator('#admin_email')).toHaveValue(new RegExp(`@${domain}$`));

  // version dropdown is populated async from GitHub releases; fall back to any option
  await page.waitForFunction(() => {
    const select = document.querySelector('#mautic_version') as HTMLSelectElement | null;
    return !!select && select.options.length > 0;
  }, { timeout: 30000 });

  await page.locator('#installButton').click();

  const statusMessage = page.locator('#statusMessage');
  await expect(statusMessage).toBeVisible({ timeout: 30000 });
  await expect(page.locator('body')).toContainText(/Mautic installation completed/i, { timeout: 8 * 60 * 1000 });
  console.log('mautic install completed');

  await page.waitForURL(/\/sites/, { timeout: 30000 });
  await expect(page.locator('body')).toContainText(domain);
  console.log('mautic site appears in sites list');
});

test('mautic manager dashboard data', async ({ page }) => {
  await page.goto(`/website?domain=${domain}`);

  await expect(page.locator('#mautic-version')).not.toHaveText('', { timeout: 30000 });
  await expect(page.locator('#php-version')).toContainText(/\d+\.\d+/, { timeout: 30000 });
  await expect(page.locator('#mysql-version')).toContainText(/\d+\.\d+/, { timeout: 30000 });
  await expect(page.locator('#database-name')).not.toHaveText('');
  await expect(page.locator('#database-user')).not.toHaveText('');

  console.log('mautic manager dashboard data is populated');
});

test('mautic live preview', async ({ page }) => {
  await page.goto(`/website?domain=${domain}`);

  const popupPromise = page.waitForEvent('popup');
  await page.locator('button[onclick="sendDataToPreview(event)"]').click();

  const previewPage = await popupPromise;
  await previewPage.waitForLoadState();
  expect(previewPage.url()).toBeTruthy();
  console.log('mautic live preview is working');
});

test('mautic backup', async ({ page }) => {
  test.setTimeout(3 * 60 * 1000);
  await page.goto(`/mautic/backup/run/${domain}`);
  await expect(page.locator('body')).toContainText(/backup.*complete|successfully.*backup|done/i, { timeout: 2 * 60 * 1000 });
  console.log('mautic backup completed');
});

test('mautic uninstall', async ({ page }) => {
  await page.goto(`/website?domain=${domain}`);
  await page.locator('a#remove-tab').click();

  await page.getByRole('button', { name: 'Uninstall Mautic' }).click();
  await page.getByRole('button', { name: 'Confirm Uninstall' }).click();

  await page.waitForURL(/\/sites/, { timeout: 60000 });
  await expect(page.locator(`tr#site-row-${domain}`)).toHaveCount(0);
  console.log('mautic uninstall is working');
});
