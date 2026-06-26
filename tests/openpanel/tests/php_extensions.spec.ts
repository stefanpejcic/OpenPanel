import { test, expect } from '@playwright/test';

// Covers modules/phpd/php_extensions.py — /php/extensions, /php/php<version>/extensions,
// /php/php<version>/available-extensions, /php/php<version>/install-extensions(/status)

async function openExtensionsSelector(page: any) {
  await page.goto('/php/extensions');
  await expect(page).toHaveURL(/php\/extensions/);
}

test('extensions version selector page loads', async ({ page }) => {
  await openExtensionsSelector(page);
  await expect(page.getByRole('heading', { name: /PHP Extensions/i })).toBeVisible();

  const dropdown = page.locator('#php_version');
  await expect(dropdown).toBeVisible();
  const optionCount = await dropdown.locator('option').count();
  expect(optionCount).toBeGreaterThan(0);
});

test('select version navigates to per-version extensions page', async ({ page }) => {
  await openExtensionsSelector(page);

  const dropdown = page.locator('#php_version');
  const values = await dropdown.locator('option').evaluateAll(options =>
    options.map(opt => (opt as HTMLOptionElement).value).filter(v => v !== '')
  );

  if (values.length === 0) {
    test.skip(true, 'No PHP versions installed');
    return;
  }

  const version = values[0];
  await dropdown.selectOption(version);
  await page.click('#submit_version');

  await expect(page).toHaveURL(new RegExp(`/php/php${version}/extensions`));
  await expect(page.locator('table')).toBeVisible();
  console.log(`extensions page for PHP ${version} loaded`);
});

test('toggle an extension enable/disable', async ({ page }) => {
  await openExtensionsSelector(page);

  const dropdown = page.locator('#php_version');
  const values = await dropdown.locator('option').evaluateAll(options =>
    options.map(opt => (opt as HTMLOptionElement).value).filter(v => v !== '')
  );

  if (values.length === 0) {
    test.skip(true, 'No PHP versions installed');
    return;
  }

  const version = values[0];
  await page.goto(`/php/php${version}/extensions`);

  const rows = page.locator('tbody tr');
  const rowCount = await rows.count();
  if (rowCount === 0) {
    test.skip(true, 'No extensions installed for this version');
    return;
  }

  const firstRow = rows.first();
  const statusCell = firstRow.locator('td').nth(1);
  const wasEnabled = (await statusCell.textContent())?.includes('Enabled') ?? false;
  const button = firstRow.locator('button[type="submit"]');
  const expectedLabel = wasEnabled ? 'Enable' : 'Disable';

  await Promise.all([
    page.waitForLoadState('load'),
    button.click(),
  ]);

  await expect(page.locator('tbody tr').first().locator('td').nth(2).locator('button')).toHaveText(expectedLabel);
  console.log(`extension toggle for PHP ${version} is working`);
});

test('available extensions list loads in install modal', async ({ page }) => {
  await openExtensionsSelector(page);

  const dropdown = page.locator('#php_version');
  const values = await dropdown.locator('option').evaluateAll(options =>
    options.map(opt => (opt as HTMLOptionElement).value).filter(v => v !== '')
  );

  if (values.length === 0) {
    test.skip(true, 'No PHP versions installed');
    return;
  }

  const version = values[0];
  await page.goto(`/php/php${version}/extensions`);

  await Promise.all([
    page.waitForResponse(res => res.url().includes(`/php/php${version}/available-extensions`) && res.status() === 200),
    page.click('#open-install-modal'),
  ]);
  await expect(page.locator('#install-modal')).toBeVisible();

  const list = page.locator('#available-extensions-list');
  await expect(list).not.toContainText('Loading...');
  console.log(`available extensions list for PHP ${version} loaded`);

  await page.click('#cancel-install-modal');
  await expect(page.locator('#install-modal')).toBeHidden();
});
