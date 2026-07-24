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


test('emails accounts search filters rows', async ({ page }) => {
  await page.goto('/emails');

  const rows = page.locator('#email-accounts tbody tr[x-show]');
  if (await rows.count() === 0) { test.skip(); return; }

  const search = page
    .getByRole('region', { name: 'Emails Header' })
    .getByRole('searchbox', { name: 'Search' });

  await search.fill('zzz_nomatch_xyz');
  await expect(rows).toHaveCount(await rows.count()); // rows stay in DOM, x-show hides them
  for (const row of await rows.all()) {
    await expect(row).toBeHidden();
  }

  await search.fill('');
  await expect(rows.first()).toBeVisible();
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

test('webmail autologin and send/receive', async ({ page }) => {
  test.setTimeout(90_000);

  await page.goto('/emails');
  const webmailBtn = page.locator('a[href*="/webmail/"]').first();
  const hasBtn = await webmailBtn.isVisible().catch(() => false);
  if (!hasBtn) { test.skip(); return; }

  const emailAddress = await webmailBtn.getAttribute('href').then(h => h!.split('/webmail/')[1]);
  expect(emailAddress).toBeTruthy();

  // --- Open webmail popup ---
  const popupPromise = page.waitForEvent('popup');
  await webmailBtn.click();
  const popup = await popupPromise;
  await popup.waitForLoadState('load');
  expect(popup.url()).toBeTruthy();

  // --- Wait for Roundcube inbox to be ready ---
  await popup.waitForSelector('#mailboxlist', { timeout: 20_000 });

  // --- Click compose ---
  await popup.locator('a.compose, button.compose').first().click();
  await popup.waitForSelector('#compose-subject', { timeout: 10_000 });

  // --- Fill To via the visible token widget input ---
  const toInput = popup.locator('.token-input input, #compose-to .token-input, #compose_to .recipient input').first();
  const hasTokenInput = await toInput.isVisible().catch(() => false);
  if (hasTokenInput) {
    await toInput.fill(emailAddress!);
    await popup.keyboard.press('Enter');
  } else {
    // fallback: force into the hidden textarea
    await popup.locator('#_to').scrollIntoViewIfNeeded();
    await popup.locator('#_to').pressSequentially(emailAddress!, { delay: 50 });
    await popup.keyboard.press('Enter');
  }

  // --- Subject ---
  const subject = `Playwright self-send ${Date.now()}`;
  await popup.locator('#compose-subject').fill(subject);

  // --- Body ---
  const bodyText = 'Automated Playwright send/receive test.';
  const tinyFrame = popup.frameLocator('#composebody_ifr');
  const hasHtmlEditor = await tinyFrame.locator('body').isVisible().catch(() => false);
  if (hasHtmlEditor) {
    await tinyFrame.locator('body').fill(bodyText);
  } else {
    await popup.locator('#composebody').fill(bodyText);
  }

  // --- Send ---
  await popup.locator('button.btn-primary.send').click();
  await popup.waitForSelector('#messagestack .confirmation, div.notice.confirmation', {
    timeout: 15_000,
  });

  // --- Sent folder ---
  await popup.locator('#mailboxlist a[rel="sent"], #mailboxlist .sent a').first().click();
  await popup.waitForLoadState('networkidle');
  await expect(popup.locator('#messagelist tbody tr')).toContainText(subject, { timeout: 10_000 });
  
  // --- Inbox ---
  await popup.locator('#mailboxlist a[rel="inbox"], #mailboxlist .inbox a').first().click();
  await popup.waitForLoadState('networkidle');
  
  await expect.poll(
    async () => {
      await popup.reload();
      await popup.waitForLoadState('networkidle');
      return popup.locator('#messagelist tbody tr').filter({ hasText: subject }).count();
    },
    { timeout: 60_000, intervals: [5_000] },
  ).toBeGreaterThan(0);

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

  if (!(await firstRow.isVisible().catch(() => false))) {
    test.skip();
    return;
  }

  const email = (await firstRow.locator('td').first().textContent())?.trim();

  if (!email?.includes('@')) {
    test.skip();
    return;
  }

  await page.goto(`/emails/filter/${encodeURIComponent(email)}/gui`);

  const filterNameInputs = page.getByPlaceholder('Filter name…');
  const existingFilterCount = await filterNameInputs.count();

  await page.getByRole('button', { name: /^add filter$/i }).click();

  // Wait until Alpine adds one new rendered filter.
  await expect(filterNameInputs).toHaveCount(existingFilterCount + 1);

  const nameInput = filterNameInputs.nth(existingFilterCount);
  await expect(nameInput).toBeVisible();

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
  const textarea = page.locator('#raw_content');
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
  await page.goto('/emails');
  const [download] = await Promise.all([
    page.waitForEvent('download'),
    page.evaluate(() => { window.location.href = '/emails/export'; }),
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

  const hasTable = await table.isVisible().catch(() => false);
  const isEmpty  = await emptyMsg.isVisible().catch(() => false);
  expect(hasTable || isEmpty).toBe(true);
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

  const domainSelect = page.getByLabel('Domain');
  await expect(domainSelect).toBeVisible();

  const validOptions = domainSelect.locator('option[value]:not([value=""])');
  const optionCount = await validOptions.count();

  if (optionCount === 0) {
    test.skip();
    return;
  }

  const firstValue = await validOptions.first().getAttribute('value');

  if (!firstValue) {
    test.skip();
    return;
  }

  await domainSelect.selectOption(firstValue);

  const atPreview = page.locator('#at_address');
  await expect(atPreview).toHaveValue(`@${firstValue}`);
});

const TEST_ALIAS = 'pw-test-alias';
const TEST_TARGET = 'target@external-test.example.com';

test('create alias and verify in list', async ({ page }) => {
  await page.goto('/emails/aliases/new');

  const domainSelect = page.getByLabel('Domain');
  await expect(domainSelect).toBeVisible();

  const validOptions = domainSelect.locator(
    'option[value]:not([value=""]):not([disabled])'
  );

  if ((await validOptions.count()) === 0) {
    test.skip();
    return;
  }

  const domain = await validOptions.first().getAttribute('value');

  if (!domain) {
    test.skip();
    return;
  }

  const source = `${TEST_ALIAS}@${domain}`;

  // Clean up if it already exists.
  await page.goto('/emails/aliases');
  if (await page.getByText(source, { exact: true }).count()) {
    await page.goto(`/emails/aliases/delete/${encodeURIComponent(source)}`);
    await page.getByRole('button', { name: /confirm delete/i }).click();
    await expect(page).toHaveURL(/\/emails\/aliases\/?$/);
  }

  await page.goto('/emails/aliases/new');

  await domainSelect.selectOption(domain);
  await page.getByLabel(/alias address/i).fill(TEST_ALIAS);
  await page.getByLabel(/destination address/i).fill(TEST_TARGET);

  await page.getByRole('button', { name: /create alias/i }).click();

  await expect(page).toHaveURL(/\/emails\/aliases\/?$/, {
    timeout: 10_000,
  });

  await expect(page.getByText(source, { exact: true })).toBeVisible();
});

test('aliases search filters rows', async ({ page }) => {
  await page.goto('/emails/aliases');

  const aliasesSection = page.getByRole('region', {
    name: 'Aliases Header',
  });

  const rows = aliasesSection.locator('#aliases-table tbody tr[x-show]');
  const rowCount = await rows.count();

  if (rowCount === 0) {
    test.skip();
    return;
  }

  const search = aliasesSection.getByPlaceholder('Search');
  await expect(search).toBeVisible();

  await search.fill('zzz_nomatch_xyz');

  for (let i = 0; i < rowCount; i++) {
    await expect(rows.nth(i)).toBeHidden();
  }

  await search.fill('');

  for (let i = 0; i < rowCount; i++) {
    await expect(rows.nth(i)).toBeVisible();
  }
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

test('delete alias and verify removal', async ({ page }) => {
  await page.goto('/emails/aliases');

  const aliasLink = page.getByText(new RegExp(`^${TEST_ALIAS}@`));

  if (!(await aliasLink.count())) {
    test.skip();
    return;
  }

  const source = (await aliasLink.first().textContent())!.trim();

  await page.goto(`/emails/aliases/delete/${encodeURIComponent(source)}`);

  await page.getByRole('button', { name: /confirm delete/i }).click();

  await expect(page).toHaveURL(/\/emails\/aliases\/?$/, {
    timeout: 10_000,
  });

  await expect(page.getByText(source, { exact: true })).toHaveCount(0);
});

// ─── Default address ─────────────────────────────────────────────────────────

async function getFirstDefaultAddressDomain(page: Page): Promise<string | null> {
  await page.goto('/emails/default/');

  const domainSelect = page.locator('#domains');
  await expect(domainSelect).toBeVisible();

  const domainOptions = domainSelect.locator(
    'option[value]:not([value=""]):not([disabled])'
  );

  if ((await domainOptions.count()) === 0) {
    return null;
  }

  return domainOptions.first().getAttribute('value');
}

test('default address selector page loads', async ({ page }) => {
  await page.goto('/emails/default/');

  await expect(page).toHaveURL(/\/emails\/default\/?$/);

  await expect(
    page.getByRole('heading', { name: /default email address/i })
  ).toBeVisible();

  const domainSelect = page.locator('#domains');

  await expect(domainSelect).toBeVisible();
  await expect(domainSelect).toHaveValue('');
});

test('default address domain selector navigates to domain page', async ({
  page,
}) => {
  const domain = await getFirstDefaultAddressDomain(page);

  if (!domain) {
    test.skip();
    return;
  }

  const domainSelect = page.locator('#domains');

  await Promise.all([
    page.waitForURL(
      url =>
        url.pathname === `/emails/default/${domain}`,
      { timeout: 10_000 }
    ),
    domainSelect.selectOption(domain),
  ]);

  await expect(page).toHaveURL(
    new RegExp(`/emails/default/${domain.replace(/\./g, '\\.')}/?$`)
  );
});

test('default address detail page shows current config or empty state', async ({
  page,
}) => {
  const domain = await getFirstDefaultAddressDomain(page);

  if (!domain) {
    test.skip();
    return;
  }

  await page.goto(`/emails/default/${encodeURIComponent(domain)}`);

  await expect(
    page.getByRole('heading', { name: /default email address/i })
  ).toBeVisible();

  await expect(
    page.getByRole('heading', { name: /current configuration/i })
  ).toBeVisible();

  await expect(page.locator('#destination')).toBeVisible();

  await expect(
    page.getByRole('button', { name: /^save$/i })
  ).toBeVisible();

  const configuredDestination = page.locator('#current-destination');
  const emptyState = page.getByText(/no catch-all configured/i);

  await expect(
    configuredDestination.or(emptyState)
  ).toBeVisible();
});

test('default address — set and clear catch-all', async ({ page }) => {
  const domain = await getFirstDefaultAddressDomain(page);

  if (!domain) {
    test.skip();
    return;
  }

  await page.goto(`/emails/default/${encodeURIComponent(domain)}`);

  const destinationInput = page.locator('#destination');
  const saveButton = page.getByRole('button', { name: /^save$/i });

  await expect(destinationInput).toBeVisible();
  await expect(saveButton).toBeVisible();

  const destination = `catchall-pw-test@${domain}`;

  // Set catch-all.
  await destinationInput.fill(destination);
  await saveButton.click();

  /*
   * The save operation may update the page through JavaScript rather than
   * perform a full navigation, so assert the resulting UI instead of waiting
   * for the "load" event.
   */
  const currentDestination = page.locator('#current-destination');

  if (await currentDestination.count()) {
    await expect(currentDestination).toContainText(destination);
  } else {
    await expect(destinationInput).toHaveValue(destination);
  }

  // Clear catch-all.
  const deleteButton = page.locator('#delete-btn');

  if (await deleteButton.isVisible().catch(() => false)) {
    page.once('dialog', async dialog => {
      await dialog.accept();
    });

    await deleteButton.click();
  } else {
    /*
     * On the HTML shown, clearing is done by saving an empty destination.
     */
    await destinationInput.click();
    await destinationInput.press('ControlOrMeta+A');
    await destinationInput.press('Delete');
    await expect(destinationInput).toHaveValue('');
    await saveButton.click();
  }

  await expect(
    page.getByText(/no catch-all configured/i)
  ).toBeVisible({ timeout: 10_000 });

  await expect(destinationInput).toHaveValue('');
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
