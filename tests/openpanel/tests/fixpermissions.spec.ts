import { test, expect } from '@playwright/test';

async function deleteSelected(page: any, skipTrash = false) {
  await page.locator('#deleteButton').click();
  if (skipTrash) {
    await page.getByRole('checkbox', { name: 'Skip the trash and' }).check();
  }
  await page.getByRole('button', { name: 'Delete', exact: true }).click();
}

async function selectItem(page: any, name: string, multiSelect = false) {
  await page.locator('#filemanager_table div').filter({ hasText: name }).click(
    multiSelect ? { modifiers: ['ControlOrMeta'] } : undefined
  );
}

test('fix permissions', async ({ page }) => {

  // create test files
  await page.goto(`/file-manager/edit-file/test.txt?editor=text&new=true`);
  await page.locator('#editor-text').fill(`nista`);
  await page.getByRole('button', { name: 'Save' }).click();

  await page.goto(`/files`);
  await page.getByRole('button', { name: ' New Folder' }).click();
  await page.locator('#foldername').fill('testdir');
  await page.getByRole('button', { name: 'Create' }).click();

  // change their permissons
  await page.locator('#filemanager_table div').filter({ hasText: 'testdir' }).click();
  await page.getByRole('button', { name: ' Permissions' }).click();
  await page.getByPlaceholder('775').fill('200');
  await page.getByRole('button', { name: 'Confirm' }).click();
  await page.locator('tr[data-file="test.txt"]').click();
  await page.getByRole('button', { name: ' Permissions' }).click();
  await page.getByPlaceholder('775').fill('200');
  await page.getByRole('button', { name: 'Confirm' }).click();

  // fix permissions
  await page.goto(`/fix-permissions`)
  await page.getByRole('button', { name: 'Fix Permissions' }).click();
  const message = page.locator('#scan-complete-message');
  await expect(message).toBeVisible();
  await expect(message).toHaveText('Permissions are fixed!');

  // validate
  await page.goto(`/files`);
  await page.locator('#filemanager_table div').filter({ hasText: 'testdir' }).click();
  await page.getByRole('button', { name: ' Permissions' }).click();
  await expect(page.locator('#c-oct')).toHaveValue('775');

  await page.goto(`/files`);
  await selectItem(page, 'test.txt');
  await page.getByRole('button', { name: ' Permissions' }).click();
  await expect(page.locator('#c-oct')).toHaveValue('644');

  // cleanup
  await page.goto(`/files`);
  await selectItem(page, 'testdir');
  await selectItem(page, 'test.txt', true);
  await deleteSelected(page, true);
});
