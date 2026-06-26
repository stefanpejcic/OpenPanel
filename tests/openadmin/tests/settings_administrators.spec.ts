import { test, expect } from '@playwright/test';

test('administrators page loads with table', async ({ page }) => {
  await page.goto('/administrators');
  await expect(page).toHaveURL(/administrators/);
  await expect(page.locator('#exiting_users')).toBeVisible();

  console.log('administrators page loaded');
});

test('search filters the administrators table', async ({ page }) => {
  await page.goto('/administrators');

  const rows = page.locator('#exiting_users tbody tr');
  const count = await rows.count();
  test.skip(count === 0, 'No administrators to search');

  const firstUsername = (await rows.first().locator('td').nth(1).innerText()).trim();
  await page.locator('input[x-model="searchQuery"]').fill(firstUsername);
  await page.waitForTimeout(150);

  await expect(rows.filter({ hasText: firstUsername })).toBeVisible();
  console.log(`search filtered administrators table to "${firstUsername}"`);
});

test('create, inspect, and delete a new administrator (Enterprise only)', async ({ page }) => {
  await page.goto('/administrators');

  const createBtn = page.locator('#tour-create-admin-btn');
  const isEnterprise = await createBtn.isVisible().catch(() => false);
  test.skip(!isEnterprise, 'Create administrator is an Enterprise-only feature, not available on this license');

  await createBtn.click();

  const testUsername = `tstadmin${Date.now() % 100000}`;
  await page.locator('#username').fill(testUsername);
  await page.locator('#password').fill('TestPassword123');
  await page.getByRole('button', { name: 'Create' }).click();

  const row = page.locator('#exiting_users tbody tr').filter({ hasText: testUsername });
  await expect(row).toBeVisible({ timeout: 15_000 });
  await expect(row.getByText('Active')).toBeVisible();
  await expect(row.getByText('Admin', { exact: true })).toBeVisible();

  // delete the test administrator we just created
  await row.locator('button[data-dropdown-toggle]').click();
  const dropdown = page.locator(`#dropdown-${testUsername}`);
  await expect(dropdown).toBeVisible();
  await dropdown.getByRole('button', { name: 'Delete' }).click();

  await expect(page.locator('#exiting_users tbody tr').filter({ hasText: testUsername })).toHaveCount(0);
  console.log(`created and deleted test administrator "${testUsername}"`);
});

test('rename and change-password links navigate to dedicated pages', async ({ page }) => {
  await page.goto('/administrators');

  const rows = page.locator('#exiting_users tbody tr');
  const count = await rows.count();
  test.skip(count === 0, 'No administrators to inspect');

  const username = (await rows.first().locator('td').nth(1).innerText()).trim();
  await rows.first().locator('button[data-dropdown-toggle]').click();
  const dropdown = page.locator(`#dropdown-${username}`);
  await expect(dropdown.getByRole('link', { name: 'Rename' })).toHaveAttribute('href', `/administrators/rename/${username}`);
  await expect(dropdown.getByRole('link', { name: 'Change Password' })).toHaveAttribute('href', `/administrators/password/${username}`);

  console.log(`verified rename/password links for "${username}"`);
});
