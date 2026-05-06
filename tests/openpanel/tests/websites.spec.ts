import { test, expect } from '@playwright/test';

test('auto-installer page has install links', async ({ page }) => {
  await page.goto('/auto-installer');

  const expectedHrefs = [
    '/wordpress/install',
    '/website-builder/install',
    '/pm2/install#python',
    '/pm2/install#node',
  ];

  for (const href of expectedHrefs) {
    const links = page.locator(`a[href="${href}"]`);
    await expect(links.first()).toBeVisible();
  }

  console.log('all links present');  
});


test('wordpress install', async ({ page }) => {
  await page.goto('/wordpress/install');
  await page.fill('input[name="website_name"]', 'My Site');
  await page.fill('input[name="site_description"]', 'another site testing');
  await page.fill('input[name="admin_username"]', 'admin');
  await page.fill('input[name="admin_password"]', 'b67sf97sfs3sedf45');
  await page.locator('#domain_id').selectOption('wp.tests.openpanel.org');
  await page.locator('#wordpress_version').selectOption('6.9.4'); // TODO: randomize!
  await page.locator('#installButton').click();

  await expect(page.locator('text=WordPress installation complete!')).toBeVisible({ timeout: 30000 });

  await page.goto('http://wp.tests.openpanel.org');
  await expect(page.locator('body')).toContainText('Hello world!');
  console.log('wordpress install is working');
});



test('wordpress links', async ({ page }) => {

  // 3. test links on /wordpress
  console.log('skipped');
});



// TODO: cover wp-manager tabs: security, updates, debugging, backup, clone, remove

test('wp manager data', async ({ page }) => {
  await page.goto('/website?domain=wp.tests.openpanel.org');

  // WP version (e.g. 6.5.2)
  const wpVersion = await page.locator('#wp-version').textContent();      
  expect(wpVersion).toMatch(/\b\d+\.\d+(\.\d+)?\b/);

  // PHP version (e.g. 8.1.12)
  const phpVersion = await page.locator('#php-version').textContent();     
  expect(phpVersion).toMatch(/\b\d+\.\d+(\.\d+)?\b/);

  // MySQL / MariaDB version (e.g. 10.6.12-MariaDB or 8.0.36)
  const mysqlVersion = await page.locator('#mysql-version').textContent(); 
  expect(mysqlVersion).toMatch(/\b(\d+\.\d+(\.\d+)?)([-\w]*)\b/i);

  // Created date
  const createdDate = await page.locator('#created_date').textContent();   
  expect(createdDate?.trim().length).toBeGreaterThan(0);

  // files size (e.g. 83M, 1.2 GB, 512KB)
  const filesSize = await page.locator('#filesSize').textContent();        
  expect(filesSize).toMatch(/\b\d+(\.\d+)?\s?(K|M|G|T)?B?\b/i);

  // db size (e.g. 0.78 MB)
  const databaseSize = await page.locator('#databaseSize').textContent();
  expect(databaseSize).toMatch(/\b\d+(\.\d+)?\s?(K|M|G|T)?B\b/i);

  // db logins
  await expect(page.locator('#database-host')).toHaveText(/mysql|mariadb/i);
  const selectors = [
    '#database-table-prefix',
    '#database-password',
    '#database-name',
    '#wp_cache_type'
  ];
  
  for (const selector of selectors) {
    const text = await page.locator(selector).textContent();
    expect(text?.trim().length).toBeGreaterThan(0);
  }

  console.log('wp manager options are functional');
});



test('live preview', async ({ page }) => {
  await page.goto('/website?domain=wp.tests.openpanel.org');

  const popupPromise = page.waitForEvent('popup');
  await page.locator('button[onclick="sendDataToPreview(event)"]').click();

  const previewPage = await popupPromise;
  await previewPage.waitForLoadState();
  await expect(previewPage.locator('body')).toContainText('Hello world!');

  console.log('live preview is working');
});



test('wp-admin autologin', async ({ page }) => {
  await page.goto('/website?domain=wp.tests.openpanel.org');

  const popupPromise = page.waitForEvent('popup');
  await page.locator('#login_button_text').click();
  await expect(page.locator('body')).toContainText('Generating auto-login link');

  const previewPage = await popupPromise;
  await previewPage.waitForLoadState();
  await expect(previewPage.locator('body')).toContainText('admin');

  console.log('wp-admin login is working');
});



test('general options', async ({ page }) => {
  await page.goto('/website?domain=wp.tests.openpanel.org');
  await page.locator('#settings-tab').click();
  await expect(page.getByText('Checking options from WP-CLI')).toBeVisible();
  await expect(page.getByText('WordPress Options loaded')).toBeVisible();

  // new values
  const newBlogName = 'Test Blog Name';
  const newBlogDescription = 'Test Blog Description';
  const newAdminEmail = 'test@example.com';

  // Edit text inputs
  await page.locator('#blogname').clear();
  await page.locator('#blogname').fill(newBlogName);

  await page.locator('#blogdescription').clear();
  await page.locator('#blogdescription').fill(newBlogDescription);

  await page.locator('#admin_email').clear();
  await page.locator('#admin_email').fill(newAdminEmail);

  // get current, then click to change
  const usersCanRegister = page.locator('#users_can_register');
  const usersCanRegisterChecked = await usersCanRegister.isChecked();
  await usersCanRegister.click();
  const expectedUsersCanRegister = !usersCanRegisterChecked;

  const blogPublic = page.locator('#blog_public');
  const blogPublicChecked = await blogPublic.isChecked();
  await blogPublic.click();
  const expectedBlogPublic = !blogPublicChecked;

  const defaultPingStatus = page.locator('#default_ping_status');
  const defaultPingStatusChecked = await defaultPingStatus.isChecked();
  await defaultPingStatus.click();
  const expectedDefaultPingStatus = !defaultPingStatusChecked;

  // Save
  await page.locator('#saveGeneralBtn').click();
  await expect(page.getByText('Saving general settings')).toBeVisible();
  await expect(page.getByText('General options edited successfully')).toBeVisible();

  // Reload and re-navigate to settings
  await page.goto('/website?domain=wp.tests.openpanel.org');
  await page.locator('#settings-tab').click();
  await expect(page.getByText('Checking options from WP-CLI')).toBeVisible();
  await expect(page.getByText('WordPress Options loaded')).toBeVisible();

  // Verify updated text values
  await expect(page.locator('#blogname')).toHaveValue(newBlogName);
  await expect(page.locator('#blogdescription')).toHaveValue(newBlogDescription);
  await expect(page.locator('#admin_email')).toHaveValue(newAdminEmail);

  // Verify checkbox states persisted
  await expect(page.locator('#users_can_register')).toBeChecked({ checked: expectedUsersCanRegister });
  await expect(page.locator('#blog_public')).toBeChecked({ checked: expectedBlogPublic });
  await expect(page.locator('#default_ping_status')).toBeChecked({ checked: expectedDefaultPingStatus });

  console.log('general options are working');
});



test('maintenance mode', async ({ page }) => {
  await page.goto('/website?domain=wp.tests.openpanel.org');
  await page.locator('#maintenance-tab').click();

  // Wait for the status span to appear and get its text
  await page.waitForSelector('#current_maintenance_mode_status span');
  const statusText = await page.locator('#current_maintenance_mode_status span').innerText();
  const wasEnabled = statusText.trim().toLowerCase() === 'enabled';

  // --- Round 1: toggle to the opposite state and verify ---
  await page.locator('#runMaintenance_action').click();

  if (wasEnabled) {
    await expect(page.getByText('Maintenance mode is now disabled.')).toBeVisible();
  } else {
    await expect(page.getByText('Maintenance mode is now enabled.')).toBeVisible();
  }

  await page.waitForTimeout(5000);

  const randomUri = `https://wp.tests.openpanel.org?${Math.random().toString(36).slice(2)}`;
  await page.goto(randomUri);

  if (wasEnabled) {
    await expect(page.locator('body')).toContainText('Hello world!');
  } else {
    await expect(page.locator('body')).toContainText('Briefly unavailable for scheduled maintenance.');
  }

  // --- Round 2: toggle back to original state and verify ---
  await page.goto('/website?domain=wp.tests.openpanel.org');
  await page.locator('#maintenance-tab').click();
  await page.locator('#runMaintenance_action').click();

  if (wasEnabled) {
    await expect(page.getByText('Maintenance mode is now enabled.')).toBeVisible();
  } else {
    await expect(page.getByText('Maintenance mode is now disabled.')).toBeVisible();
  }

  await page.waitForTimeout(5000);

  const randomUri2 = `https://wp.tests.openpanel.org?${Math.random().toString(36).slice(2)}`;
  await page.goto(randomUri2);

  if (wasEnabled) {
    await expect(page.locator('body')).toContainText('Briefly unavailable for scheduled maintenance.');
  } else {
    await expect(page.locator('body')).toContainText('Hello world!');
  }

  console.log('maintenance mode on/off is working');
});



test('cache flush', async ({ page }) => {
  await page.goto('/website?domain=wp.tests.openpanel.org');

  await page.locator('button[click="flushCache"]').click();
  const message = await page.getByText(/Cache flushed successfully/).innerText();  
  console.log('flush wp cache is working');
});



test('live visitors count', async ({ page }) => {
  await page.goto('/website?domain=wp.tests.openpanel.org');

  // 1. Get initial value
  const initialText = await page.locator('#visitors_data').innerText();
  const initialValue = parseInt(initialText, 10);

  // 2. Trigger a real request
  await page.request.get('https://wp.tests.openpanel.org');

  // 3. Reload page
  await page.goto('/website?domain=wp.tests.openpanel.org');

  // 4. Wait up to 60s for value to change
  await page.waitForFunction(
    (selector, oldValue) => {
      const el = document.querySelector(selector);
      if (!el) return false;

      const current = parseInt(el.textContent || '0', 10);
      return current !== oldValue;
    },
    '#visitors_data',
    initialValue,
    { timeout: 60_000 }
  );

  console.log('live visitor counter is working');
});



test('waf on/off', async ({ page }) => {
  await page.goto('/website?domain=wp.tests.openpanel.org');

  await page.locator('#waf_toggle_btn').click();
  const message = await page.getByText(/Firewall for wp\.tests\.openpanel\.org is now (On|Off)/).innerText();
  const isOn = message.includes('On');

  await page.waitForTimeout(5000);

  const response = await page.request.get('https://wp.tests.openpanel.org');
  
  if (isOn) {
    // WAF ON
    expect([403, 406, 409, 422]).toContain(response.status());
  } else {
    // WAF OFF
    expect(response.status()).toBe(200);
  }
  console.log('wp manager firewall on/off is working as expected!');
});



test('wp remove', async ({ page }) => {

  // 5. test remove
  await page.goto('/website?domain=website-builder.tests.openpanel.org');
  await page.locator('a#remove-tab').click();
  await page.locator('button#delete-site').click();
  await page.locator('button#confirm-delete-site').click();
  await expect(page.locator('text=Website deleted successfully!')).toBeVisible({ timeout: 30000 });
  await page.goto('/sites');
  await expect(page.locator('tr#site-row-website-builder.tests.openpanel.org')).not.toBeVisible();
  console.log('website uninstall is working');

  // 6. install again and test detach
  // TODO: remove files

});




test('website builder', async ({ page }) => {
  // 1. install
  await page.goto('/website-builder/install');
  await page.locator('#domain_id').selectOption('website-builder.tests.openpanel.org');
  await page.locator('#installButton').click();
  await expect(page.locator('text=Website creation completed!')).toBeVisible({ timeout: 60000 });
  await expect(page).toHaveURL(/\/website-builder\/edit\?domain=website-builder\.tests\.openpanel\.org\/.+/);

  // 2. test edit and save
  await page.locator('span.gjs-pn-btn.fa.fa-save').click();
  await expect(page.locator('text=Saved successfully!')).toBeVisible({ timeout: 30000 });
  await page.goto('http://website-builder.tests.openpanel.org/');
  await expect(page.locator('body')).toContainText('tailwindcss');

  // 3. test view
  await page.goto('/sites');
  const table = page.locator('tbody.divide-y.divide-gray-200.dark\\:divide-gray-800');
  await expect(table).toBeVisible();
  await expect(page.locator('tr#site-row-website-builder.tests.openpanel.org')).toBeVisible();
  console.log('website install is working');
  await expect(page.locator('a[href="/website-builder/edit?domain=website-builder.tests.openpanel.org"]')).toBeVisible();
  console.log('website edit is working');

  // 4. test remove
  await page.goto('/website?domain=website-builder.tests.openpanel.org');
  await page.locator('a#remove-tab').click();
  await page.locator('button#delete-site').click();
  await page.locator('button#confirm-delete-site').click();
  await expect(page.locator('text=Website deleted successfully!')).toBeVisible({ timeout: 30000 });
  await page.goto('/sites');
  await expect(page.locator('tr#site-row-website-builder.tests.openpanel.org')).not.toBeVisible();
  console.log('website uninstall is working');

  // 5. install again and test detach
  await page.goto('/website-builder/install');
  await page.locator('#domain_id').selectOption('website-builder.tests.openpanel.org');
  await page.locator('#installButton').click();
  await expect(page.locator('text=Website creation completed!')).toBeVisible({ timeout: 60000 });
  await expect(page).toHaveURL(/\/website-builder\/edit\?domain=website-builder\.tests\.openpanel\.org\/.+/);

  await page.goto('/website?domain=website-builder.tests.openpanel.org');
  await page.locator('a#remove-tab').click();
  await page.locator('button#detach-site').click();
  await page.locator('button#confirm-detach-site').click();
  await expect(page.locator('text=Website detached successfully!')).toBeVisible({ timeout: 30000 });
  await page.goto('/sites');
  await expect(page.locator('tr#site-row-website-builder.tests.openpanel.org')).not.toBeVisible();
  console.log('website detach is working');
  // TODO: remove files

});
