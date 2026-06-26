import { test, expect } from '@playwright/test';

test('log viewer page loads with selects and settings link', async ({ page }) => {
  await page.goto('/services/logs');
  await expect(page).toHaveURL(/services\/logs/);

  await expect(page.locator('#log-select')).toBeVisible();
  await expect(page.locator('#lines-select')).toBeVisible();
  await expect(page.locator('#tour-log-settings-link')).toHaveAttribute('href', '/services/logs/edit');
  await expect(page.locator('#log-content')).toHaveText('Select a log file to view its content here.');

  console.log('log viewer page loaded with selects and settings link');
});

test('selecting a log file loads its content and reveals download/delete buttons', async ({ page }) => {
  await page.goto('/services/logs');

  const select = page.locator('#log-select');
  const options = select.locator('optgroup option');
  const count = await options.count();
  test.skip(count === 0, 'No log files registered to view');

  const firstValue = await options.first().getAttribute('value');
  await select.selectOption(firstValue!);

  await expect(page.locator('#log-content')).not.toHaveText('Select a log file to view its content here.', { timeout: 10_000 });
  await expect(page.locator('#download-btn')).toBeVisible();
  await expect(page.locator('#truncate-btn')).toBeVisible();

  console.log(`loaded log content for "${firstValue}", download/delete buttons revealed`);
});

test('lines-select limits the number of fetched log lines via URL params', async ({ page }) => {
  await page.goto('/services/logs');

  const select = page.locator('#log-select');
  const options = select.locator('optgroup option');
  const count = await options.count();
  test.skip(count === 0, 'No log files registered to view');

  const firstValue = await options.first().getAttribute('value');
  await select.selectOption(firstValue!);
  await page.locator('#lines-select').selectOption('20');

  await expect(page).toHaveURL(/lines=20/);
  console.log('lines select updated URL with lines=20');
});

test('edit log paths page validates JSON before submitting', async ({ page }) => {
  await page.goto('/services/logs/edit');
  await expect(page).toHaveURL(/services\/logs\/edit/);

  const textarea = page.locator('textarea[name="data"]');
  await expect(textarea).toBeVisible();
  const original = await textarea.inputValue();

  // submit invalid JSON, should be rejected client-side and stay on the page
  await textarea.fill('{not valid json');
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page).toHaveURL(/services\/logs\/edit/);
  await expect(textarea).toHaveValue('{not valid json');

  // restore original content without saving invalid data
  await textarea.fill(original);

  console.log('edit log paths rejected invalid JSON client-side');
});
