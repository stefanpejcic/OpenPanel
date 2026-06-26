import { test, expect } from '@playwright/test';

test('timezone page loads with current timezone and select options', async ({ page }) => {
  await page.goto('/server/timezone');
  await expect(page).toHaveURL(/server\/timezone/);

  await expect(page.getByText('Current timezone:')).toBeVisible();

  const select = page.locator('#timezone');
  await expect(select).toBeVisible();
  const optionCount = await select.locator('option').count();
  expect(optionCount).toBeGreaterThan(100); // pytz.all_timezones has 400+ entries

  console.log(`timezone page loaded with ${optionCount} timezone options`);
});

test('change timezone updates host setting and can be reverted', async ({ page }) => {
  await page.goto('/server/timezone');

  const select = page.locator('#timezone');
  const original = await select.inputValue();
  const target = original === 'UTC' ? 'Europe/Belgrade' : 'UTC';

  await select.selectOption(target);
  await page.getByRole('button', { name: 'Change Timezone' }).click();

  // route redirects back to GET with no flash message (known issue), so verify via the select value
  await expect(page).toHaveURL(/server\/timezone/);
  await expect(page.locator('#timezone')).toHaveValue(target);
  await expect(page.getByText(`Current timezone:`)).toContainText(target);

  // revert to original timezone
  await page.locator('#timezone').selectOption(original);
  await page.getByRole('button', { name: 'Change Timezone' }).click();
  await expect(page.locator('#timezone')).toHaveValue(original);

  console.log(`timezone changed ${original} -> ${target} -> ${original}`);
});
