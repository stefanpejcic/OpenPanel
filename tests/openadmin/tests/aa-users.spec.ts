// NOT FINISHED!
import { test, expect } from '@playwright/test';

async function navigateToUsersPage(page: any) {
  await page.goto(`/users`);
  await expect(page).toHaveURL(/users/);
}



test('view users', async ({ page }) => {
  await navigateToUsersPage(page);
  await expect(page.getByText(/create new/i)).toBeVisible();
  console.log('Users page is accessible');
});



test('create user', async ({ page }) => {
  await page.goto(`/user/new`);
  await expect(page).toHaveURL(/user\/new/);

  await page.fill('[name="admin_username"]', 'testinguser');
  await page.fill('[name="admin_password"]', 'testingpassword');
  await page.fill('[name="admin_email"]', 'stefan@test.rs');

  await page.click('#CreateUserButton');

  const successMessage = page.getByText('user created successfully');
  await expect(successMessage).toBeVisible({ timeout: 20_000 });

  console.log('User created successfully');
});



test('test autologin', async ({ page, context }) => {
  await navigateToUsersPage(page);
  await expect(page.getByText(/create new/i)).toBeVisible();

  const [newTab] = await Promise.all([
    context.waitForEvent('page'),
    page.click('a[href="/login/token/testinguser"]'),
  ]);

  await newTab.waitForLoadState();
  await expect(newTab).toHaveURL(/dashboard/);
  await expect(newTab.getByText(/last login ip address/i)).toBeVisible();

  console.log('autologin is working');
});


test('open single user', async ({ page }) => {
  await page.goto(`/users/testinguser`);
  await expect(page).toHaveURL(/users\/testinguser/);

  const expectedItems = [
    'Statistics',
    'Services',
    'Storage',
    'Overview',
    'Edit',
    'Transfer',
    'Suspend',
    'Delete',
    'Activity Log',
    'Login Log',
  ];

  for (const item of expectedItems) {
    await expect(
      page.locator('nav[aria-label="core navigation links"] a').filter({ hasText: item })
    ).toHaveCount(1);
  }

  console.log('single user page working');
});

test('test tabs', async ({ page }) => {
  await page.goto(`/users/testinguser`);
  await expect(page).toHaveURL(/users\/testinguser/);
  const nav = page.getByRole('navigation', { name: 'core navigation links' });
  
  // SERVICES
  await nav.getByText('Services').click();
  await expect(page).toHaveURL(/#services/);
  const expectedServices = ['cpu', 'ram', 'actions'];
  for (const col of expectedServices) {
    await expect(page.locator(`th[x-show="columns.${col}"]`)).toBeVisible();
  }
  console.log('services tab ok');

  // STORAGE
  await nav.getByText('Storage').click();
  await expect(page).toHaveURL(/#storage/);
  // TODO //
  console.log('storage tab ok');

  // OVERVIEW
  await nav.getByText('Overview').click();
  await expect(page).toHaveURL(/#info/);
    
  const infoPanel = page.locator('[x-show="activeTab === \'info\'"]');
  
  // Get row by matching the label span text exactly (case-insensitive)
  function getRow(panel, labelText) {
    return panel.locator('.rounded-lg').filter({
      has: panel.page().locator('span.text-sm', { hasText: labelText })
    }).first();
  }
  
  // Extract only direct text nodes from value span, ignoring child elements (badges, SVGs, imgs)
  async function extractValue(row) {
    return row.locator('span.font-medium').first().evaluate(el =>
      Array.from(el.childNodes)
        .filter(n => n.nodeType === Node.TEXT_NODE)
        .map(n => n.textContent.trim())
        .filter(Boolean)
        .join('')
        .trim()
    );
  }
  
  // For fields where the value is inside a nested badge span
  async function extractBadgeValue(row) {
    return row.locator('span.font-medium span').first().innerText().then(v => v.trim());
  }
  
  const fields = [
    { label: 'Username:',        validate: v => /^[a-z0-9_-]+$/i.test(v) },
    { label: 'Email address:',   validate: v => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v) },
    { label: 'Locale:',          validate: v => v.length > 0 },
    { label: '2FA status:',      validate: v => ['Active', 'Inactive'].includes(v), badge: true },
    { label: 'User ID (UID):',   validate: v => /^\d+$/.test(v) },
    { label: 'IP address:',      validate: v => /^(\d{1,3}\.){3}\d{1,3}$|^[0-9a-fA-F:]+$/.test(v) },
    { label: 'Geo Location:',    validate: v => /^[A-Z]{2}$/.test(v) },
    { label: 'Server:',          validate: v => v.length > 0 },
    { label: 'Docker Context:',  validate: v => v.length > 0 },
    { label: 'Home dir:',        validate: v => v.startsWith('/home/') },
    { label: 'Web server:',      validate: v => ['apache', 'nginx', 'openlitespeed', 'openresty'].includes(v.toLowerCase())},
    { label: 'Varnish Caching:', validate: v => ['Enabled', 'Disabled'].includes(v), badge: true },
    { label: 'Database type:',   validate: v => ['mariadb', 'mysql'].includes(v.toLowerCase())},
    { label: 'Setup time:',      validate: v => /^\d{2}\.\d{2}\.\d{4} \d{2}:\d{2}:\d{2}$/.test(v) },
  ];

  for (const { label, validate, badge } of fields) {
    const row = getRow(infoPanel, label);
    await expect(row).toBeVisible();
  
    const value = badge ? await extractBadgeValue(row) : await extractValue(row);
  
    if (!validate(value)) {
      throw new Error(`Field "${label}" failed validation with value: "${value}"`);
    }
    console.log(`  ✓ ${label.padEnd(20)} "${value}"`);
  }
  
  console.log('overview tab ok');

  // EDIT
  await nav.getByText('Edit').click();
  await expect(page).toHaveURL(/#edit/);
  const expectedFields = ['input[name="new_username"]', 'input[name="new_email"]', 'input[name="new_password"]', 'select[name="new_ip"]',];
  for (const selector of expectedFields) {
    await expect(page.locator(selector)).toBeVisible();
  }
  console.log('edit tab ok');

  // TRANSFER
  await nav.getByText('Transfer').click();
  await expect(page).toHaveURL(/#transfer/);
  await expect(page.locator('#server')).toBeVisible();
  await expect(page.locator('#port')).toBeVisible();
  await expect(page.locator('#username')).toBeVisible();
  await expect(page.locator('#password')).toBeVisible();
  await expect(page.locator('#live_transfer')).toBeVisible();
  console.log('transfer tab ok');

  // SUSPEND
  await nav.getByText('Suspend').click();
  await expect(page).toHaveURL(/#suspend/);
  
  await expect(page.getByText('suspend user account', { exact: false })).toBeVisible();
  await page.locator('[x-model="confirmationText"]:visible').fill('testinguser');
  await expect(page.getByRole('button', { name: /suspend account/i })).toBeVisible();
  
  console.log('suspend tab ok');

  // DELETE
  await nav.getByText('Delete').click();
  await expect(page).toHaveURL(/#delete/);
  
  await expect(page.getByText('delete user account', { exact: false })).toBeVisible();
  await page.locator('[x-model="confirmationText"]:visible').fill('testinguser');
  await expect(page.getByRole('button', { name: /delete account permanently/i })).toBeVisible();
  
  console.log('delete tab ok');

  // ACTIVITY LOG
  await nav.getByText('Activity Log').click();
  await expect(page).toHaveURL(/#activity/);
  
  const activityLink = page.locator('a[href="/json/user-activity/testinguser?raw=true"]');
  await expect(activityLink).toBeVisible();
  console.log('activity tab ok');

  // LOGIN LOG
  await nav.getByText('Login Log').click();
  await expect(page).toHaveURL(/#logins/);
  
  const loginlogLink = page.locator('a[href="/json/user-logins/testinguser?raw=true"]');
  await expect(loginlogLink).toBeVisible();
  console.log('loginlog tab ok');
});



test('search users', async ({ page }) => {
  await navigateToUsersPage(page);
  await page.locator('[x-model="searchQuery"]').fill('testinguser');
  
  const row = page.getByRole('row').filter({ hasText: 'testinguser' });
  await expect(row).toHaveCount(1);
  await expect(row).toHaveText(/testinguser/i);

  console.log('Users search is functional');
});



test('toggle columns', async ({ page }) => {
  await navigateToUsersPage(page);

  await page.getByRole('button', { name: 'Show Columns' }).click();

  const rows = page.locator('ul[aria-labelledby="dropdownToggleButton"] li');

  const count = await rows.count();

  for (let i = 0; i < count; i++) {
    const row = rows.nth(i);

    const checkbox = row.locator('input[type="checkbox"]');

    const xModel = await checkbox.getAttribute('x-model');
    if (!xModel) continue;

    const columnKey = xModel.replace('columns.', '');
    const th = page.locator(`th[x-show="columns.${columnKey}"]`);

    const initialState = await checkbox.isChecked();

    await row.locator('label').click();
    await page.waitForTimeout(100); // needed for alpine.js x-show

    const expectedStateAfterToggle = !initialState;

    if (expectedStateAfterToggle) {
      await expect(th).toBeVisible();
    } else {
      await expect(th).toBeHidden();
    }

    await row.locator('label').click();
    await page.waitForTimeout(100);

    if (initialState) {
      await expect(th).toBeVisible();
    } else {
      await expect(th).toBeHidden();
    }
  }

  console.log('column toggle is working');
});



test('delete user', async ({ page }) => {
  await navigateToUsersPage(page);

  await page.goto('/users/testinguser#delete');
  await expect(page).toHaveURL(/\/users\/testinguser#delete$/);

  await expect(page.getByText('delete user account', { exact: false })).toBeVisible();

  const confirmationInput = page.locator('[x-model="confirmationText"]:visible');

  const deleteButton = page.getByRole('button', {name: /delete account permanently/i,});

  await expect(
  page.getByRole('button', { name: /delete account permanently/i })
).toHaveCount(0); 

  //await expect(deleteButton).toBeDisabled(); nije radio ovaj check pa sam ga zamenio sa ovim iznad

  // https://github.com/stefanpejcic/OpenPanel/issues/948
  await confirmationInput.click();
  await confirmationInput.press('Enter');
  await expect(page).toHaveURL(/\/users\/testinguser#delete$/);
  await expect(page.getByText("User 'testinguser' deleted successfully", {exact: false,})).not.toBeVisible();
  console.log('submit is not working without confirmation (issue #948)');

  // valid deletion
  await confirmationInput.fill('testinguser');
  await expect(deleteButton).toBeEnabled();
  await expect(
  page.getByRole('button', { name: /delete account permanently/i })
).toHaveCount(1);
  await deleteButton.click();
  await expect(page).toHaveURL(/\/users\/?$/);
  await expect(page.getByText("User 'testinguser' deleted successfully", {exact: false,})).toBeVisible();

  const userRows = page.locator('table tbody tr');
  await expect(userRows.filter({ hasText: 'testinguser' })).toHaveCount(0);

  console.log('delete user is working');
});
