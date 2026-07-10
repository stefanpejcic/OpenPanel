import { test, expect, type Page } from '@playwright/test';
import { generate } from 'otplib';

const USERNAME = process.env.PANEL_USERNAME;
const PASSWORD = process.env.PANEL_PASSWORD;

// ENABLE
test('enable 2FA', async ({ page }) => {
  await page.goto(`/account/2fa`);
  await expect(page.locator('#twofa_code')).not.toBeVisible();
  await expect(page.getByText('2FA is currently disabled')).toBeVisible();
  await page.click('button:has-text("Click to enable 2FA")');
  await page.click('#showLink');
  const secretEl = page.locator('#initiallyhiddencode');
  await expect(secretEl).toBeVisible();
  const totpSecret = (await secretEl.textContent())!.trim();
  expect(totpSecret.length).toBeGreaterThan(10);
  console.log('Captured TOTP secret:', totpSecret);
  await page.goto(`/account/2fa`);
  const step2Tab = page.locator('#dashboard-styled-tab');
  if (await step2Tab.isVisible()) {
    await step2Tab.click();
  }
  // fill the 6-digit code BEFORE confirming
  const otpInput = page.getByRole('textbox', {
    name: 'Enter the 6-digit code from your app to confirm',
  });
  await expect(otpInput).toBeVisible();
  await otpInput.fill(await generate({ secret: totpSecret }));
  await page.getByRole('button', { name: 'Confirm' }).click();
  await expect(page.getByText('2FA is currently enabled')).toBeVisible();
  // VERIFY dashboard shows 2FA enabled
  await page.goto(`/dashboard`);
  const dashboardTwofa = page.locator('#dashboard_twofa_content');
  await expect(dashboardTwofa).toBeVisible();
  await expect(dashboardTwofa.locator('b')).toHaveText('enabled');
  // TEST with incorrect
  await page.goto(`/login`);
  await page.getByRole('textbox', { name: 'Username' }).fill(USERNAME!);
  await page.getByRole('textbox', { name: 'Password' }).fill(PASSWORD!);
  await page.getByRole('button', { name: 'Sign In', exact: true }).click();
  await expect(page.locator('#twofa_code')).toBeVisible();
  await page.fill('#twofa_code', '000000');
  await page.click('button[type="submit"]');
  await expect(page.locator('.bg-red-50, .bg-red-500\\/10')).toBeVisible();
  await expect(page.locator('#twofa_code')).toBeVisible();
  // test with correct
  await page.goto(`/login`);
  await page.getByRole('textbox', { name: 'Username' }).fill(USERNAME!);
  await page.getByRole('textbox', { name: 'Password' }).fill(PASSWORD!);
  await page.getByRole('button', { name: 'Sign In', exact: true }).click();
  await expect(page.locator('#twofa_code')).toBeVisible();
  const totpToken = await generate({ secret: totpSecret });
  await page.fill('#twofa_code', totpToken);
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/.*dashboard/);
});

test('disable 2FA', async ({ page }) => {
  await page.goto(`/account/2fa`);
  await page.click('button:has-text("Click to disable 2FA")');
  await expect(page.getByText('2FA is currently disabled')).toBeVisible();
  // VERIFY dashboard shows 2FA disabled
  await page.goto(`/dashboard`);
  const dashboardTwofa = page.locator('#dashboard_twofa_content');
  await expect(dashboardTwofa).toBeVisible();
  await expect(dashboardTwofa.locator('b')).toHaveText('disabled');
  await page.goto(`/login`);
  await page.getByRole('textbox', { name: 'Username' }).fill(USERNAME!);
  await page.getByRole('textbox', { name: 'Password' }).fill(PASSWORD!);
  await page.getByRole('button', { name: 'Sign In', exact: true }).click();
  await expect(page).toHaveURL(/.*dashboard/);
});
