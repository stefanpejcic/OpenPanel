import { test, expect } from '@playwright/test';
import fs from 'fs';
import os from 'os';
import path from 'path';

async function navigateToPostgreSQLPage(page: any) {
  await page.goto('/postgresql');
  await expect(page).toHaveURL(/postgresql/);
}

async function expectDatabaseInTable(page: any, dbName: string) {
  const row = page.locator('tr', { hasText: dbName });
  await expect(row).toBeVisible();
}

async function expectDatabaseNotInTable(page: any, dbName: string) {
  await expect(page.locator('tr', { hasText: dbName })).toHaveCount(0);
}


// ACCESS
test('list databases', async ({ page }) => {
  await navigateToPostgreSQLPage(page);
  await expect(page.locator('body')).toContainText(/create your first database|no databases/i, { timeout: 20000 });
  console.log('postgresql initialized');
});


test('create database', async ({ page }) => {
  await navigateToPostgreSQLPage(page);
  await page.getByRole('link', { name: 'Create your first database' }).click();
  await page.getByRole('textbox', { name: 'Database Name' }).fill('stefan_psql');
  await page.getByRole('button', { name: 'Create Database' }).click();
  await expect(page.locator('body')).toContainText(/successfully created/i);
  await expectDatabaseInTable(page, 'stefan_psql');
  console.log('postgresql database created');
});


test('list users', async ({ page }) => {
  await page.goto('/postgresql/users');
  await expect(page).toHaveURL(/postgresql\/users/);
  await expect(page.locator('body')).toContainText(/no users yet|users/i);
  console.log('postgresql users page accessible');
});


test('create user', async ({ page }) => {
  await page.goto('/postgresql/users');
  await page.getByRole('link', { name: 'Create your first user' }).click();
  await expect(page).toHaveURL(/postgresql\/user/);
  await page.getByRole('textbox', { name: 'Username*' }).fill('stefan_psql_user');
  await page.getByRole('textbox', { name: 'Password*' }).fill('stefan94');
  await page.getByRole('button', { name: 'Create User' }).click();
  await expect(page.locator('body')).toContainText(/successfully created.*stefan_psql_user/i);
  await page.getByRole('link', { name: 'Back to Users' }).click();
  await expect(page).toHaveURL(/postgresql\/users/);
  await expect(page.locator('body')).toContainText(/stefan_psql_user/i);
  console.log('postgresql user created');
});


test('change password', async ({ page }) => {
  await page.goto('/postgresql/users');
  await page.getByRole('link', { name: ' Change Password' }).click();
  await expect(page).toHaveURL(/postgresql\/password/);
  await page.locator('#generatePassword').click();
  await page.getByRole('button', { name: 'Change Password' }).click();
  await expect(page).toHaveURL(/postgresql\/users/);
  await expect(page.locator('body')).toContainText(/successfully changed password/i);
  console.log('postgresql change password working');
});


test('assign user to database', async ({ page }) => {
  await page.goto('/postgresql/users');
  await page.getByRole('link', { name: 'Assign User to Database' }).click();
  await expect(page).toHaveURL(/postgresql\/assign/);
  await page.waitForResponse(resp => resp.url().includes('/postgresql/info') && resp.status() === 200);
  await page.locator('select[name="db_user"]').selectOption('stefan_psql_user');
  await page.locator('select[name="database_name"]').selectOption('stefan_psql');
  await page.getByRole('button', { name: 'Assign' }).click();
  await expect(page.locator('body')).toContainText(/successfully assigned|privileges granted/i);
  console.log('postgresql user assigned to database');
});


test('revoke user from database', async ({ page }) => {
  await page.goto('/postgresql/users');
  await page.getByRole('link', { name: 'Remove User from DB' }).click();
  await expect(page).toHaveURL(/postgresql\/remove/);
  await page.waitForResponse(resp => resp.url().includes('/postgresql/info') && resp.status() === 200);
  await page.locator('select[name="db_user"]').selectOption('stefan_psql_user');
  await page.locator('select[name="database_name"]').selectOption('stefan_psql');
  await page.getByRole('button', { name: 'Remove User from Database' }).click();
  await expect(page.locator('body')).toContainText(/successfully revoked|removed/i);
  console.log('postgresql user revoked from database');
});


test('database wizard', async ({ page }) => {
  await page.goto('/postgresql/wizard');
  await expect(page).toHaveURL(/postgresql\/wizard/);
  await page.getByRole('textbox', { name: 'Database Name' }).fill('psql_proba');
  await page.getByRole('textbox', { name: 'Database User' }).fill('psql_novi_user');
  await page.getByRole('textbox', { name: 'Password' }).fill('stefan456g7dsd');
  await page.getByRole('button', { name: 'Create DB, User, and Grant' }).click();
  await expect(page.getByText('Process completed!')).toBeVisible();
  await page.getByRole('link', { name: 'Back to Databases' }).click();
  await expect(page).toHaveURL(/postgresql/);
  const row = page.locator('#databases-table tr', { hasText: 'psql_proba' });
  await expect(row).toContainText(/psql_proba/i);
  console.log('postgresql database wizard is working');
});


test('processlist', async ({ page }) => {
  await page.goto('/postgresql/processlist');
  await expect(page).toHaveURL(/postgresql\/processlist/);
  await expect(page.locator('body')).toContainText(/state|query|pid/i);
  console.log('postgresql processlist is working');
});


test('configuration editor', async ({ page }) => {
  await page.goto('/postgresql/configuration');
  await expect(page).toHaveURL(/postgresql\/configuration/);
  await expect(page.locator('body')).toContainText(/max_connections|shared_buffers/i);
  await page.locator('#max_connections').fill('110');
  await page.getByRole('button', { name: 'Save Changes' }).click();
  await expect(page.locator('body')).toContainText(/configuration updated|saved/i);
  await page.waitForLoadState('networkidle');
  await expect(page.locator('#max_connections')).toHaveValue('110');
  console.log('postgresql configuration is saved');
});


test('remote access', async ({ page }) => {
  await page.goto('/postgresql/remote-postgresql');

  const statusText = page.locator('dd p.text-lg.font-semibold');
  await expect(statusText).toHaveText('Disabled');

  // ON
  const enableBtn = page.locator('dd button', { hasText: 'Click to Enable' });
  await enableBtn.click();
  await expect(page.locator('text=Remote PostgreSQL access is now enabled')).toBeVisible();
  await expect(statusText).toHaveText('Enabled');
  console.log('postgresql remote access enabled');

  // OFF
  await page.goto('/postgresql/remote-postgresql');
  const disableBtn = page.locator('dd button', { hasText: 'Click to Disable' });
  await disableBtn.click();
  await expect(page.locator('text=Remote PostgreSQL access is now disabled')).toBeVisible();
  await expect(statusText).toHaveText('Disabled');
  console.log('postgresql remote access disabled');
});


// IMPORT
test('import', async ({ page }) => {
  const tempFilePath = path.join(os.tmpdir(), 'test-psql-import.sql');
  const sqlContent = `
DROP TABLE IF EXISTS users;
CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(50));
INSERT INTO users VALUES (1, 'John');
`;
  fs.writeFileSync(tempFilePath, sqlContent);

  await page.goto('/postgresql/import/stefan_psql');
  await expect(page).toHaveURL(/postgresql\/import\/stefan_psql/);
  await page.waitForResponse(resp => resp.url().includes('/postgresql/info') && resp.status() === 200);
  await page.locator('select[name="database_name"]').selectOption('stefan_psql');
  await page.locator('input[name="db_file"]').setInputFiles(tempFilePath);
  await page.getByRole('button', { name: 'Upload & Import' }).click();
  await expect(page.locator('body')).toContainText(/successfully imported|import.*success/i);
  console.log('postgresql import working');
});


test('pgadmin settings', async ({ page }) => {
  await page.goto('/postgresql/pgadmin');
  await expect(page).toHaveURL(/postgresql\/pgadmin/);
  await expect(page.locator('body')).toContainText(/pgadmin|postgresql/i);
  console.log('pgadmin settings page accessible');
});


test('delete user', async ({ page }) => {
  await page.goto('/postgresql/users');
  const deleteButtons = page.locator('button.btn-danger');
  const count = await deleteButtons.count();
  expect(count).toBeGreaterThan(0);
  await deleteButtons.first().click();
  const confirmButton = page.locator('button.btn-dark');
  await expect(confirmButton.first()).toBeVisible();
  await confirmButton.first().click();
  await expect(page.locator('body')).toContainText(/successfully deleted/i);
  console.log('postgresql delete user is working');
});


test('delete database', async ({ page }) => {
  await navigateToPostgreSQLPage(page);
  const dbName = 'stefan_psql';
  const row = page.locator('tr', { hasText: dbName });
  const deleteButton = row.locator('button.btn-danger');
  await expect(deleteButton).toBeVisible();
  await deleteButton.click();
  const confirmButton = page.locator('button.btn-dark');
  await expect(confirmButton).toBeVisible();
  await confirmButton.click();
  await expect(page.locator('body')).toContainText(/successfully deleted/i);
  await expectDatabaseNotInTable(page, dbName);
  console.log('postgresql database deleted');
});
