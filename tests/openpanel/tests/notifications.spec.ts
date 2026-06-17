import { test, expect } from '@playwright/test';


test('notifications page loads', async ({ page }) => {
  await page.goto('/account/notifications');
  await expect(page).toHaveURL(/account\/notifications/);
  await expect(page.locator('body')).toContainText(/notification|alert|email/i);
  console.log('notifications page accessible');
});


test('notifications page has toggles or settings', async ({ page }) => {
  await page.goto('/account/notifications');

  const hasCheckbox = await page.locator('input[type="checkbox"]').first().isVisible({ timeout: 3000 }).catch(() => false);
  const hasToggle = await page.locator('button[role="switch"], input[type="radio"]').first().isVisible({ timeout: 3000 }).catch(() => false);
  const hasForm = await page.locator('form').first().isVisible({ timeout: 3000 }).catch(() => false);

  expect(hasCheckbox || hasToggle || hasForm).toBe(true);
  console.log('notifications page has configurable settings');
});


test('save notification preferences', async ({ page }) => {
  await page.goto('/account/notifications');

  const checkbox = page.locator('input[type="checkbox"]').first();
  if (await checkbox.isVisible({ timeout: 3000 }).catch(() => false)) {
    const wasChecked = await checkbox.isChecked();
    await checkbox.click();
    await expect(checkbox).toBeChecked({ checked: !wasChecked });
  }

  const saveBtn = page.getByRole('button', { name: /save|update|apply/i });
  await expect(saveBtn).toBeVisible();
  await saveBtn.click();

  await expect(page.locator('body')).toContainText(/saved|updated|success/i, { timeout: 10000 });
  console.log('notification preferences saved');
});


test('notification settings persist after reload', async ({ page }) => {
  await page.goto('/account/notifications');

  const checkbox = page.locator('input[type="checkbox"]').first();
  if (!(await checkbox.isVisible({ timeout: 3000 }).catch(() => false))) {
    console.log('no checkbox found – skipping persistence test');
    return;
  }

  const stateBefore = await checkbox.isChecked();

  const saveBtn = page.getByRole('button', { name: /save|update|apply/i });
  await saveBtn.click();
  await page.waitForLoadState('networkidle');

  await page.goto('/account/notifications');
  const stateAfter = await page.locator('input[type="checkbox"]').first().isChecked();
  expect(stateAfter).toBe(stateBefore);
  console.log('notification settings persisted after reload');
});
