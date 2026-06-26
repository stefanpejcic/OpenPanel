import { test, expect } from '@playwright/test';

test('notifications page loads with table and edit settings link', async ({ page }) => {
  await page.goto('/notifications');
  await expect(page).toHaveURL(/notifications/);

  await expect(page.locator('#tour-notifications-table')).toBeVisible();
  await expect(page.locator('#tour-edit-settings-link')).toHaveAttribute('href', '/settings/notifications');
  await expect(page.locator('#tour-bulk-actions')).toBeVisible();

  console.log('notifications page loaded');
});

test('search filters the notifications table', async ({ page }) => {
  await page.goto('/notifications');

  const rows = page.locator('#tour-notifications-table tbody tr');
  const initialCount = await rows.count();

  test.skip(initialCount === 0, 'No notifications recorded on this environment');

  const firstTitle = (await rows.first().locator('td').nth(1).innerText()).trim();
  test.skip(!firstTitle, 'Notification title cell empty, cannot build a meaningful search');

  const search = page.locator('#tour-notifications-search');
  await search.fill(firstTitle);

  const filteredRows = rows.filter({ hasText: firstTitle });

  await expect(filteredRows).toHaveCount(1);
  await expect(filteredRows.first()).toBeVisible();

  await expect(rows).toHaveCount(1);

  console.log(`search filtered notifications table to "${firstTitle}"`);
});

test('acknowledge a single unread notification', async ({ page }) => {
  await page.goto('/notifications');

  const unreadRow = page.locator('#tour-notifications-table tbody tr').filter({ has: page.locator('a[title="Click to Acknowledge notification"]') }).first();
  const count = await unreadRow.count();
  test.skip(count === 0, 'No unread notifications to acknowledge');

  await unreadRow.locator('a[title="Click to Acknowledge notification"]').click();
  await page.waitForLoadState();

  console.log('acknowledged a single unread notification');
});

test('delete single notification requires a second confirm click', async ({ page }) => {
  await page.goto('/notifications');

  const deleteLink = page.locator('a[title="Click to Delete notification"]').first();
  const count = await deleteLink.count();
  test.skip(count === 0, 'No notifications to test delete-confirm flow on');

  await deleteLink.click();
  await expect(page.getByText('Confirm', { exact: true }).first()).toBeVisible();

  // do not click again -- this is a destructive action; just verify the confirm gate exists
  await page.waitForTimeout(3100); // confirm state auto-reverts after 3s
  console.log('delete notification requires a second confirming click (not confirmed)');
});

test('acknowledge all button is present and submits without confirmation', async ({ page }) => {
  await page.goto('/notifications');

  const ackAllButton = page.getByRole('button', { name: 'Acknowledge All' });
  await expect(ackAllButton).toBeVisible();

  await ackAllButton.click();
  await page.waitForLoadState();
  await expect(page.locator('a[title="Click to Acknowledge notification"]')).toHaveCount(0);

  console.log('acknowledge all marked all notifications as read');
});

test('delete all button requires a second confirm click (not confirmed)', async ({ page }) => {
  // NOTE: verify-only -- "Delete All" permanently wipes notification history with no undo;
  // we only verify the two-click confirm gate, never actually confirm the deletion.
  await page.goto('/notifications');

  const deleteAllButton = page.getByRole('button', { name: 'Delete All' });
  await expect(deleteAllButton).toBeVisible();

  await deleteAllButton.click();
  await expect(page.getByText('Click to confirm')).toBeVisible();

  console.log('delete all shows confirm-to-proceed state (not confirmed)');
});
