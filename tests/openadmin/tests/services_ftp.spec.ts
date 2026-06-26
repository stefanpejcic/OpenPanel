import { test, expect } from '@playwright/test';

test('ftp accounts page renders branch matching current ftp server status', async ({ page }) => {
  await page.goto('/services/ftp');
  await expect(page).toHaveURL(/services\/ftp/);

  await expect(page.locator('#tour-ftp-status')).toBeVisible();

  const notInstalled = page.getByText('FTP server is not installed');
  const stopped = page.getByText('FTP server is not running');
  const crashed = page.getByText('FTP server is crashed');

  if (await notInstalled.isVisible().catch(() => false)) {
    await expect(page.getByText('docker --context default compose up -d openadmin_ftp')).toBeVisible();
    console.log('ftp: not installed branch rendered');
  } else if (await stopped.isVisible().catch(() => false)) {
    await expect(page.locator('#start_ftpserver')).toBeVisible();
    console.log('ftp: stopped branch rendered, start link present (not clicked)');
  } else if (await crashed.isVisible().catch(() => false)) {
    console.log('ftp: crashed/unknown branch rendered');
  } else {
    // running branch
    await expect(page.locator('table')).toBeVisible();
    console.log('ftp: running branch rendered with accounts table');
  }
});

test('configuration tab navigates to ftp settings page', async ({ page }) => {
  await page.goto('/services/ftp');

  const configTab = page.locator('#tour-ftp-config-tab');
  const visible = await configTab.isVisible().catch(() => false);
  test.skip(!visible, 'Configuration tab not rendered for current ftp status branch');

  await configTab.click();
  await expect(page).toHaveURL(/services\/ftp\/settings/);
  await expect(page.locator('#config_content')).toBeVisible();

  console.log('navigated to ftp settings configuration page');
});

test('edit and save ftp configuration, then revert', async ({ page }) => {
  await page.goto('/services/ftp/settings');
  await expect(page).toHaveURL(/services\/ftp\/settings/);

  const textarea = page.locator('#config_content');
  const original = await textarea.inputValue();
  const comment = `\n# test-comment-${Date.now()}`;

  await textarea.fill(original + comment);
  await page.getByRole('button', { name: 'Save Configuration' }).click();

  await page.waitForLoadState();
  await expect(page.locator('#config_content')).toHaveValue(new RegExp(comment.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')));

  // revert
  await page.locator('#config_content').fill(original);
  await page.getByRole('button', { name: 'Save Configuration' }).click();
  await page.waitForLoadState();
  await expect(page.locator('#config_content')).toHaveValue(original);

  console.log('ftp configuration edited, saved, and reverted');
});
