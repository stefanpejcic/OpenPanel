import { test, expect, type Page } from '@playwright/test';

// ─── Config ──────────────────────────────────────────────────────────────────

const DOMAIN = 'wp.tests.openpanel.org';
const EMAILS = ['test1', 'test2'];
const PASSWORD = 'Password123_';

// ─── Helpers ─────────────────────────────────────────────────────────────────

async function getEmailCount(page: Page): Promise<number> {
  const text = await page.locator('#dashboard_usage_emails').locator('p').nth(1).textContent();
  if (!text) throw new Error('Cannot read email count');
  const match = text.match(/(\d+)\s*\//);
  if (!match) throw new Error(`Cannot parse email count from: ${text}`);
  return parseInt(match[1], 10);
}

async function createEmail(page: Page, username: string, password = PASSWORD) {
  await page.goto('/emails/new');
  await expect(page).toHaveURL(/emails\/new/);

  // Select first available domain
  const domainSelect = page.locator('select[name="domain"]');
  await expect(domainSelect).toBeVisible();
  await domainSelect.selectOption({ index: 1 });

  await page.getByRole('textbox', { name: /username/i }).fill(username);

  // Password field — page auto-generates one; overwrite it
  const passwordField = page.locator('input[name="password"]');
  await passwordField.fill(password);

  await page.getByRole('button', { name: /create email/i }).click();
  await expect(page.getByText(new RegExp(`Email ${username}@`, 'i'))).toBeVisible({ timeout: 15000 });
  console.log(`Created: ${username}@${DOMAIN}`);
}

async function expectEmailInTable(page: Page, username: string) {
  await page.goto('/emails');
  const row = page.locator('#email-accounts').getByText(new RegExp(`${username}@`));
  await expect(row).toBeVisible();
}

async function expectEmailNotInTable(page: Page, username: string) {
  await page.goto('/emails');
  await expect(page.locator('#email-accounts').getByText(new RegExp(`${username}@`))).toHaveCount(0);
}

async function deleteEmail(page: Page, username: string) {
  const address = `${username}@${DOMAIN}`;
  await page.goto(`/emails/delete/${address}`);
  await expect(page).toHaveURL(/emails\/delete/);
  await page.getByRole('button', { name: /confirm delete/i }).click();
  await expect(page).toHaveURL(/\/emails$/, { timeout: 10000 });
  console.log(`Deleted: ${address}`);
}

// ─── Accounts list ───────────────────────────────────────────────────────────

test('emails accounts page loads and shows table', async ({ page }) => {
  await page.goto('/emails');
  await expect(page).toHaveURL(/\/emails$/);
  await expect(page.getByRole('heading', { name: /emails/i })).toBeVisible();

  // Either has the table or the empty-state CTA
  const table    = page.locator('#email-accounts');
  const emptyMsg = page.getByText(/no emails yet/i);
  const hasTable = await table.isVisible().catch(() => false);
  const isEmpty  = await emptyMsg.isVisible().catch(() => false);
  expect(hasTable || isEmpty).toBe(true);
});

test('emails accounts search filters rows', async ({ page }) => {
  await page.goto('/emails');

  const rows = page.locator('#email-accounts tbody tr[x-show]');
  if (await rows.count() === 0) { test.skip(); return; }

  const search = page.locator('input[type="search"]').first();
  await search.fill('zzz_nomatch_xyz');
  for (const row of await rows.all()) {
    await expect(row).toBeHidden();
  }
  await search.fill('');
  await expect(rows.first()).toBeVisible();
});

test('emails accounts new-email button links to /emails/new', async ({ page }) => {
  await page.goto('/emails');
  const newBtn = page.locator('a[href="/emails/new"]').first();
  await expect(newBtn).toBeVisible();
});

test('emails accounts export button is present', async ({ page }) => {
  await page.goto('/emails');
  const exportLink = page.locator('a[href="/emails/export"]');
  // It is hidden by default (class="hidden") but should be in the DOM
  await expect(exportLink).toBeAttached();
});

// ─── Create ──────────────────────────────────────────────────────────────────

test('create email new page loads', async ({ page }) => {
  await page.goto('/emails/new');
  await expect(page).toHaveURL(/emails\/new/);
  await expect(page.getByRole('heading', { name: /create an email/i })).toBeVisible();

  // Required fields present
  await expect(page.locator('select[name="domain"]')).toBeVisible();
  await expect(page.locator('input[name="username"]')).toBeVisible();
  await expect(page.locator('input[name="password"]')).toBeVisible();
  await expect(page.locator('input[name="gb"]')).toBeVisible();
  await expect(page.locator('select[name="format"]')).toBeVisible();
  await expect(page.getByRole('button', { name: /create email/i })).toBeVisible();
});

test('create email password generate button fills field', async ({ page }) => {
  await page.goto('/emails/new');
  const passwordField = page.locator('input[name="password"]');
  const before = await passwordField.inputValue();
  await page.locator('#generatePassword').click();
  const after = await passwordField.inputValue();
  expect(after).not.toBe('');
  expect(after.length).toBeGreaterThanOrEqual(8);
});

test('create email toggle password visibility', async ({ page }) => {
  await page.goto('/emails/new');
  const passwordField = page.locator('input[name="password"]');
  // Page JS sets it to text on load; toggle to password
  await page.locator('#togglePassword').click();
  await expect(passwordField).toHaveAttribute('type', 'password');
  await page.locator('#togglePassword').click();
  await expect(passwordField).toHaveAttribute('type', 'text');
});

test('create emails and verify dashboard count', async ({ page }) => {
  await page.goto('/dashboard');
  const initialCount = await getEmailCount(page);
  let expected = initialCount;

  for (const username of EMAILS) {
    await createEmail(page, username);
    await expectEmailInTable(page, username);
    expected++;
    await page.goto('/dashboard');
    await expect.poll(() => getEmailCount(page), { timeout: 10000 }).toBe(expected);
  }
});

// ─── Edit / single account ───────────────────────────────────────────────────

test('edit email page loads for existing account', async ({ page }) => {
  // Ensure at least one email exists
  await page.goto('/emails');
  const firstRow = page.locator('#email-accounts tbody tr').first();
  const hasRows = await firstRow.isVisible().catch(() => false);
  if (!hasRows) { test.skip(); return; }

  const manageLink = firstRow.locator('a[href*="/emails/edit/"]');
  const href = await manageLink.getAttribute('href');
  await page.goto(href!);
  await expect(page).toHaveURL(/emails\/edit\/.+@/);

  await expect(page.getByRole('heading', { name: /edit email account/i })).toBeVisible();
  await expect(page.locator('input[name="gb"]')).toBeVisible();
  await expect(page.locator('select[name="format"]')).toBeVisible();
  await expect(page.locator('#suspend_incoming')).toBeVisible();
  await expect(page.locator('#suspend_outgoing')).toBeVisible();
  await expect(page.locator('#email_password')).toBeVisible();
  await expect(page.getByRole('button', { name: /update email settings/i })).toBeVisible();
});

test('edit email — suspend incoming', async ({ page }) => {
  for (const username of EMAILS) {
    const address = `${username}@${DOMAIN}`;
    await page.goto(`/emails/edit/${address}`);
    await expect(page).toHaveURL(/emails\/edit/);

    await page.locator('#suspend_incoming').check();
    await page.getByRole('button', { name: /update email settings/i }).click();

    const alert = page.locator('#alert-1');
    await expect(alert).toBeVisible({ timeout: 10000 });
    await expect(alert).toContainText(/settings saved for email/i);
  }
});

test('edit email — suspend outgoing', async ({ page }) => {
  for (const username of EMAILS) {
    const address = `${username}@${DOMAIN}`;
    await page.goto(`/emails/edit/${address}`);
    await expect(page).toHaveURL(/emails\/edit/);

    await page.locator('#suspend_outgoing').check();
    await page.getByRole('button', { name: /update email settings/i }).click();

    const alert = page.locator('#alert-1');
    await expect(alert).toBeVisible({ timeout: 10000 });
    await expect(alert).toContainText(/settings saved for email/i);
  }
});

test('edit email — restore allow incoming and outgoing', async ({ page }) => {
  for (const username of EMAILS) {
    const address = `${username}@${DOMAIN}`;
    await page.goto(`/emails/edit/${address}`);

    await page.locator('#allow_incoming').check();
    await page.locator('#allow_outgoing').check();
    await page.getByRole('button', { name: /update email settings/i }).click();

    const alert = page.locator('#alert-1');
    await expect(alert).toBeVisible({ timeout: 10000 });
    await expect(alert).toContainText(/settings saved for email/i);
  }
});

test('edit email — change password via generate button', async ({ page }) => {
  for (const username of EMAILS) {
    const address = `${username}@${DOMAIN}`;
    await page.goto(`/emails/edit/${address}`);

    await page.locator('#generatePassword').click();
    const newPass = await page.locator('#email_password').inputValue();
    expect(newPass.length).toBeGreaterThanOrEqual(8);

    await page.getByRole('button', { name: /update email settings/i }).click();
    const alert = page.locator('#alert-1');
    await expect(alert).toBeVisible({ timeout: 10000 });
    await expect(alert).toContainText(/settings saved for email/i);
  }
});

test('edit email — manual password change', async ({ page }) => {
  for (const username of EMAILS) {
    const address = `${username}@${DOMAIN}`;
    await page.goto(`/emails/edit/${address}`);

    await page.locator('#email_password').fill('NewPass_9876');
    await page.getByRole('button', { name: /update email settings/i }).click();
    const alert = page.locator('#alert-1');
    await expect(alert).toBeVisible({ timeout: 10000 });
    await expect(alert).toContainText(/settings saved for email/i);
  }
});

test('edit email — delete button links to delete page', async ({ page }) => {
  await page.goto('/emails');
  const firstManage = page.locator('a[href*="/emails/edit/"]').first();
  const hasLink = await firstManage.isVisible().catch(() => false);
  if (!hasLink) { test.skip(); return; }

  const href = await firstManage.getAttribute('href');
  await page.goto(href!);

  const deleteLink = page.locator(`a[href*="/emails/delete/"]`);
  await expect(deleteLink).toBeVisible();
  const deleteHref = await deleteLink.getAttribute('href');
  expect(deleteHref).toMatch(/\/emails\/delete\/.+@/);
});

// ─── Connect Devices / Info ───────────────────────────────────────────────────

test('connect devices page loads for an email', async ({ page }) => {
  await page.goto('/emails');
  const infoLink = page.locator('a[href*="/emails/info/"]').first();
  const hasLink = await infoLink.isVisible().catch(() => false);
  if (!hasLink) { test.skip(); return; }

  const href = await infoLink.getAttribute('href');
  await page.goto(href!);
  await expect(page).toHaveURL(/emails\/info\/.+@/);
  await expect(page.getByRole('heading', { name: /connect devices/i })).toBeVisible();

  // Server settings table
  await expect(page.getByText(/incoming server/i)).toBeVisible();
  await expect(page.getByText(/outgoing server/i)).toBeVisible();
  await expect(page.getByText(/imap port/i)).toBeVisible();
  await expect(page.getByText(/smtp port/i)).toBeVisible();
});

test('connect devices back button goes to /emails', async ({ page }) => {
  await page.goto('/emails');
  const infoLink = page.locator('a[href*="/emails/info/"]').first();
  const hasLink = await infoLink.isVisible().catch(() => false);
  if (!hasLink) { test.skip(); return; }

  await page.goto((await infoLink.getAttribute('href'))!);
  await page.getByRole('link', { name: /back to emails/i }).click();
  await expect(page).toHaveURL(/\/emails$/);
});

// ─── Webmail autologin ────────────────────────────────────────────────────────

test('webmail autologin link opens popup', async ({ page }) => {
  await page.goto('/emails');
  const webmailBtn = page.locator('a[data-email]').first();
  const hasBtn = await webmailBtn.isVisible().catch(() => false);
  if (!hasBtn) { test.skip(); return; }

  const popupPromise = page.waitForEvent('popup');
  await webmailBtn.click();
  const popup = await popupPromise;
  await popup.waitForLoadState('load');
  // Should land somewhere in webmail (roundcube, or a redirect)
  expect(popup.url()).toBeTruthy();
  await popup.close();
});

// ─── Filters ─────────────────────────────────────────────────────────────────

test('email filters selector page loads', async ({ page }) => {
  await page.goto('/emails/filter');
  await expect(page).toHaveURL(/emails\/filter/);
  await expect(page.getByRole('heading', { name: /email filters/i })).toBeVisible();
  await expect(page.locator('select')).toBeVisible();
});

test('email filters selector navigates to email filter page', async ({ page }) => {
  await page.goto('/emails/filter');

  const select = page.locator('select');
  const options = await select.locator('option').all();
  const emailOptions = options.filter(async o => (await o.getAttribute('value') || '').includes('@'));

  // Find first option with an email value
  let emailValue: string | null = null;
  for (const opt of options) {
    const val = await opt.getAttribute('value');
    if (val && val.includes('@')) { emailValue = val; break; }
  }

  if (!emailValue) { test.skip(); return; }

  await select.selectOption(emailValue);
  await expect(page).toHaveURL(new RegExp(`emails/filter/${emailValue.replace('+', '\\+')}`));
});

test('email filter GUI page loads for email', async ({ page }) => {
  await page.goto('/emails');
  const firstRow = page.locator('#email-accounts tbody tr').first();
  if (!await firstRow.isVisible().catch(() => false)) { test.skip(); return; }

  const address = await firstRow.locator('td').first().textContent();
  if (!address?.includes('@')) { test.skip(); return; }

  const email = address.trim();
  await page.goto(`/emails/filter/${email}/gui`);
  await expect(page.getByRole('heading', { name: /email filters/i })).toBeVisible();
  await expect(page.getByText(/sieve filters for/i)).toBeVisible();

  // Add filter button present
  await expect(page.getByRole('button', { name: /add filter/i })).toBeVisible();
});

test('email filter — add a filter rule in GUI mode', async ({ page }) => {
  await page.goto('/emails');
  const firstRow = page.locator('#email-accounts tbody tr').first();
  if (!await firstRow.isVisible().catch(() => false)) { test.skip(); return; }

  const email = (await firstRow.locator('td').first().textContent())?.trim();
  if (!email?.includes('@')) { test.skip(); return; }

  await page.goto(`/emails/filter/${email}/gui`);
  await page.getByRole('button', { name: /add filter/i }).click();

  // A filter card should appear
  const filterCard = page.locator('.rounded-lg.border').last();
  await expect(filterCard).toBeVisible();

  // Fill the filter name
  const nameInput = filterCard.locator('input[type="text"]').first();
  await nameInput.fill('Test Filter');
  await expect(nameInput).toHaveValue('Test Filter');
});

test('email filter raw mode shows textarea', async ({ page }) => {
  await page.goto('/emails');
  const firstRow = page.locator('#email-accounts tbody tr').first();
  if (!await firstRow.isVisible().catch(() => false)) { test.skip(); return; }

  const email = (await firstRow.locator('td').first().textContent())?.trim();
  if (!email?.includes('@')) { test.skip(); return; }

  await page.goto(`/emails/filter/${email}/raw`);
  const textarea = page.locator('textarea[name="new_content"]');
  await expect(textarea).toBeVisible();
  await expect(page.getByRole('button', { name: /save file/i })).toBeVisible();
});

// ─── Import ───────────────────────────────────────────────────────────────────

test('import emails page loads', async ({ page }) => {
  await page.goto('/emails/import');
  await expect(page).toHaveURL(/emails\/import/);
  await expect(page.getByRole('heading', { name: /import emails/i })).toBeVisible();
  await expect(page.locator('input[type="file"]')).toBeVisible();
  await expect(page.getByRole('button', { name: /upload/i })).toBeVisible();
});

test('import emails file input accepts csv and xlsx only', async ({ page }) => {
  await page.goto('/emails/import');
  const input = page.locator('input[type="file"]');
  const accept = await input.getAttribute('accept');
  expect(accept).toContain('.csv');
  expect(accept).toContain('.xls');
});

test('import emails back button links to /emails', async ({ page }) => {
  await page.goto('/emails/import');
  const backLink = page.getByRole('link', { name: /back to emails/i });
  await expect(backLink).toBeVisible();
  await backLink.click();
  await expect(page).toHaveURL(/\/emails$/);
});

// ─── Export ───────────────────────────────────────────────────────────────────

test('export emails returns a CSV download', async ({ page }) => {
  const [download] = await Promise.all([
    page.waitForEvent('download'),
    page.goto('/emails/export'),
  ]);
  expect(download.suggestedFilename()).toMatch(/\.csv$/);
});

// ─── Aliases ─────────────────────────────────────────────────────────────────

test('aliases list page loads', async ({ page }) => {
  await page.goto('/emails/aliases');
  await expect(page).toHaveURL(/emails\/aliases/);
  await expect(page.getByRole('heading', { name: /aliases/i })).toBeVisible();

  const table    = page.locator('table');
  const emptyMsg = page.getByText(/no aliases yet/i);
  const newBtn   = page.locator('a[href="/emails/aliases/new"]');

  await expect(newBtn).toBeVisible();
  const hasTable = await table.isVisible().catch(() => false);
  const isEmpty  = await emptyMsg.isVisible().catch(() => false);
  expect(hasTable || isEmpty).toBe(true);
});

test('aliases search filters rows', async ({ page }) => {
  await page.goto('/emails/aliases');
  const rows = page.locator('tbody tr[x-show]');
  if (await rows.count() === 0) { test.skip(); return; }

  const search = page.locator('input[type="search"]');
  await search.fill('zzz_nomatch_xyz');
  for (const row of await rows.all()) {
    await expect(row).toBeHidden();
  }
  await search.fill('');
});

test('new alias page loads', async ({ page }) => {
  await page.goto('/emails/aliases/new');
  await expect(page).toHaveURL(/emails\/aliases\/new/);
  await expect(page.getByRole('heading', { name: /create an alias/i })).toBeVisible();

  await expect(page.locator('select[name="domain"]')).toBeVisible();
  await expect(page.locator('input[name="username"]')).toBeVisible();
  await expect(page.locator('input[name="target"]')).toBeVisible();
  await expect(page.getByRole('button', { name: /create alias/i })).toBeVisible();
});

test('new alias — domain selector updates @domain preview', async ({ page }) => {
  await page.goto('/emails/aliases/new');

  const domainSelect = page.locator('select[name="domain"]');
  const options = await domainSelect.locator('option[value]').all();
  if (options.length === 0) { test.skip(); return; }

  const firstVal = await options[0].getAttribute('value');
  if (!firstVal) { test.skip(); return; }
  await domainSelect.selectOption(firstVal);

  const atPreview = page.locator('#at_address');
  await expect(atPreview).toHaveValue(`@${firstVal}`);
});

test('create alias, verify in list, then delete it', async ({ page }) => {
  await page.goto('/emails/aliases/new');

  const domainSelect = page.locator('select[name="domain"]');
  const options = await domainSelect.locator('option[value]').all();
  if (options.length === 0) { test.skip(); return; }

  const domain = await options[0].getAttribute('value');
  if (!domain) { test.skip(); return; }

  await domainSelect.selectOption(domain);
  await page.locator('input[name="username"]').fill('pw-test-alias');
  await page.locator('input[name="target"]').fill('target@external-test.example.com');
  await page.getByRole('button', { name: /create alias/i }).click();

  await expect(page).toHaveURL(/emails\/aliases/, { timeout: 10000 });
  const source = `pw-test-alias@${domain}`;
  await expect(page.getByText(source)).toBeVisible();

  // Delete it
  await page.goto(`/emails/aliases/delete/${source}`);
  await page.getByRole('button', { name: /confirm delete/i }).click();
  await expect(page).toHaveURL(/emails\/aliases/, { timeout: 10000 });
  await expect(page.getByText(source)).toHaveCount(0);
});

test('alias detail page loads for existing alias', async ({ page }) => {
  await page.goto('/emails/aliases');
  const manageLink = page.locator('a[href*="/emails/aliases/"]').filter({ hasText: /manage/i }).first();
  const hasLink = await manageLink.isVisible().catch(() => false);
  if (!hasLink) { test.skip(); return; }

  await manageLink.click();
  await expect(page).toHaveURL(/emails\/aliases\/.+@/);
  await expect(page.getByRole('heading', { name: /manage alias/i })).toBeVisible();
  await expect(page.locator('input#new-target')).toBeVisible();
  await expect(page.locator('#add-destination-btn')).toBeVisible();
});

test('alias detail — add then remove a destination', async ({ page }) => {
  await page.goto('/emails/aliases');
  const manageLink = page.locator('a[href*="/emails/aliases/"]').filter({ hasText: /manage/i }).first();
  if (!await manageLink.isVisible().catch(() => false)) { test.skip(); return; }

  await manageLink.click();
  await expect(page).toHaveURL(/emails\/aliases\/.+@/);

  const newTarget = 'pw-extra-dest@external-test.example.com';
  await page.locator('#new-target').fill(newTarget);
  await page.locator('#add-destination-btn').click();

  // Row with the new target should appear
  const newRow = page.locator(`tr[data-target="${newTarget}"]`);
  await expect(newRow).toBeVisible({ timeout: 8000 });

  // Remove it
  await newRow.locator('[data-action="remove-target"]').click();
  await expect(newRow).toBeHidden({ timeout: 8000 });
});

// ─── Default address ─────────────────────────────────────────────────────────

test('default address selector page loads', async ({ page }) => {
  await page.goto('/emails/default/');
  await expect(page).toHaveURL(/emails\/default/);
  await expect(page.getByRole('heading', { name: /default email address/i })).toBeVisible();
  await expect(page.locator('select#domains')).toBeVisible();
});

test('default address domain selector navigates to domain page', async ({ page }) => {
  await page.goto('/emails/default/');
  const select = page.locator('select#domains');
  const options = await select.locator('option[value]').all();
  if (options.length === 0) { test.skip(); return; }

  const domain = await options[0].getAttribute('value');
  if (!domain) { test.skip(); return; }

  await select.selectOption(domain);
  await expect(page).toHaveURL(new RegExp(`emails/default/${domain}`));
});

test('default address detail page shows current config or empty state', async ({ page }) => {
  await page.goto('/emails/default/');
  const select = page.locator('select#domains');
  const options = await select.locator('option[value]').all();
  if (options.length === 0) { test.skip(); return; }

  const domain = await options[0].getAttribute('value');
  if (!domain) { test.skip(); return; }

  await page.goto(`/emails/default/${domain}`);
  await expect(page.getByRole('heading', { name: /default email address/i })).toBeVisible();

  const destInput = page.locator('input#destination');
  await expect(destInput).toBeVisible();
  await expect(page.locator('#save-btn')).toBeVisible();
});

test('default address — set and clear catch-all', async ({ page }) => {
  await page.goto('/emails/default/');
  const select = page.locator('select#domains');
  const options = await select.locator('option[value]').all();
  if (options.length === 0) { test.skip(); return; }

  const domain = await options[0].getAttribute('value');
  if (!domain) { test.skip(); return; }

  await page.goto(`/emails/default/${domain}`);

  // Set a catch-all
  const dest = `catchall-pw-test@${domain}`;
  await page.locator('input#destination').fill(dest);
  await page.locator('#save-btn').click();
  await page.waitForLoadState('load');

  const currentDest = page.locator('#current-destination');
  if (await currentDest.isVisible().catch(() => false)) {
    await expect(currentDest).toContainText(dest.split('@')[0]);
  }

  // Clear it (danger zone delete)
  const deleteBtn = page.locator('#delete-btn');
  if (await deleteBtn.isVisible().catch(() => false)) {
    page.on('dialog', d => d.accept());
    await deleteBtn.click();
    await page.waitForLoadState('load');
    await expect(page.getByText(/no catch-all configured/i)).toBeVisible({ timeout: 8000 });
  }
});

// ─── Delete emails ────────────────────────────────────────────────────────────

test('delete emails and verify dashboard count', async ({ page }) => {
  await page.goto('/dashboard');
  const initialCount = await getEmailCount(page);
  let expected = initialCount;

  for (const username of EMAILS) {
    // Ensure it exists before trying to delete
    await page.goto('/emails');
    const row = page.locator('#email-accounts').getByText(new RegExp(`${username}@`));
    if (!await row.isVisible().catch(() => false)) continue;

    await deleteEmail(page, username);
    await expectEmailNotInTable(page, username);
    expected--;

    await page.goto('/dashboard');
    await expect.poll(() => getEmailCount(page), { timeout: 10000 }).toBe(expected);
  }
});

test('delete email page shows selector when no address given', async ({ page }) => {
  await page.goto('/emails/delete');
  await expect(page.getByRole('heading', { name: /select email/i })).toBeVisible();
  await expect(page.locator('select#domains')).toBeVisible();
});
