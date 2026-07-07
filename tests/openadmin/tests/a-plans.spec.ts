import { test, expect } from '@playwright/test';

const PLAN_DATA = {
  name: 'probni',
  description: 'plan za test',
  disk_limit: '400',
  inodes_limit: '699999',
  cpu: '15',
  ram: '25',
  bandwidth: '28',
  domains_limit: '44',
  websites_limit: '55',
  db_limit: '88',
  email_limit: '99',
  max_email_quota: '77',
  max_hourly_email: '66',
  ftp_limit: '22',
};

const PLAN_DISPLAY = {
  name: 'probni',
  ram: '25 GB',
  cpu: '15 Core',
  disk_limit: '400 GB',
  inodes_limit: '699.999',   // UI adds thousands separator
  bandwidth: '28 mbits',
  domains_limit: '44',
  websites_limit: '55',
  db_limit: '88',
  email_limit: '99',
  max_email_quota: '77G',
  ftp_limit: '22',
};

async function fillPlanForm(page: any) {
  for (const [field, value] of Object.entries(PLAN_DATA)) {
    await page.locator(`input[name="${field}"]`).fill(value);
  }
}

async function verifyPlanRow(page: Page, planName: string) {
  const row = page.getByRole('row').filter({ hasText: planName });
  await expect(row).toHaveCount(1);

  for (const value of Object.values(PLAN_DISPLAY)) {
    await expect(row).toContainText(value);   // scoped to the row, not body
  }
  await expect(row).toContainText('mysql_only');
}

async function navigateToUserPackages(page: any) {
  await page.getByRole('button', { name: 'Hosting Plans' }).click();
  await page.getByRole('link', { name: 'User Packages' }).click();
}

test('create new hosting plan and verify all fields', async ({ page }) => {
  await page.goto(`/dashboard`);
  await expect(page).toHaveURL(/dashboard/);

  await navigateToUserPackages(page);
  await expect(page.getByText('developer plus')).toBeVisible();

  await page.getByRole('link', { name: 'Create New' }).click();
  await expect(page).toHaveURL(/\/plans\/new/);

  await fillPlanForm(page);
  await page.getByRole('combobox').selectOption('mysql_only');
  await page.getByRole('button', { name: 'Create Plan' }).click();

  await expect(page.getByText('plan probni created successfully')).toBeVisible();
  await verifyPlanRow(page, 'probni');

  console.log('Plan "probni" created successfully with all fields verified in table');
});

test('delete hosting plan', async ({ page }) => {
  await page.goto(`/dashboard`);
  await expect(page).toHaveURL(/dashboard/);

  await navigateToUserPackages(page);

  const row = page.getByRole('row').filter({ hasText: 'probni' }).first();
  await row.getByRole('button').click();                      // opens the Edit/Delete menu in the last cell
  await page.getByRole('button', { name: 'Delete' }).click();

  await expect(page.getByText(/plan deleted successfully/i)).toBeVisible();
  await expect(page.getByRole('row').filter({ hasText: 'probni' })).toHaveCount(0);
});

test('edit hosting plan and verify all fields', async ({ page }) => {
  await page.goto(`/dashboard`);
  await expect(page).toHaveURL(/dashboard/);

  await navigateToUserPackages(page);

  const row = page.getByRole('row').filter({ hasText: 'Developer Plus' });
  await row.getByRole('button').click();                       // open row actions menu
  await page.getByRole('link', { name: 'Edit' }).click();

  await fillPlanForm(page);
  await page.getByRole('combobox').selectOption('mysql_only');
  await page.getByRole('button', { name: 'Save changes' }).click();

  await expect(page.getByText(/successfully updated plan id/i)).toBeVisible();

  await navigateToUserPackages(page);
  await verifyPlanRow(page, 'probni');   // form renamed it to 'probni'
});



test('search hosting plans', async ({ page }) => {
  await page.goto(`/dashboard`);
  await expect(page).toHaveURL(/dashboard/);

  await navigateToUserPackages(page);
  await page.getByRole('searchbox', { name: 'Search by plan name...' }).fill('developer');
  
  const row = page.getByRole('row').filter({ hasText: 'Developer Plus' });
  await expect(row).toHaveCount(1);
  await expect(row).toHaveText(/developer plus/i);

  console.log('Plan search is functional');
});



test('check columns for hosting plans', async ({ page }) => {
  await page.goto(`/dashboard`);
  await expect(page).toHaveURL(/dashboard/);

  await navigateToUserPackages(page);

  await page.getByRole('button', { name: 'Show Columns' }).click();

  const rows = page.locator('ul[aria-labelledby="dropdownToggleButton"] li');

  const count = await rows.count();

  for (let i = 0; i < count; i++) {
    const row = rows.nth(i);

    const checkbox = row.locator('input[type="checkbox"]');

    const xModel = await checkbox.getAttribute('x-model');
    if (!xModel) continue;

    const columnKey = xModel.replace('columns.', '');
    const th = page.locator(`th[x-show="columns.${columnKey}"]`);

    const initialState = await checkbox.isChecked();

    await row.locator('label').click();
    await page.waitForTimeout(100); // needed for alpine.js x-show

    const expectedStateAfterToggle = !initialState;

    if (expectedStateAfterToggle) {
      await expect(th).toBeVisible();
    } else {
      await expect(th).toBeHidden();
    }

    await row.locator('label').click();
    await page.waitForTimeout(100);

    if (initialState) {
      await expect(th).toBeVisible();
    } else {
      await expect(th).toBeHidden();
    }
  }
  console.log('column toggle is working'); 
});
