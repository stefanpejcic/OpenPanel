import { test, expect } from '@playwright/test';

test('notifications settings page loads with email/webhook fields', async ({ page }) => {
  await page.goto('/settings/notifications');
  await expect(page).toHaveURL(/settings\/notifications/);

  await expect(page.locator('[name="email"]')).toBeVisible();
  await expect(page.locator('[name="webhook_url"]')).toBeVisible();
  await expect(page.locator('[name="ssh_whitelist"]')).toBeVisible();

  console.log('notifications settings page loaded');
});

test('edit and save notification email, then revert', async ({ page }) => {
  await page.goto('/settings/notifications');

  const input = page.locator('[name="email"]');
  const original = await input.inputValue();
  const testValue = `test-${Date.now()}@example.com`;

  await input.fill(testValue);

  await Promise.all([
    page.waitForResponse(res =>
      res.url().includes('/settings/notifications') &&
      res.request().method() !== 'GET'
    ),
    page.locator('button[type="submit"]').first().click(),
  ]);

  await expect(input).toHaveValue(testValue);

  // revert
  await input.fill(original);

  await Promise.all([
    page.waitForResponse(res =>
      res.url().includes('/settings/notifications') &&
      res.request().method() !== 'GET'
    ),
    page.locator('button[type="submit"]').first().click(),
  ]);

  await expect(input).toHaveValue(original);
});

test('attack-prevention thresholds appear only when attack toggle is enabled', async ({ page }) => {
  // NOTE: verify-only -- the connection-limit thresholds gate live HTTP firewall rules;
  // we just check the conditional x-show wiring rather than submitting new values.
  await page.goto('/settings/notifications');

  const attackToggle = page.locator('#attack');
  const conditionalSection = page.locator('input[name="max_total_conn"]').locator('xpath=ancestor::div[@x-show="attackEnabled"][1]');

  const isChecked = await attackToggle.isChecked();
  if (isChecked) {
    await expect(conditionalSection).toBeVisible();
  } else {
    await expect(conditionalSection).toBeHidden();
  }

  console.log(`attack-prevention thresholds visibility matches toggle state (checked=${isChecked})`);
});

test('SMTP fields are present for mail delivery configuration', async ({ page }) => {
  // NOTE: verify-only -- saving SMTP settings affects how every outgoing OpenAdmin
  // notification email is sent; submitting test SMTP creds could break real delivery.
  await page.goto('/settings/notifications');

  for (const name of ['mail_server', 'mail_port', 'mail_username', 'mail_password', 'mail_default_sender']) {
    await expect(page.locator(`[name="${name}"]`)).toBeVisible();
  }

  console.log('SMTP configuration fields present (not saved)');
});

test('service monitoring checkboxes are present', async ({ page }) => {
  await page.goto('/settings/notifications');
  await expect(page).toHaveURL(/settings\/notifications/);

  const checkboxes = page.locator('input.toggle-checkbox');
  const count = await checkboxes.count();
  expect(count).toBeGreaterThan(0);

  console.log(`found ${count} notification toggle checkboxes`);
});
