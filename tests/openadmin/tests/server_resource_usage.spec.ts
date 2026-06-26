import { test, expect } from '@playwright/test';

test('resource usage page shows drop-cache and clear-swap controls', async ({ page }) => {
  await page.goto('/server/resource-usage');
  await expect(page).toHaveURL(/server\/resource-usage/);

  await expect(page.locator('#human-readable-info')).toBeVisible({ timeout: 10000 });
  await expect(page.locator('#swap-human-readable-info')).toBeVisible();
  await expect(page.locator('#clear-cache')).toBeVisible();
  await expect(page.locator('#clear-swap')).toBeVisible();

  // NOTE: not clicked -- dropping cache/swap affects the live host's memory and could
  // interfere with other processes/tests running concurrently on shared infrastructure.
  console.log('resource usage page loaded with drop-cache/clear-swap controls present');
});

test('resource usage history page loads with default line view', async ({ page }) => {
  await page.goto('/server/resource-usage/history');

  await expect(page).toHaveURL(/server\/resource-usage\/history/);

  await expect(
    page.getByRole('heading', { name: 'Resource Usage History' })
  ).toBeVisible();

  await expect(page.locator('#chart-cpu')).toBeVisible();
  await expect(page.locator('#chart-mem')).toBeVisible();
  await expect(page.locator('#chart-disk')).toBeVisible();
  await expect(page.locator('#chart-conn')).toBeVisible();
});

test('time range selector changes period via query param', async ({ page }) => {
  await page.goto('/server/resource-usage/history');

  const select = page.locator('select').first();
  await select.selectOption('60');
  await expect(page).toHaveURL(/minutes=60/);
  await expect(page.locator('text=Last 1 hour').first()).toBeVisible();

  console.log('time range selector navigated to 1 hour period');
});

test('view toggle switches between line and grid views', async ({ page }) => {
  await page.goto('/server/resource-usage/history');

  await page.locator('a[title="Grid / table"]').click();
  await expect(page).toHaveURL(/view=grid/);
  await expect(page.locator('#view-line')).toBeHidden();

  await page.locator('a[title="Line chart"]').click();
  await expect(page).toHaveURL(/view=line/);
  await expect(page.locator('#view-line')).toBeVisible();

  console.log('view toggle switched grid <-> line');
});

test('refresh button reloads the page', async ({ page }) => {
  await page.goto('/server/resource-usage/history');

  await Promise.all([
    page.waitForLoadState('load'),
    page.getByRole('button', { name: 'Refresh' }).click(),
  ]);

  await expect(page.getByText('Resource Usage History')).toBeVisible();
  console.log('refresh button reloaded the page');
});
