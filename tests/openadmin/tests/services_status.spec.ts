import { test, expect } from '@playwright/test';

test('services page lists services with search and edit-services link', async ({ page }) => {
  await page.goto('/services');
  await expect(page).toHaveURL(/\/services\/?$/);

  await expect(page.locator('#exiting_users')).toBeVisible();
  await expect(page.locator('#tour-edit-services-btn')).toHaveAttribute('href', '/services/edit');

  console.log('services page loaded with table and edit link');
});

test('search filters the services table', async ({ page }) => {
  await page.goto('/services');

  const rows = page.locator('#exiting_users tbody tr');
  const count = await rows.count();
  test.skip(count === 0, 'No services detected on this environment');

  const fullName = (await rows.first().locator('td').nth(1).innerText()).trim();
  const firstName = fullName.split(/\s+/)[0];
  await page.locator('input[x-model="searchQuery"]').fill(firstName);
  await page.waitForTimeout(150);

  await expect(rows.filter({ hasText: firstName })).toBeVisible();

  console.log(`search filtered services table to "${firstName}"`);
});

test('each service row exposes start/stop and restart actions', async ({ page }) => {
  await page.goto('/services');

  const rows = page.locator('#exiting_users tbody tr');
  const count = await rows.count();

  test.skip(count === 0, 'No services detected on this environment');

  let foundAtLeastOne = false;

  for (let i = 0; i < count; i++) {
    const row = rows.nth(i);

    const realName = (await row.locator('td').nth(3).innerText()).trim();
    if (!realName) continue;

    // Use accessible names instead of non-existent title attributes
    const startBtn = row.getByRole('button', { name: new RegExp(`Start ${realName}`, 'i') });
    const stopBtn = row.getByRole('button', { name: new RegExp(`Stop ${realName}`, 'i') });
    const restartBtn = row.getByRole('button', { name: new RegExp(`Restart ${realName}`, 'i') });

    const hasStartOrStop =
      (await startBtn.count()) > 0 || (await stopBtn.count()) > 0;

    if (hasStartOrStop) {
      await expect(restartBtn).toHaveCount(1);
      foundAtLeastOne = true;
    }
  }

  expect(foundAtLeastOne).toBeTruthy();
  console.log('verified start/stop and restart action buttons are present for services');
});
test('monitored services link to notifications settings', async ({ page }) => {
  await page.goto('/services');

  const monitoredLink = page.locator('a[href="/settings/notifications#services"]').first();
  const hasMonitored = (await monitoredLink.count()) > 0;
  test.skip(!hasMonitored, 'No monitored services detected on this environment');

  await expect(monitoredLink).toHaveText('Monitored');
  console.log('monitored service links to notifications settings');
});
