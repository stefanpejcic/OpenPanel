import { test, expect } from '@playwright/test';

test('email address', async ({ page }) => {
  await page.goto('/account');

  const newEmail = `user_${Date.now()}@openpanel.org`;

  const emailInput = page.locator('#email');
  await emailInput.fill(newEmail);
  await page.getByRole('button', { name: 'Update' }).click();

  await expect(page.locator('body')).toContainText(/has been changed successfully/i);
  await expect(emailInput).toHaveValue(newEmail);

  console.log(`email changed to: ${newEmail}`);
});




// TODO: username change



test('password', async ({ page }) => {
  await page.goto('/account');

  const newPassword = 'novipassword';

  // CHANGE
  await page.locator('#password').fill(newPassword);
  await page.locator('#confirm_password').fill(newPassword);

  await page.locator('#save-button').click();

  await expect(page.locator('body')).toContainText(/has been changed successfully/i);

  await page.goto('/files');
  await expect(page).toHaveURL(/.*login/);

  // TEST LOGIN
  await page.getByRole('textbox', { name: 'Username' }).fill('testinguser');
  await page.locator('#password').fill(newPassword);
  await page.getByRole('button', { name: 'Sign In', exact: true }).click();

  await expect(page).toHaveURL(/.*dashboard/);

  console.log(`password changed and login successful`);

  // REVERT
  await page.goto('/account');

  const oldPassword = 'testingpassword';

  await page.locator('#password').fill(oldPassword);
  await page.locator('#confirm_password').fill(oldPassword);

  await page.locator('#save-button').click();

  await expect(page.locator('body')).toContainText(/has been changed successfully/i);
});


