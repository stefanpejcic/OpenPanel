import { test as setup, expect } from '@playwright/test';
import path from 'path';

export const AUTH_FILE = path.join(__dirname, '../.auth/session.json');

const BASE_URL = process.env.BASE_URL;
const USERNAME = process.env.PANEL_USERNAME;
const PASSWORD = process.env.PANEL_PASSWORD;

setup('authenticate', async ({ page }) => {
  if (!BASE_URL || !USERNAME || !PASSWORD) {
    throw new Error('BASE_URL, USERNAME and PASSWORD must be set in .env');
  }

  await page.goto(`${BASE_URL}/login`);
  await page.getByRole('textbox', { name: 'Username' }).fill(USERNAME);
  await page.getByRole('textbox', { name: 'Password' }).fill(PASSWORD);
  await page.getByRole('button', { name: 'Sign In', exact: true }).click();
  await expect(page).toHaveURL(/.*dashboard/);
  await page.context().storageState({ path: AUTH_FILE });
});
