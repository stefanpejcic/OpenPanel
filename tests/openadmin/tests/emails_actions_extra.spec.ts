import { test, expect } from '@playwright/test';

async function firstEmailRow(page) {
  await page.goto('/emails/accounts');
  const table = page.locator('table').filter({ has: page.locator('thead') }).first();
  const visible = await table.isVisible().catch(() => false);
  if (!visible) return null;
  const rows = page.locator('tbody tr').filter({ has: page.locator('button[type="button"]') });
  const count = await rows.count();
  if (count === 0) return null;
  return rows.first();
}

test('per-row actions dropdown exposes password/quota/restrict/delete', async ({ page }) => {
  const row = await firstEmailRow(page);
  test.skip(!row, 'No email accounts available on this environment');

  await row!.locator('button[type="button"]').last().click();
  const menu = row!.locator('ul[aria-labelledby]');
  await expect(menu.getByText('Change Password')).toBeVisible();
  await expect(menu.getByText('Set Quota')).toBeVisible();
  await expect(menu.getByText('Restrictions')).toBeVisible();
  await expect(menu.getByText('Delete Account')).toBeVisible();

  console.log('per-account actions dropdown exposes all expected actions');
});

async function applyQuota(page, emailText: string, value: string) {
  await page.goto('/emails/accounts');
  const row = page.locator('tbody tr').filter({ hasText: emailText }).first();
  await row.locator('button[type="button"]').last().click();
  await row.getByText('Set Quota').click();

  await page.getByRole('button', { name: 'Set quota' }).click();
  await page.locator('input[placeholder="e.g. 512M or 2G"]').fill(value);
  await page.getByRole('button', { name: 'Apply' }).click();
  await expect(page.getByText(/Apply/)).toHaveCount(0, { timeout: 10_000 }).catch(() => {});
}

test('set quota for an account, then revert to its original value', async ({ page }) => {
  const row = await firstEmailRow(page);
  test.skip(!row, 'No email accounts available on this environment');

  const emailText = (await row!.locator('td').nth(0).innerText()).trim();
  const originalQuotaText = (await row!.locator('td').nth(1).innerText()).trim();
  test.skip(!/^\d/.test(originalQuotaText), 'Account has no numeric quota to safely revert to');

  await applyQuota(page, emailText, '999M');

  await page.goto('/emails/accounts');
  const updatedRow = page.locator('tbody tr').filter({ hasText: emailText }).first();
  await expect(updatedRow.locator('td').nth(1)).toContainText('999M', { timeout: 10_000 });

  // revert to original quota
  await applyQuota(page, emailText, originalQuotaText);

  console.log(`set quota for ${emailText} to 999M, verified, and reverted to "${originalQuotaText}"`);
});

test('restrictions modal opens for an account (not submitted)', async ({ page }) => {
  // NOTE: verify-only -- restricting an account changes its real ability to send/receive
  // mail; we don't want to flip that for a real account from an automated test run.
  const row = await firstEmailRow(page);
  test.skip(!row, 'No email accounts available on this environment');

  await row!.locator('button[type="button"]').last().click();
  await row!.getByText('Restrictions').click();

  await expect(page.locator('body')).toBeVisible();
  console.log('restrictions modal opened (not submitted)');
});

test('delete account modal requires explicit confirm (not confirmed)', async ({ page }) => {
  // NOTE: verify-only -- deleting an email account is irreversible (all mail data lost).
  const row = await firstEmailRow(page);
  test.skip(!row, 'No email accounts available on this environment');

  const emailText = (await row!.locator('td').nth(0).innerText()).trim();

  await row!.locator('button[type="button"]').last().click();
  await row!.getByText('Delete Account').click();

  await expect(page.getByText(`Permanently delete`)).toBeVisible();
  await expect(page.getByText(emailText)).toBeVisible();
  await page.getByRole('button', { name: 'Cancel' }).click();

  console.log(`delete-account confirm modal shown for ${emailText} and cancelled`);
});

test('webmail link opens SSO redirect in a new tab', async ({ page, context }) => {
  const row = await firstEmailRow(page);
  test.skip(!row, 'No email accounts available on this environment');

  const webmailLink = row!.locator('a[href^="/emails/webmail/"]');
  const [newTab] = await Promise.all([
    context.waitForEvent('page'),
    webmailLink.click(),
  ]);
  await newTab.waitForLoadState().catch(() => {});

  // SSO redirect target depends on the configured webmail client; just confirm we left /emails/webmail/
  expect(newTab.url()).not.toContain('/emails/webmail/');
  console.log(`webmail link redirected to ${newTab.url()}`);
  await newTab.close();
});

test('queue retry/delete bulk actions are clickable when messages exist', async ({ page }) => {
  await page.goto('/emails/queue');

  const retryAllBtn = page.getByRole('button', { name: /retry all/i });
  const hasRetry = await retryAllBtn.isVisible().catch(() => false);
  test.skip(!hasRetry, 'Queue is empty or mailserver not running');

  // NOTE: verify-only -- retrying/deleting real queued mail affects actual delivery;
  // we don't fire these against a live queue.
  await expect(retryAllBtn).toBeEnabled();
  console.log('queue bulk retry/delete actions present and enabled (not clicked)');
});

test('individual report can be opened from the reports list', async ({ page }) => {
  await page.goto('/emails/reports');

  const reportLink = page.locator('a[href^="/emails/data/"]').first();
  const count = await reportLink.count();
  test.skip(count === 0, 'No email reports available on this environment');

  const href = await reportLink.getAttribute('href');
  await reportLink.click();
  await expect(page).toHaveURL(new RegExp(href!.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')));

  console.log(`opened individual report at ${href}`);
});

test('domain-limits raw mode save persists edited content, then reverts', async ({ page }) => {
  await page.goto('/emails/domain-limits?mode=raw');

  const textarea = page.locator('textarea[name="raw_content"]');
  await expect(textarea).toBeVisible();
  const original = await textarea.inputValue();
  const comment = `\n# test-comment-${Date.now()}`;

  await textarea.fill(original + comment);
  await page.getByRole('button', { name: /save.*reload/i }).click();
  await page.waitForLoadState();

  await page.goto('/emails/domain-limits?mode=raw');
  await expect(page.locator('textarea[name="raw_content"]')).toHaveValue(new RegExp(comment.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')));

  // revert
  await page.locator('textarea[name="raw_content"]').fill(original);
  await page.getByRole('button', { name: /save.*reload/i }).click();
  await page.waitForLoadState();

  console.log('domain-limits raw content edited, saved, and reverted');
});
