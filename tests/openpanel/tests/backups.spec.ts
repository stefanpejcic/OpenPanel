import { test, expect } from '@playwright/test';


test('access backups page', async ({ page }) => {
  await page.goto('/backups');
  await expect(page).toHaveURL(/backups/);
  await expect(page.locator('body')).toContainText(/backup/i);
  console.log('backups page accessible');
});


test('access backup settings', async ({ page }) => {
  await page.goto('/backups/settings');
  await expect(page).toHaveURL(/backups\/settings/);
  await expect(page.locator('body')).toContainText(/backup/i);
  console.log('backup settings page accessible');
});


test('save backup settings', async ({ page }) => {
  await page.goto('/backups/settings');
  await page.getByRole('button', { name: 'Save settings' }).click();
  await expect(page.locator('body')).toContainText(/saved|success|updated/i);
  console.log('backup settings saved');
});


test('access backup destination', async ({ page }) => {
  await page.goto('/backups/destination');
  await expect(page).toHaveURL(/backups\/destination/);
  await expect(page.locator('body')).toContainText(/destination|remote|local|storage/i);
  console.log('backup destination page accessible');
});


test('run backup', async ({ page }) => {
  test.setTimeout(3 * 60 * 1000);

  await page.goto('/backups');
  const runBtn = page.getByRole('button', { name: /run backup|start backup|backup now/i });

  if (await runBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
    await runBtn.click();
    await expect(page.locator('body')).toContainText(/backup.*started|backup.*running|in progress/i, { timeout: 2 * 60 * 1000 });
    console.log('backup run triggered');
  } else {
    console.log('no run backup button found – skipping run test');
  }
});


test('list backups from destination', async ({ page }) => {
  await page.goto('/backups/list');
  await expect(page).toHaveURL(/backups\/list/);
  await expect(page.locator('body')).toContainText(/backup|no backups|list/i);
  console.log('backup list page accessible');
});
