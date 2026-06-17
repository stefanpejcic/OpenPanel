import { test, expect } from '@playwright/test';

const domain = 'wp.tests.openpanel.org';


test('capitalize domains page loads', async ({ page }) => {
  await page.goto(`/domains/capitalize?domain=${domain}`);
  await expect(page).toHaveURL(/domains\/capitalize/);
  await expect(page.locator('body')).toContainText(/capitalize|letter|domain/i);
  console.log('capitalize domains page accessible');
});


test('letter buttons are rendered for domain', async ({ page }) => {
  await page.goto(`/domains/capitalize?domain=${domain}`);

  // the template renders one button per letter of the domain
  const letterButtons = page.locator('#lettersContainer button, button.cursor-pointer');
  await expect(letterButtons.first()).toBeVisible({ timeout: 5000 });
  const count = await letterButtons.count();
  expect(count).toBeGreaterThan(0);
  console.log(`capitalize: ${count} letter buttons rendered`);
});


test('toggle a letter to uppercase and save', async ({ page }) => {
  await page.goto(`/domains/capitalize?domain=${domain}`);

  const letterButtons = page.locator('#lettersContainer button');
  await expect(letterButtons.first()).toBeVisible({ timeout: 5000 });

  // click first lowercase letter to make it uppercase
  const firstBtn = letterButtons.first();
  const letterBefore = await firstBtn.textContent();
  await firstBtn.click();
  const letterAfter = await firstBtn.textContent();

  // should have toggled case
  expect(letterAfter?.toLowerCase()).toBe(letterBefore?.toLowerCase());
  expect(letterAfter).not.toBe(letterBefore);

  // submit
  await page.getByRole('button', { name: /save|apply|submit|capitalize/i }).click();
  await expect(page.locator('body')).toContainText(/saved|updated|success/i, { timeout: 10000 });
  console.log('capitalize domain letter toggle and save working');
});


test('revert domain capitalization to original', async ({ page }) => {
  await page.goto(`/domains/capitalize?domain=${domain}`);

  const letterButtons = page.locator('#lettersContainer button');
  await expect(letterButtons.first()).toBeVisible({ timeout: 5000 });

  // ensure all letters are lowercase (click uppercase ones to toggle back)
  const count = await letterButtons.count();
  for (let i = 0; i < count; i++) {
    const btn = letterButtons.nth(i);
    const text = await btn.textContent();
    if (text && text === text.toUpperCase() && text !== text.toLowerCase()) {
      await btn.click();
    }
  }

  await page.getByRole('button', { name: /save|apply|submit|capitalize/i }).click();
  await expect(page.locator('body')).toContainText(/saved|updated|success/i, { timeout: 10000 });
  console.log('domain capitalization reverted to original');
});
