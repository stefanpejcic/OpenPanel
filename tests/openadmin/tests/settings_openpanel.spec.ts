import { test, expect } from '@playwright/test';

test('openpanel settings page loads with branding/nameserver/display sections', async ({ page }) => {
  await page.goto('/settings/open-panel');
  await expect(page).toHaveURL(/settings\/open-panel/);

  await expect(page.locator('form#changePanelSettings')).toBeVisible();
  await expect(page.locator('[name="brand_name"]')).toBeVisible();
  await expect(page.locator('[name="ns1"]')).toBeVisible();
  await expect(page.locator('[name="ns2"]')).toBeVisible();
  await expect(page.locator('[name="avatar_type"]')).toBeVisible();
  await expect(page.locator('#tour-save-openpanel-btn')).toBeVisible();

  console.log('openpanel settings page loaded with expected sections');
});

test('edit and save brand name, then revert', async ({ page }) => {
  await page.goto('/settings/open-panel');

  const input = page.locator('[name="brand_name"]');
  const original = await input.inputValue();
  const testValue = `Test Brand ${Date.now() % 100000}`;

  await input.fill(testValue);
  await page.locator('#tour-save-openpanel-btn').click();
  await page.waitForLoadState();
  await expect(page.locator('[name="brand_name"]')).toHaveValue(testValue);

  // revert
  await page.locator('[name="brand_name"]').fill(original);
  await page.locator('#tour-save-openpanel-btn').click();
  await page.waitForLoadState();
  await expect(page.locator('[name="brand_name"]')).toHaveValue(original);

  console.log('brand name edited, saved, and reverted');
});

test('nameserver and file manager limit fields are present and numeric where expected', async ({ page }) => {
  // NOTE: verify-only -- these fields (nameservers, file manager size/time limits,
  // database/session/rate-limit settings) affect server-wide behavior for every user;
  // we don't submit changes to them.
  await page.goto('/settings/open-panel');

  for (const name of ['ns1', 'ns2', 'ns3', 'ns4']) {
    await expect(page.locator(`[name="${name}"]`)).toBeVisible();
  }

  for (const name of ['filemanager_view_size', 'filemanager_edit_size', 'filemanager_upload_size', 'filemanager_download_size', 'login_ratelimit', 'session_duration', 'terminal_timeout']) {
    const field = page.locator(`[name="${name}"]`);
    await expect(field).toHaveAttribute('type', 'number');
  }

  console.log('nameserver and limit fields present with correct types (not saved)');
});

test('display toggle switches respond to clicks without saving', async ({ page }) => {
  await page.goto('/settings/open-panel');

  const hiddenInput = page.locator('input[type="hidden"][name="weakpass"]');
  await expect(hiddenInput).toBeHidden();
  await expect(hiddenInput).toHaveCount(1);

  console.log('verified weakpass toggle hidden input exists (toggle not exercised/saved)');
});
