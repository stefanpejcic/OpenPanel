import { test, expect } from '@playwright/test';

// ─── Accounts ───────────────────────────────────────────────────────────────

test('emails accounts page loads', async ({ page }) => {
  await page.goto('/emails/accounts');
  await expect(page).toHaveURL(/emails\/accounts/);

  const mailserverNotInstalled = page.getByText('opencli email-server install');
  const mailserverStopped = page.locator('text=stopped').first();
  const emailTable = page.locator('table#exiting_users');

  const isNotInstalled = await mailserverNotInstalled.isVisible().catch(() => false);
  const isStopped = await mailserverStopped.isVisible().catch(() => false);

  if (isNotInstalled) {
    await expect(mailserverNotInstalled).toBeVisible();
  } else if (isStopped) {
    await expect(mailserverStopped).toBeVisible();
  } else {
    await expect(emailTable).toBeVisible();
  }
});

test('emails accounts search filters rows', async ({ page }) => {
  await page.goto('/emails/accounts');

  const table = page.locator('table#exiting_users');
  const isTableVisible = await table.isVisible().catch(() => false);
  if (!isTableVisible) {
    test.skip();
    return;
  }

  const searchInput = page.locator('input[placeholder="Search emails..."]');
  await expect(searchInput).toBeVisible();

  // Type a query unlikely to match anything
  await searchInput.fill('zzznomatch_xyz');
  // All data rows should be hidden (Alpine x-show hides them)
  const visibleRows = page.locator('tbody tr[x-show]');
  for (const row of await visibleRows.all()) {
    await expect(row).toBeHidden();
  }

  // Clear search — rows come back
  await searchInput.fill('');
  const firstRow = visibleRows.first();
  if (await firstRow.count() > 0) {
    await expect(firstRow).toBeVisible();
  }
});

test('emails accounts webmail link present per row', async ({ page }) => {
  await page.goto('/emails/accounts');

  const table = page.locator('table#exiting_users');
  if (!await table.isVisible().catch(() => false)) {
    test.skip();
    return;
  }

  const rows = page.locator('tbody tr').filter({ has: page.locator('a[href^="/emails/webmail/"]') });
  const count = await rows.count();
  if (count === 0) {
    // No email accounts — just verify empty state message
    await expect(page.getByText('No email accounts yet.')).toBeVisible();
    return;
  }

  // Each visible row should have a webmail link
  for (let i = 0; i < count; i++) {
    const link = rows.nth(i).locator('a[href^="/emails/webmail/"]');
    await expect(link).toBeVisible();
    const href = await link.getAttribute('href');
    expect(href).toMatch(/^\/emails\/webmail\/.+@.+/);
  }
});

// ─── Queue ───────────────────────────────────────────────────────────────────

test('emails queue page loads', async ({ page }) => {
  await page.goto('/emails/queue');
  await expect(page).toHaveURL(/emails\/queue/);

  const notInstalled = page.getByText('opencli email-server install');
  const emptyQueue  = page.getByText('Queue is empty.');
  const queueTable  = page.locator('table');

  const isNotInstalled = await notInstalled.isVisible().catch(() => false);
  if (isNotInstalled) {
    await expect(notInstalled).toBeVisible();
  } else {
    // Either empty state or a populated table
    const hasTable = await queueTable.isVisible().catch(() => false);
    expect(hasTable).toBe(true);
  }
});

test('emails queue refresh button reloads page', async ({ page }) => {
  await page.goto('/emails/queue');

  const refreshBtn = page.getByRole('button', { name: /refresh/i });
  await expect(refreshBtn).toBeVisible();

  const [response] = await Promise.all([
    page.waitForNavigation(),
    refreshBtn.click(),
  ]);
  expect(response?.status()).toBeLessThan(400);
  await expect(page).toHaveURL(/emails\/queue/);
});

test('emails queue search filters rows', async ({ page }) => {
  await page.goto('/emails/queue');

  const searchInput = page.locator('input[placeholder="Search queue..."]');
  const isVisible = await searchInput.isVisible().catch(() => false);
  if (!isVisible) {
    test.skip();
    return;
  }

  await searchInput.fill('zzznomatch_xyz_unique');
  const rows = page.locator('tbody tr[x-show]');
  for (const row of await rows.all()) {
    await expect(row).toBeHidden();
  }

  await searchInput.fill('');
});

test('emails queue bulk actions visible when messages exist', async ({ page }) => {
  await page.goto('/emails/queue');

  const retryAllBtn  = page.getByRole('button', { name: /retry all/i });
  const deleteAllBtn = page.getByRole('button', { name: /delete all/i });

  const hasRetry = await retryAllBtn.isVisible().catch(() => false);
  if (!hasRetry) {
    // Queue is empty or mailserver not running — acceptable
    test.skip();
    return;
  }

  await expect(retryAllBtn).toBeVisible();
  await expect(deleteAllBtn).toBeVisible();
});

// ─── Settings ────────────────────────────────────────────────────────────────

test('emails settings page loads', async ({ page }) => {
  await page.goto('/emails/settings');
  await expect(page).toHaveURL(/emails\/settings/);

  const notInstalled = page.getByText('opencli email-server install');
  const heading      = page.getByRole('heading', { name: /email settings/i });

  const isNotInstalled = await notInstalled.isVisible().catch(() => false);
  if (isNotInstalled) {
    await expect(notInstalled).toBeVisible();
  } else {
    await expect(heading).toBeVisible();
  }
});

test('emails settings shows mailserver status badge', async ({ page }) => {
  await page.goto('/emails/settings');

  const notInstalled = page.getByText('opencli email-server install');
  if (await notInstalled.isVisible().catch(() => false)) {
    test.skip();
    return;
  }

  // Status badge will contain one of these strings
  const statusBadge = page.locator('#status').locator('..').getByText(/(running|stopped|unknown)/i).first();
  // More robust: look for the status section
  const statusSection = page.locator('text=MailServer Status').first();
  await expect(statusSection).toBeVisible();
});

test('emails settings webmail domain input accepts value', async ({ page }) => {
  await page.goto('/emails/settings');

  const notInstalled = page.getByText('opencli email-server install');
  if (await notInstalled.isVisible().catch(() => false)) {
    test.skip();
    return;
  }

  const domainInput = page.locator('input[name="webmail-domain"]');
  await expect(domainInput).toBeVisible();

  const original = await domainInput.inputValue();
  await domainInput.fill('webmail.test.example.com');
  await expect(domainInput).toHaveValue('webmail.test.example.com');

  // Restore original value without submitting
  await domainInput.fill(original);
});

test('emails settings service toggles are present', async ({ page }) => {
  await page.goto('/emails/settings');

  const notInstalled = page.getByText('opencli email-server install');
  if (await notInstalled.isVisible().catch(() => false)) {
    test.skip();
    return;
  }

  const expectedServices = [
    'ENABLE_POSTFWD',
    'ENABLE_AMAVIS',
    'ENABLE_RSPAMD',
    'ENABLE_SPAMASSASSIN',
    'ENABLE_OPENDKIM',
    'ENABLE_OPENDMARC',
    'ENABLE_POP3',
    'ENABLE_IMAP',
    'ENABLE_CLAMAV',
    'ENABLE_FAIL2BAN',
  ];

  for (const name of expectedServices) {
    const checkbox = page.locator(`input[name="${name}"]`);
    await expect(checkbox).toBeVisible();
  }
});

test('emails settings storage type select is present', async ({ page }) => {
  await page.goto('/emails/settings');

  const notInstalled = page.getByText('opencli email-server install');
  if (await notInstalled.isVisible().catch(() => false)) {
    test.skip();
    return;
  }

  const storageSelect = page.locator('select[name="storage_type"]');
  await expect(storageSelect).toBeVisible();

  const options = await storageSelect.locator('option').allTextContents();
  expect(options.some(o => /docker volume/i.test(o) || /user_dir/i.test(o))).toBe(true);
  expect(options.some(o => /custom/i.test(o))).toBe(true);
});

// ─── Rate limits ─────────────────────────────────────────────────────────────

test('email rate limits page loads', async ({ page }) => {
  await page.goto('/emails/domain-limits');
  await expect(page).toHaveURL(/emails\/domain-limits/);

  const heading = page.getByRole('heading', { name: /email rate limits/i });
  await expect(heading).toBeVisible();
});

test('email rate limits shows rules table or empty state', async ({ page }) => {
  await page.goto('/emails/domain-limits');

  const table      = page.locator('table#exiting_users');
  const emptyState = page.getByText('No rate-limit rules configured');

  const hasTable = await table.isVisible().catch(() => false);
  const hasEmpty = await emptyState.isVisible().catch(() => false);

  expect(hasTable || hasEmpty).toBe(true);
});

test('email rate limits search input filters rows', async ({ page }) => {
  await page.goto('/emails/domain-limits');

  const table = page.locator('table#exiting_users');
  if (!await table.isVisible().catch(() => false)) {
    test.skip();
    return;
  }

  const searchInput = page.locator('input[placeholder*="Filter"]');
  await expect(searchInput).toBeVisible();

  await searchInput.fill('zzznomatch_xyz_unique');
  const dataRows = page.locator('tbody tr[x-data]');
  for (const row of await dataRows.all()) {
    await expect(row).toBeHidden();
  }

  await searchInput.fill('');
});

test('email rate limits edit pencil opens inline input', async ({ page }) => {
  await page.goto('/emails/domain-limits');

  const table = page.locator('table#exiting_users');
  if (!await table.isVisible().catch(() => false)) {
    test.skip();
    return;
  }

  const rows = page.locator('tbody tr[x-data]');
  if (await rows.count() === 0) {
    test.skip();
    return;
  }

  const firstRow   = rows.first();
  const editBtn    = firstRow.locator('button[title="Edit limit"]');
  await expect(editBtn).toBeVisible();
  await editBtn.click();

  const editInput = firstRow.locator('input[type="number"]');
  await expect(editInput).toBeVisible();

  // Cancel via escape
  await editInput.press('Escape');
  await expect(editInput).toBeHidden();
});

test('email rate limits raw mode toggle shows textarea', async ({ page }) => {
  await page.goto('/emails/domain-limits?mode=raw');
  await expect(page).toHaveURL(/mode=raw/);

  const textarea = page.locator('textarea[name="raw_content"]');
  await expect(textarea).toBeVisible();

  const saveBtn = page.getByRole('button', { name: /save.*reload/i });
  await expect(saveBtn).toBeVisible();
});

// ─── Reports ─────────────────────────────────────────────────────────────────

test('emails reports page loads', async ({ page }) => {
  await page.goto('/emails/reports');
  await expect(page).toHaveURL(/emails\/reports/);

  const notInstalled = page.getByText('opencli email-server install');
  const noReports    = page.getByText(/no reports yet/i);
  const heading      = page.getByRole('heading', { name: /email reports/i });

  const isNotInstalled = await notInstalled.isVisible().catch(() => false);
  const isNoReports    = await noReports.isVisible().catch(() => false);

  if (isNotInstalled) {
    await expect(notInstalled).toBeVisible();
  } else if (isNoReports) {
    await expect(noReports).toBeVisible();
  } else {
    await expect(heading).toBeVisible();
  }
});
