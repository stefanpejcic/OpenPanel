import { test, expect } from '@playwright/test';

test('active sessions: search, activity logs, terminate session', async ({ page, request }) => {

  const ipResponse = await request.get('https://ip.openpanel.com/');
  const ip = (await ipResponse.text()).trim();

  expect(ip).toBeTruthy();


  await page.goto('/account/sessions');

  const table = page.locator('#sessions-table');
  const searchInput = page.locator('[x-model="searchQuery"]');


  const visibleRows = table.locator('tbody tr:visible');


  const ipCell = table.locator('tbody tr', {
    hasText: ip,
  });


  if (await ipCell.count()) {
    await expect(ipCell.first()).toBeVisible();
  }


  await searchInput.fill('randomsearch');


  await page.waitForTimeout(300);

  await expect(visibleRows).toHaveCount(0);


  await searchInput.fill(ip);

  await page.waitForTimeout(300);

  const ipRows = table.locator('tbody tr', { hasText: ip });
  await expect(ipRows.first()).toBeVisible();
  await expect(ipRows.count()).resolves.toBeGreaterThan(0);


  const activityLink = ipRows.first().getByRole('link', {
    name: /view activity logs/i,
  });

  await activityLink.click();

  await expect(page).toHaveURL(new RegExp(`/account/activity\\?search=${ip}`));


  await page.goBack();
  await expect(page).toHaveURL(/\/account\/sessions/);


  const targetRow = table.locator('tbody tr', { hasText: ip }).first();

  const terminateButton = targetRow.getByRole('button', {
    name: /terminate/i,
  });

  await terminateButton.click();


  await expect(
    page.getByText('Session terminated successfully.')
  ).toBeVisible();
});
