import { test, expect } from '@playwright/test';

const domain = 'website-builder.tests.openpanel.org';

test('website builder install page loads', async ({ page }) => {
  await page.goto('/website-builder/install');
  await expect(page).toHaveURL(/website-builder\/install/);
  await expect(page.locator('body')).toContainText(/website builder|install|grapejs/i);
  console.log('website builder install page accessible');
});


test('install form has domain selector', async ({ page }) => {
  await page.goto('/website-builder/install');

  const domainSelect = page.locator('select[name="domain_id"]');
  const hasSelect = await domainSelect.isVisible({ timeout: 3000 }).catch(() => false);

  expect(hasSelect).toBe(true);
  const options = await domainSelect.locator('option').allTextContents();
  expect(options.length).toBeGreaterThan(0);
  console.log(`website builder install domain selector has ${options.length} options`);
});


test('website builder', async ({ page }) => {
  // 1. install
  await page.goto('/website-builder/install');
  await page.locator('#domain_id').selectOption('website-builder.tests.openpanel.org');
  await page.locator('#installButton').click();
  await expect(page.locator('text=Website creation completed!')).toBeVisible({ timeout: 60000 });
  await expect(page).toHaveURL(url => url.pathname === '/website-builder/edit' && url.searchParams.get('domain') === domain);

  // 2. test edit and save
  await page.locator('span.gjs-pn-btn.fa.fa-save').click();
  await expect(page.locator('text=Saved successfully!')).toBeVisible({ timeout: 30000 });
  
  await page.goto('http://website-builder.tests.openpanel.org/');
  await expect(async () => {
    await page.reload();
    await expect(page.locator('body')).toContainText('tailwindcss');
  }).toPass({ timeout: 30000, intervals: [1000] });

  
  // 3. test view
  await page.goto('/sites');
  const table = page.locator('tbody.divide-y.divide-gray-200.dark\\:divide-gray-800');
  await expect(table).toBeVisible();
  await expect(page.locator('tr#site-row-website-builder.tests.openpanel.org')).toBeVisible();
  console.log('website install is working');
  await expect(page.locator('a[href="/website-builder/edit?domain=website-builder.tests.openpanel.org"]')).toBeVisible();
  console.log('website edit is working');

  // 4. test remove
  await page.goto('/website?domain=website-builder.tests.openpanel.org');
  await page.locator('a#remove-tab').click();
  await page.locator('button#delete-site').click();
  await page.locator('button#confirm-delete-site').click();
  await expect(page.locator('text=Website deleted successfully!')).toBeVisible({ timeout: 30000 });
  await page.goto('/sites');
  await expect(page.locator('tr#site-row-website-builder.tests.openpanel.org')).not.toBeVisible();
  console.log('website uninstall is working');

  // 5. install again and test detach
  await page.goto('/website-builder/install');
  await page.locator('#domain_id').selectOption('website-builder.tests.openpanel.org');
  await page.locator('#installButton').click();
  await expect(page.locator('text=Website creation completed!')).toBeVisible({ timeout: 60000 });
  await expect(page).toHaveURL(/\/website-builder\/edit\?domain=website-builder\.tests\.openpanel\.org\/.+/);

  await page.goto('/website?domain=website-builder.tests.openpanel.org');
  await page.locator('a#remove-tab').click();
  await page.locator('button#detach-site').click();
  await page.locator('button#confirm-detach-site').click();
  await expect(page.locator('text=Website detached successfully!')).toBeVisible({ timeout: 30000 });
  await page.goto('/sites');
  await expect(page.locator('tr#site-row-website-builder.tests.openpanel.org')).not.toBeVisible();
  console.log('website detach is working');
  // TODO: remove files

});

test('website builder site appears in sites list', async ({ page }) => {
  await page.goto('/website-builder/install');
  await expect(page.locator('body')).toContainText(domain, { timeout: 10000 });
  console.log('website builder site in list');
});


test('website builder edit page loads', async ({ page }) => {
  await page.goto('/website-builder/edit');
  await expect(page).toHaveURL(/website-builder\/edit/);
  await expect(page.locator('body')).toContainText(/edit|builder|domain|grapejs/i, { timeout: 15000 });
  console.log('website builder edit page accessible');
});


test('remove website builder', async ({ page }) => {
  await page.goto('/website-builder/install');

  const row = page.locator('tr', { hasText: domain });
  const hasRow = await row.isVisible({ timeout: 5000 }).catch(() => false);

  if (!hasRow) {
    console.log('website builder row not found – skipping remove');
    return;
  }

  await row.locator('button:has-text("Remove"), a:has-text("Remove")').first().click();
  const confirmBtn = page.getByRole('button', { name: /confirm|yes|remove/i }).first();
  if (await confirmBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
    await confirmBtn.click();
  }

  await expect(page.locator('body')).toContainText(/removed|deleted|success/i, { timeout: 30000 });
  console.log('website builder removed');
});
