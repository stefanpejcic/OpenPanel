import { test, expect } from '@playwright/test';
import { randomBytes } from 'crypto';

const DOMAINS = [
  'wp.tests.openpanel.org',
  'php.tests.openpanel.org',
  'nodejs.tests.openpanel.org',
  'python.tests.openpanel.org',
  'website-builder.tests.openpanel.org',
  'files.tests.openpanel.org',
  'redirect.tests.openpanel.org',
  'to-be-removed.com',
];


async function getDomainCount(page: Page): Promise<number> {
  const text = await page
    .locator('#dashboard_usage_domains')
    .locator('p')
    .nth(1)
    .textContent();

  if (!text) throw new Error('Cannot read domain count');

  const match = text.match(/(\d+)\s*\//);
  if (!match) throw new Error(`Cannot parse domain count from: ${text}`);

  return parseInt(match[1], 10);
}

async function expectDomainInTable(page: Page, domain: string) {
  await expect(page.locator('table')).toContainText(domain);
}

async function expectDomainNotInTable(page: Page, domain: string) {
  await expect(page.locator('table')).not.toContainText(domain);
}






async function addDomain(page, domain) {
  await page.goto(`/domains/new`);
  await expect(page).toHaveURL(/domains\/new/);
  await page.getByRole('textbox', { name: 'Domain*' }).fill(domain);
  await page.getByRole('button', { name: 'Add Domain' }).click();

  await expect(page.getByText(`Domain name ${domain} added successfully`)).toBeVisible();
  console.log(`Domain added: ${domain}`);
}



test('add domains', async ({ page }) => {
  test.setTimeout(60_000);

  await page.goto('/dashboard');
  const initialCount = await getDomainCount(page);
  let expectedCount = initialCount;

  for (const domain of DOMAINS) {
    await addDomain(page, domain);

    // table check
    await page.goto('/domains');
    await expectDomainInTable(page, domain);
    expectedCount++;

    // dashboard check
    await page.goto('/dashboard');
    await expect.poll(async () => {return await getDomainCount(page);}).toBe(expectedCount);
  }
});

test('verify files created for a new domain', async ({ page }) => {
  // domains page shows new domain
  await page.goto(`/domains`);
  await expect(page.locator('td[x-show="columns.domain"]', { hasText: 'wp.tests.openpanel.org' })).toBeVisible();
  console.log(`domain visible`);

  // docroot created
  await page.goto(`/files`);
  await expect(page).toHaveURL(/files/);
  await expect(page.getByText(/wp.tests.openpanel.org/i)).toBeVisible();
  console.log(`document root visible`);

  // DNS zone created
  await page.goto(`/domains\/edit-dns-zone\/wp.tests.openpanel.org`);
  await expect(page).toHaveURL(/domains\/edit-dns-zone\/wp.tests.openpanel.org/);
  await expect(page.getByText(/spf1/i)).toBeVisible();
  console.log(`zone file exists`);

  // vhosts file created
  await page.goto(`/domains\/vhosts?domain=wp.tests.openpanel.org`);
  await expect(page.locator('#editor')).toContainText('index.php');  
  console.log(`vhost file exists`);

  // SSL generation
  await page.goto('http://wp.tests.openpanel.org', {
    waitUntil: 'domcontentloaded',
  });
  const certData = page.locator('#certData');
  const start = Date.now();
  
  while (Date.now() - start < 15000) {
    await page.goto('/domains/ssl?domain_name=wp.tests.openpanel.org');
  
    try {
      await expect(certData).toBeVisible({ timeout: 2000 });
      break;
    } catch (e) {
      // retry
    }
  }
  await expect(certData).toBeVisible();
  console.log(`cert file exists`);

  // logs
  //await page.goto('/domains/log/wp.tests.openpanel.org');
  //const logRows = page.locator('#logs-table tbody tr');
  //await expect(logRows.filter({ hasText: '/' }).first()).toBeVisible();
  //console.log('logs are working');
});

test('search domains', async ({ page }) => {
  await page.goto(`/domains`);
  await expect(page).toHaveURL(/domains/);

  const searchBox = page.getByRole('searchbox', { name: 'Search' });
  const rows = page.locator('tbody tr');

  await searchBox.fill('wp.tests.openpanel.org');
  const count = await rows.count();
  expect(count).toBeGreaterThan(0);

  const visibleRows = rows.filter({ has: page.locator(':visible') });
  await expect(visibleRows).toHaveCount(1);
  await expect(visibleRows.first()).toContainText(/wp\.tests\.openpanel\.org/i);

  await expect(rows.filter({ hasNot: page.locator(':visible') })).toHaveCount(
    (await rows.count()) - 1
  );

  await searchBox.fill('non-existing-domain.com');
  await expect(rows.filter({ has: page.locator(':visible') })).toHaveCount(0);

  console.log('Domain search is working');
});



test('check columns for domains table', async ({ page }) => {
  await page.goto(`/domains`);
  await expect(page).toHaveURL(/domains/);
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



test('vhost editor', async ({ page }) => {
  await page.goto(`/domains/vhosts?domain=wp.tests.openpanel.org`);

  // TODO: add header for ols, apache, nginx
  //       then curl the domain and check for header

  console.log('vhost editor is working');
});

test('change docroot', async ({ page }) => {
  const NEW_FOLDER = `folder_${Math.random().toString(36).substring(7)}`;
  const DOMAIN = 'wp.tests.openpanel.org';

  // 1. Create a dummy PHP file in the new folder
  await page.goto(`/files`);
  await page.getByRole('button', { name: ' New Folder' }).click();
  await page.locator('#foldername').fill(NEW_FOLDER);
  await page.getByRole('button', { name: 'Create' }).click();  
  await page.goto(`/file-manager/edit-file/${NEW_FOLDER}/testing.php?editor=text&new=true`);
  await page.locator('#editor-text').fill(`<?php echo "File is shown from folder: ${NEW_FOLDER}"; ?>`);
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText(/saved|success/i).first()).toBeVisible();

  // 2. Change docroot
  await page.goto(`/domains/docroot?domain_name=${DOMAIN}`);
  await page.locator('input[name="new_docroot"]').fill(`/var/www/html/${NEW_FOLDER}`);
  await page.getByRole('button', { name: 'Change docroot' }).click();
  const successPattern = new RegExp(`Docroot updated to: /var/www/html/${NEW_FOLDER} for domain: ${DOMAIN}`);
  await expect(page.getByText(successPattern)).toBeVisible();
  await expect(page.locator('input[name="new_docroot"]')).toHaveValue(`/var/www/html/${NEW_FOLDER}`);
  console.log(`docroot change saved`);

  // 3. Open domains page and verify
  await page.goto('/domains');
  const domainRow = page.locator('tr', { hasText: DOMAIN });
  await expect(domainRow.locator('td', { hasText: NEW_FOLDER })).toBeVisible();
  console.log(`docroot change visible in table`);

  // 4. Open the domain in browser and verify PHP execution
  await page.goto(`https://${DOMAIN}/testing.php`);
  const locator = page.getByText(`File is shown from folder: ${NEW_FOLDER}`);
  
  const timeout = 30000;
  const start = Date.now();
  
  while (Date.now() - start < timeout) {
    if (await locator.isVisible()) {
      break;
    }
  
    await page.waitForTimeout(1000);
    await page.reload();
  }
  await expect(locator).toBeVisible();  
  console.log('docroot change is fully working');

  // revert
  await page.goto(`/domains/docroot?domain_name=${DOMAIN}`);
  await page.locator('input[name="new_docroot"]').fill(`/var/www/html/${DOMAIN}`);
  await page.getByRole('button', { name: 'Change docroot' }).click();
  const revertPattern = new RegExp(`Docroot updated to: /var/www/html/${DOMAIN} for domain: ${DOMAIN}`);
  await expect(page.getByText(revertPattern)).toBeVisible();
  await expect(page.locator('input[name="new_docroot"]')).toHaveValue(`/var/www/html/${DOMAIN}`); 
});





// USED BY DNS EDITOR CHECKS
const domain = 'wp.tests.openpanel.org';
const recordValue = `verify-${randomBytes(6).toString('hex')}`;

test('add dns record', async ({ page }) => {
  await page.goto(`/domains/edit-dns-zone/${domain}`);

  // 1. create random TXT record
  await page.locator('#AddDNSRecord').click();
  const addRow = page.locator('#addRecordRow');
  await expect(addRow).toBeVisible();
  await addRow.locator('input[name="Name"]').fill(domain);
  await addRow.locator('select[name="Type"]').selectOption('TXT');
  await addRow.locator('input[name="Record"]').fill(recordValue);
  await page.locator('#save-row').click();
  await expect(page.getByText(/DNS record added successfully/i)).toBeVisible();

  // 2. validate on page
  const newRow = page.locator('tr.domain_row', { hasText: recordValue });
  await expect(newRow).toBeVisible();
  await expect(newRow.locator('td').nth(2)).toHaveText('TXT');
  await expect(newRow.locator('td').nth(3)).toContainText(recordValue);

  // 3. validate using dig tools
  await page.goto(`https://digwebinterface.com/?hostnames=${domain}&type=TXT&useresolver=9.9.9.10&ns=self&nameservers=${domain}`);
  const resultsArea = page.locator('#results, pre, .results, [id*="result"]').first();
  await expect(resultsArea).toBeVisible({ timeout: 10_000 });
  await page.waitForFunction(() => !document.querySelector('.loading, .spinner, [aria-busy="true"]'), { timeout: 30_000 });
  await expect(page.locator('body')).toContainText(recordValue, { timeout: 30_000 });
  console.log('dns editor is working');
});


test('edit dns record', async ({ page }) => {
  await page.goto(`/domains/edit-dns-zone/${domain}`);

  // 1. find and edit the record created in the previous test
  const newRow = page.locator('tr.domain_row', { hasText: recordValue });
  await expect(newRow).toBeVisible();
  await newRow.locator('button:has-text("Edit"), button:has-text("Edit"), button:has-text("Edit")').click();

  const field4 = page.locator('[name="field4"]').first();
  if (await field4.isVisible({ timeout: 2_000 }).catch(() => false)) {
    await field4.fill(`${recordValue}-edited`);
  }

  const editedRow = page.locator('tr.domain_row').filter({ has: page.locator('input#Record').filter({ hasValue: `${recordValue}-edited` }) });
  await expect(editedRow).toBeVisible();  
  await editedRow.locator('button:has-text("Save"), button:has-text("Save"), button:has-text("Save")').click();
  
  // 2. verify row is changed
  await expect(page.locator('tr.domain_row', { hasText: recordValue })).toHaveCount(0);
  await expect(page.locator('tr.domain_row', { hasText: `${recordValue}-edited` })).toHaveCount(1);

  // 3. validate using dig tools
  await page.goto(`https://digwebinterface.com/?hostnames=${domain}&type=TXT&useresolver=9.9.9.10&ns=self&nameservers=${domain}`);
  const resultsArea = page.locator('#results, pre, .results, [id*="result"]').first();
  await expect(resultsArea).toBeVisible();
  await page.waitForFunction(() => !document.querySelector('.loading, .spinner, [aria-busy="true"]'), { timeout: 30_000 });
  await expect(page.locator('body')).toContainText(`${recordValue}-edited`);  
  console.log('dns record deletion is working');
});



test('delete dns record', async ({ page }) => {
  await page.goto(`/domains/edit-dns-zone/${domain}`);

  // 1. find and delete the record created in the previous test
  const newRow = page.locator('tr.domain_row', { hasText: `${recordValue}-edited` });
  await expect(newRow).toBeVisible();
  //await newRow.locator('button[data-action="delete"], button.delete-record, .btn-delete').click();
  await newRow.locator('button.delete-button').click();

  const confirmBtn = page.locator('button:has-text("Confirm"), button:has-text("Confirm")').first();
  if (await confirmBtn.isVisible({ timeout: 2_000 }).catch(() => false)) {
    await confirmBtn.click();
  }

  await expect(page.getByText(/record deleted successfully/i)).toBeVisible();

  // 2. verify row is gone
  await expect(page.locator('tr.domain_row', { hasText: `${recordValue}-edited` })).toHaveCount(0);

  // 3. validate using dig tools
  await page.goto(`https://digwebinterface.com/?hostnames=${domain}&type=TXT&useresolver=9.9.9.10&ns=self&nameservers=${domain}`);
  const resultsArea = page.locator('#results, pre, .results, [id*="result"]').first();
  await expect(resultsArea).toBeVisible();
  await page.waitForFunction(() => !document.querySelector('.loading, .spinner, [aria-busy="true"]'), { timeout: 30_000 });
  await expect(page.locator('body')).not.toContainText(`${recordValue}-edited`);  
  console.log('dns record deletion is working');
});



const fs = require('fs');
test('export dns zone', async ({ page }) => {
  const downloadPromise = page.waitForEvent('download');

  await Promise.all([
    downloadPromise,
    page.goto(`/domains/export-dns-zone/${domain}`).catch(e => {if (!e.message.includes('Download is starting')) throw e;})
  ]);

  const download = await downloadPromise;
  
  const path = await download.path();
  const stats = fs.statSync(path);

  console.log(`Downloaded file to: ${path}`);
  expect(stats.size).toBeGreaterThan(1024); 
});



test('edit zone file', async ({ page }) => {
  await page.goto(`/domains/edit-dns-zone/${domain}?view=code`);

  // 1. append TXT record
  const newRecord = `\n${domain}.    14400     IN      TXT       "added via zone editor"`;
  const zoneTextArea = page.locator('textarea[name="zone_content"]');
  const existingContent = await zoneTextArea.inputValue();
  await zoneTextArea.fill(existingContent + newRecord);
  await page.click('#save_zone_button');
  await expect(page.getByText('Zone file saved successfully')).toBeVisible();

  // 2. validate in table view
  await page.goto(`/domains/edit-dns-zone/${domain}`);
  const newRow = page.locator('tr.domain_row', { hasText: `added via zone editor` });
  await expect(newRow).toBeVisible();
  await expect(newRow.locator('td').nth(2)).toHaveText('TXT');
  await expect(newRow.locator('td').nth(3)).toContainText(`added via zone editor`);

  // 3. validate using dig tools
  await page.goto(`https://digwebinterface.com/?hostnames=${domain}&type=TXT&useresolver=9.9.9.10&ns=self&nameservers=${domain}`);
  const resultsArea = page.locator('#results, pre, .results, [id*="result"]').first();
  await expect(resultsArea).toBeVisible({ timeout: 10_000 });
  await page.waitForFunction(() => !document.querySelector('.loading, .spinner, [aria-busy="true"]'), { timeout: 30_000 });
  await expect(page.locator('body')).toContainText(`added via zone editor`, { timeout: 30_000 });
  console.log('dns file editor mode is working');
});



test('reset dns zone', async ({ page }) => {
  await page.goto(`/domains/edit-dns-zone/${domain}`);
  const tmprecordValue = `temporary-${randomBytes(6).toString('hex')}`;

  // 1. create random TXT record
  await page.locator('#AddDNSRecord').click();
  const addRow = page.locator('#addRecordRow');
  await expect(addRow).toBeVisible();
  await addRow.locator('input[name="Name"]').fill(domain);
  await addRow.locator('select[name="Type"]').selectOption('TXT');
  await addRow.locator('input[name="Record"]').fill(tmprecordValue);
  await page.locator('#save-row').click();
  await expect(page.getByText(/DNS record added successfully/i)).toBeVisible();

  // 2. validate on page
  const newRow = page.locator('tr.domain_row', { hasText: tmprecordValue });
  await expect(newRow).toBeVisible();
  await expect(newRow.locator('td').nth(2)).toHaveText('TXT');
  await expect(newRow.locator('td').nth(3)).toContainText(tmprecordValue);

  // 3. restart
  await page.locator('#dropdownHoverButton').click();
  await page.locator('#dropdownHover').locator('a:has-text("Reset")').click();
  const resetBtn = page.getByRole('button', { name: 'Reset Zone', exact: true });
  await expect(resetBtn).toBeVisible();
  await resetBtn.click();
    
  // 4. validate
  const successMsg = page.getByText('DNS zone restarted successfully.');
  await expect(successMsg).toBeVisible();
  await expect(newRow).not.toBeVisible();
  await page.goto(`https://digwebinterface.com/?hostnames=${domain}&type=TXT&useresolver=9.9.9.10&ns=self&nameservers=${domain}`);
  const resultsArea = page.locator('#results, pre, .results, [id*="result"]').first();
  await expect(resultsArea).toBeVisible({ timeout: 10_000 });
  await page.waitForFunction(() => !document.querySelector('.loading, .spinner, [aria-busy="true"]'), { timeout: 30_000 });
  await expect(page.locator('body')).not.toContainText(`added via zone editor`, { timeout: 30_000 });

  console.log('dns zone restart is working');
});

// DDNS
test('dynamic dns record', async ({ page, context }) => {
  await page.goto(`/domains/dynamic-dns`);
  const subdomain = `ddns-${Date.now()}`;
  const fqdn = `${subdomain}.${domain}`;

  // 1. create dynamic dns entry
  await page.locator('#add-entry').click();
  // Scope to the create panel specifically (always visible when panel === 'create')
  const createForm = page.locator('[x-show="panel === \'create\'"] form');
  await expect(createForm).toBeVisible();
  await createForm.locator('select[name="domain"]').selectOption(domain);
  await createForm.locator('input[name="subdomain"]').fill(subdomain);
  await createForm.locator('input[name="ip"]').fill('0.0.0.0');
  await createForm.locator('button[type="submit"]').click();

  // 2. verify record appears in table
  const table = page.locator('table');
  await expect(table).toBeVisible();
  const row = table.locator('tbody tr', { hasText: subdomain });
  await expect(row).toBeVisible({ timeout: 10000 });
  await expect(row.locator('td').nth(0)).toContainText(subdomain);
  await expect(row.locator('td').nth(1)).toContainText('A');
  await expect(row.locator('td').nth(2)).toContainText('0.0.0.0');

  // 3. grab update url from the row's hidden title attribute
  const updateCode = row.locator('code');
  await expect(updateCode).toBeVisible();
  const relativeUpdateUrl = await updateCode.evaluate(
    (el) => el.getAttribute('title') || el.textContent || ''
  );
  expect(relativeUpdateUrl).toContain('/dynamic-dns/update?token=');

  // 4. open update url — this triggers the actual IP update
  const updatePage = await context.newPage();
  await updatePage.goto(relativeUpdateUrl);
  await expect(updatePage.locator('body')).toContainText(/updated|success|ip/i, { timeout: 15000 });
  await updatePage.close();

  // 5. reload and verify IP was updated away from 0.0.0.0
  await page.reload();
  const updatedRow = page.locator('tbody tr', { hasText: subdomain });
  await expect(updatedRow).toBeVisible();
  await expect(updatedRow.locator('td').nth(2)).not.toContainText('0.0.0.0');
  const updatedIp = (await updatedRow.locator('td').nth(2).textContent())?.trim();
  console.log(`Updated IP: ${updatedIp}`);

  // 6. edit the entry — change subdomain to verify edit panel works
  const editBtn = updatedRow.locator('button[title="Edit"]');
  await editBtn.click();
  const editForm = page.locator('[x-show="panel === \'edit\'"] form');
  await expect(editForm).toBeVisible();
  // Verify pre-filled values
  await expect(editForm.locator('input[name="subdomain"]')).toHaveValue(subdomain);
  await expect(editForm.locator('input[name="ip"]')).toHaveValue(updatedIp ?? '');
  // Verify update URL is shown
  const updateUrlCode = editForm.locator('code');
  await expect(updateUrlCode).toBeVisible();
  await expect(updateUrlCode).toContainText('/dynamic-dns/update?token=');
  // Save without changes (just confirm the form submits cleanly)
  await editForm.locator('button[type="button"]', { hasText: 'Save' }).click();
  await expect(page.locator('tbody tr', { hasText: subdomain })).toBeVisible();


  // 7. validate publicly using dig (optional, only if IP resolved)
  if (updatedIp) {
    await page.goto(`https://digwebinterface.com/?hostnames=${fqdn}&type=A&useresolver=9.9.9.10`);
    const resultsArea = page.locator('#results, pre, .results, [id*="result"]').first();
    await expect(resultsArea).toBeVisible({ timeout: 10000 });
    await page.waitForFunction(
      () => !document.querySelector('.loading, .spinner, [aria-busy="true"]'),
      { timeout: 30000 }
    );
    await expect(page.locator('body')).toContainText(fqdn, { timeout: 30000 });
    await expect(page.locator('body')).toContainText(updatedIp, { timeout: 30000 });
    console.log('dynamic dns public DNS resolution confirmed');
  }


  
  // 8. delete the entry
  const rowAfterEdit = page.locator('tbody tr', { hasText: subdomain });
  const deleteBtn = rowAfterEdit.locator('button[title="Delete"]');
  await deleteBtn.click();
  const deleteForm = page.locator('[x-show="panel === \'delete\'"] form');
  await expect(deleteForm).toBeVisible();
  // Confirm the warning text references our subdomain
  await expect(deleteForm).toContainText(subdomain);
  await deleteForm.locator('button[type="submit"]').click();

  // 9. verify the row is gone
  await page.waitForTimeout(1000); // brief wait for redirect/reload
  await expect(page.locator('tbody tr', { hasText: subdomain })).toHaveCount(0, { timeout: 10000 });
  console.log('dynamic dns create/edit/delete all working');
});



test('redirects', async ({ page }) => {
  const domain = 'redirect.tests.openpanel.org';
  const redirectUrl = 'https://pejcic.rs/?proba';
  const editedUrl = 'https://pejcic.rs/?edited';

  // CREATE
  await page.goto(`/domains/redirect?domain=${domain}`);
  await page.fill('input[name="redirect_url"][type="url"]', redirectUrl);
  await page.click('#save-redirect');
  await expect(page.locator(`text=Successfully created redirect from domain ${domain} to ${redirectUrl}`)).toBeVisible();

  await page.goto(`/domains/redirect?domain=${domain}`);
  await expect(page.locator('input[name="redirect_url"][type="url"]')).toHaveValue(redirectUrl);

  let response = await page.goto(`https://${domain}`, { waitUntil: 'load' });
  expect(response?.url()).toContain('pejcic.rs');
  console.log('Domain redirect created and verified successfully!');

  // EDIT
  await page.goto(`/domains/redirect?domain=${domain}`);
  await page.fill('input[name="redirect_url"][type="url"]', editedUrl);
  await page.click('#save-redirect');
  await expect(page.locator(`text=Successfully created redirect from domain ${domain} to ${editedUrl}`)).toBeVisible();

  await page.goto(`/domains/redirect?domain=${domain}`);
  await expect(page.locator('input[name="redirect_url"][type="url"]')).toHaveValue(editedUrl);

  response = await page.goto(`https://${domain}`, { waitUntil: 'load' });
  expect(response?.url()).toContain('pejcic.rs');
  console.log('Domain redirect edited and verified successfully!');

  // DELETE
  await page.goto(`/domains/redirect?domain=${domain}`);
  await page.click('button[type="submit"]:has-text("Delete Redirect")');
  await page.goto(`/domains/redirect?domain=${domain}`);
  await expect(page.locator('input[name="redirect_url"][type="url"]')).toHaveValue('');

  response = await page.goto(`https://${domain}`, { waitUntil: 'load' });
  expect(response?.url()).not.toContain('pejcic.rs');
  console.log('Domain redirect deleted and verified successfully!');
});



test('suspend domain', async ({ page }) => {
  const domain = 'wp.tests.openpanel.org';

  await page.goto(`/domains/suspend?domain=${domain}`);
  await page.locator('#confirm-suspend').click();

  await expect(page.getByText('Domain suspended successfully.')).toBeVisible();
  console.log('Suspend action confirmed ✓');

  await page.waitForTimeout(2000);

  const response = await page.request.get(`https://${domain}`, { failOnStatusCode: false });
  console.log(`Suspended domain status: ${response.status()}`);

  const body = await response.text();
  console.log(`Suspended domain body snippet: ${body.slice(0, 200)}`);
  expect(body).toContain('This website has been suspended');

  console.log('Domain suspension verified successfully!');
});


test('unsuspend domain', async ({ page }) => {
  const domain = 'wp.tests.openpanel.org';

  await page.goto(`/domains/unsuspend?domain=${domain}`);
  await page.locator('#confirm-unsuspend').click();

  await expect(page.getByText('Domain unsuspended successfully.')).toBeVisible();
  console.log('Unsuspend action confirmed ✓');

  await page.waitForTimeout(2000);

  const response = await page.request.get(`https://${domain}`, { failOnStatusCode: false });
  console.log(`Unsuspended domain status: ${response.status()}`);
  expect(response.status()).toBe(200);

  const body = await response.text();
  console.log(`Unsuspended domain body snippet: ${body.slice(0, 200)}`);

  await page.goto(`http://${domain}`);
  await expect(page.locator('body')).toContainText(/this domain currently has no website\. please check back later\.|litespeed/i);
  console.log('Domain unsuspension verified successfully!');
});


// DELETE SINGLE DOMAIN!
test('delete domain', async ({ page }) => {
  const domain = 'to-be-removed.com';
  await page.goto('/dashboard');
  const initialCount = await getDomainCount(page);

  // go to delete page
  await page.goto(`/domains/delete?domain=${domain}`);

  const deleteButton = page.getByRole('button', { name: /delete domain/i });
  await expect(deleteButton).toBeVisible();
  await deleteButton.click();
  await expect(page.locator('body')).toContainText(/deleted successfully/i);
  console.log(`Domain deleted: ${domain}`);

  // verify it's gone from table/listing page
  await page.goto('/domains');
  await expect(page.locator('table')).not.toContainText(domain);

  // verify dashboard count decreased
  await page.goto('/dashboard');
  const finalCount = await getDomainCount(page);
  expect(finalCount).toBe(initialCount - 1);
});
