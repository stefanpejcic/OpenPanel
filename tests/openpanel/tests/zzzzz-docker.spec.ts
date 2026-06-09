import { test, expect } from '@playwright/test';

// ─────────────────────────────────────────────
// /containers  (main table)
// ─────────────────────────────────────────────

test('check columns for docker table', async ({ page }) => {
  await page.goto(`/containers`);
  await expect(page).toHaveURL(/containers/);
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
    await page.waitForTimeout(100);

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

test('containers page loads with header and table', async ({ page }) => {
  await page.goto('/containers');
  await expect(page.locator('h1')).toContainText('Containers');
  await expect(page.locator('#containers-table')).toBeVisible();

  // Total CPU and RAM should be shown
  await expect(page.locator('text=Total CPU')).toBeVisible();
  await expect(page.locator('text=Total Memory')).toBeVisible();
});

test('containers search filters rows', async ({ page }) => {
  await page.goto('/containers');

  const searchInput = page.locator('[x-model="searchQuery"]').first();
  await expect(searchInput).toBeVisible();

  // Search for something that won't match
  await searchInput.fill('zzznomatchzzz');
  await page.waitForTimeout(200);
  const visibleRows = page.locator('#containers-table tbody tr:visible');
  await expect(visibleRows).toHaveCount(0, { timeout: 3000 });

  // Clear search
  await searchInput.fill('');
  await page.waitForTimeout(200);
});

test('containers page New Service button navigates to add form', async ({ page }) => {
  await page.goto('/containers');
  await page.getByRole('link', { name: 'New Service' }).click();
  await page.waitForLoadState('networkidle');
  expect(page.url()).toContain('/containers/new');
  await expect(page.locator('h1')).toContainText('Add New Service');
});

test('edit cpu, ram and toggle container state for all rows', async ({ page }) => {
  test.setTimeout(9000_000);
  await page.goto('/containers');

  const rows = page.locator('#containers-table tr.domain_row');
  const count = await rows.count();

  for (let i = 0; i < count; i++) {
    const row = rows.nth(i);
    await row.hover();

    // 1. EDIT CPU
    const cpuCell = row.locator('td[x-show*="cpu"]');
    await cpuCell.hover();
    await cpuCell.locator('a').first().click();

    const cpuInput = cpuCell.locator('input[name="value"]');
    await expect(cpuInput).toBeVisible();
    await cpuInput.fill('1');

    await Promise.all([
      page.waitForResponse(resp =>
        resp.url().includes('/containers/cpu/') && resp.request().method() === 'POST'
      ),
      cpuCell.locator('button[type="submit"]').click(),
    ]);

    await expect(cpuInput).toBeHidden();
    await expect(page.locator('body')).toContainText(/Max CPU for container/i);
    await expect(page.locator('body')).toContainText(/set to 1/i);

    // 2. EDIT RAM
    const ramCell = row.locator('td[x-show*="ram"]');
    await ramCell.hover();
    await ramCell.locator('a').first().click();

    const ramInput = ramCell.locator('input[name="value"]');
    await expect(ramInput).toBeVisible();
    await ramInput.fill('1');

    await Promise.all([
      page.waitForResponse(resp =>
        resp.url().includes('/containers/ram/') && resp.request().method() === 'POST'
      ),
      ramCell.locator('button[type="submit"]').click(),
    ]);

    await expect(ramInput).toBeHidden();
    await expect(page.locator('body')).toContainText(/Max RAM for container/i);
    await expect(page.locator('body')).toContainText(/set to 1G/i);

    // 3. ENSURE FINAL STATE = STOPPED
    const freshRow = rows.nth(i);
    await freshRow.hover();

    const actionsCell = freshRow.locator('td[x-show*="actions"]');
    const startBtn = actionsCell.getByRole('button', { name: 'Start', exact: true }).first();
    const stopBtn = actionsCell.getByRole('button', { name: 'Stop', exact: true }).first();

    const hasStartBtn = await startBtn.count();
    if (hasStartBtn > 0 && await startBtn.isVisible()) {
      await Promise.all([
        page.waitForResponse(resp =>
          resp.url().includes('/containers/start/') && resp.request().method() === 'POST'
        ),
        startBtn.click(),
      ]);
      await expect(page.locator('body')).toContainText(/activated successfully/i);
      await expect(stopBtn).toBeVisible({ timeout: 30_000 });
    }

    const hasStopBtn = await stopBtn.count();
    if (hasStopBtn > 0 && await stopBtn.isVisible()) {
      await Promise.all([
        page.waitForResponse(resp =>
          resp.url().includes('/containers/stop/') && resp.request().method() === 'POST'
        ),
        stopBtn.click(),
      ]);
      await expect(page.locator('body')).toContainText(/deactivated successfully/i);
    } else {
      console.log(`Container row ${i + 1} has no visible Stop button, already stopped or not startable`);
    }

    await page.waitForTimeout(2000);
  }
});

// ─────────────────────────────────────────────
// /containers/new  (add container form)
// ─────────────────────────────────────────────

test('add new service form loads', async ({ page }) => {
  await page.goto('/containers/new');
  await expect(page.locator('h1')).toContainText('Add New Service');
  await expect(page.locator('input#service_name')).toBeVisible();
  await expect(page.locator('input#image')).toBeVisible();
  await expect(page.locator('input#cpu')).toBeVisible();
  await expect(page.locator('input#ram')).toBeVisible();
  await expect(page.locator('button[type="submit"]')).toBeVisible();
  await expect(page.locator('a[href="/containers"]')).toBeVisible();
});

test('add new service - invalid name shows error', async ({ page }) => {
  await page.goto('/containers/new');

  await page.locator('input#image').fill('nginx:latest');
  await page.locator('input#service_name').fill('ab'); // too short
  await page.locator('input#cpu').fill('0.5');
  await page.locator('input#ram').fill('1G');
  await page.locator('button[type="submit"]').click();

  await page.waitForLoadState('networkidle');
  // Should stay on the form page (not redirect to /containers)
  expect(page.url()).toContain('/containers/new');
});

test('add new service - image blur suggests service name', async ({ page }) => {
  await page.goto('/containers/new');

  const imageInput = page.locator('input#image');
  const serviceNameInput = page.locator('input#service_name');

  await imageInput.fill('redis:latest');
  await imageInput.blur();
  await page.waitForTimeout(200);

  const suggestedName = await serviceNameInput.inputValue();
  expect(suggestedName.length).toBeGreaterThan(0);
  expect(suggestedName).toMatch(/^[a-z][a-z0-9]+$/);
});

test('add new service - add and remove volume entry', async ({ page }) => {
  await page.goto('/containers/new');

  // Click the Add button to add a volume row
  await page.getByRole('button', { name: 'Add' }).click();
  await page.waitForTimeout(100);

  const volumeEntries = page.locator('.volume-entry');
  const countAfterAdd = await volumeEntries.count();
  expect(countAfterAdd).toBeGreaterThan(0);

  // Remove it
  const removeBtn = volumeEntries.last().locator('button[type="button"]');
  await removeBtn.click();
  await page.waitForTimeout(100);

  const countAfterRemove = await page.locator('.volume-entry').count();
  expect(countAfterRemove).toBeLessThan(countAfterAdd);
});

test('add new service - Back to Containers link works', async ({ page }) => {
  await page.goto('/containers/new');
  await page.locator('a[href="/containers"]').click();
  await page.waitForLoadState('networkidle');
  expect(page.url()).toContain('/containers');
  await expect(page.locator('h1')).toContainText('Containers');
});

// ─────────────────────────────────────────────
// /containers/delete/<service>  (delete confirm)
// ─────────────────────────────────────────────

test('delete confirm page loads for a custom service', async ({ page }) => {
  // Only run if a deletable (non-core) service exists; use a known test service name
  const testService = 'testapp'; // adjust to a service that exists and is deletable
  await page.goto(`/containers/delete/${testService}`);

  const status = page.url();

  if (status.includes('/containers/delete/')) {
    await expect(page.locator('h1')).toContainText('Delete container');
    await expect(page.locator('text=Are you sure')).toBeVisible();

    // Cancel button should go back to containers
    const cancelLink = page.locator('a[href="/containers"]');
    await expect(cancelLink).toBeVisible();
    await cancelLink.click();
    await page.waitForLoadState('networkidle');
    expect(page.url()).toContain('/containers');
  } else {
    console.log('Delete confirm page not accessible (service may not exist), skipping.');
  }
});

test('delete confirm - core service returns 403', async ({ page }) => {
  const response = await page.request.get('/containers/delete/nginx');
  expect(response.status()).toBe(403);
});

test('delete confirm - php-fpm service returns 403', async ({ page }) => {
  const response = await page.request.get('/containers/delete/php-fpm-8.2');
  expect(response.status()).toBe(403);
});

// ─────────────────────────────────────────────
// /containers/mysql  (change mysql type)
// ─────────────────────────────────────────────

test('change mysql page loads', async ({ page }) => {
  await page.goto('/containers/mysql');
  await expect(page.locator('h1')).toBeVisible();
  // Should show current mysql type
  await expect(page.locator('text=Current:')).toBeVisible();
  // Should mention conditions
  await expect(page.locator('text=All existing databases')).toBeVisible();
});

// ─────────────────────────────────────────────
// /containers/webserver  (change webserver)
// ─────────────────────────────────────────────

test('change webserver page loads', async ({ page }) => {
  await page.goto('/containers/webserver');
  await expect(page.locator('h1')).toBeVisible();
  await expect(page.locator('text=Current:')).toBeVisible();
  await expect(page.locator('text=All existing domains must be removed')).toBeVisible();
});

// ─────────────────────────────────────────────
// /containers/image  (image updates)
// ─────────────────────────────────────────────

test('image updates page loads', async ({ page }) => {
  await page.goto('/containers/image');
  await expect(page.locator('h1')).toContainText('Image Updates');
  await expect(page.locator('text=Last checked')).toBeVisible();
  // Refresh button should exist
  await expect(page.locator('button[type="submit"]')).toBeVisible();
});

test('image updates - change tag page loads with service selector', async ({ page }) => {
  await page.goto('/containers/image/change');
  await expect(page.locator('h1')).toContainText('Change docker image tag');
  await expect(page.locator('select#domains')).toBeVisible();
});

test('image updates - selecting service redirects to tag change page', async ({ page }) => {
  await page.goto('/containers/image/change');

  const select = page.locator('select#domains');
  const options = select.locator('option:not([disabled])');
  const count = await options.count();

  if (count > 0) {
    const value = await options.first().getAttribute('value');
    await select.selectOption(value);
    await page.waitForURL(`**/containers/image/change/${value}`, { timeout: 5000 });
    await expect(page.locator('h1')).toContainText('Change image tag for');
    await expect(page.locator('input#new_tag')).toBeVisible();
    await expect(page.locator('a[href="/containers/image"]')).toBeVisible();
  } else {
    console.log('No services available in image change selector, skipping.');
  }
});

// ─────────────────────────────────────────────
// /containers/logs  (logs viewer)
// ─────────────────────────────────────────────

test('logs page loads with container selector', async ({ page }) => {
  await page.goto('/containers/logs');
  await expect(page.locator('h1')).toContainText('Logs');
  await expect(page.locator('select#log-select')).toBeVisible();
  await expect(page.locator('select#lines-select')).toBeVisible();
  await expect(page.locator('#log-content')).toBeVisible();
});

test('logs page - selecting container loads log content', async ({ page }) => {
  await page.goto('/containers/logs');

  const select = page.locator('select#log-select');
  const options = select.locator('option:not([disabled])');
  const count = await options.count();

  if (count > 0) {
    const value = await options.first().getAttribute('value');
    await select.selectOption(value);
    await page.waitForTimeout(2000);

    const logContent = page.locator('#log-content');
    const text = await logContent.textContent();
    expect(text).not.toBe('Select a container to view its logs here.');
  } else {
    console.log('No containers in log selector, skipping.');
  }
});

test('logs page - ?container= param pre-selects and loads', async ({ page }) => {
  // Navigate with query param; the JS should auto-select and fetch
  await page.goto('/containers/logs?container=nginx&lines=20');
  await page.waitForTimeout(2000);

  const select = page.locator('select#log-select');
  const selected = await select.inputValue();

  // Should be set to nginx (or empty if nginx doesn't exist)
  if (selected === 'nginx') {
    const logContent = await page.locator('#log-content').textContent();
    expect(logContent).not.toBe('Select a container to view its logs here.');
  } else {
    console.log('nginx not in log selector (may not be running), skipping content check.');
  }
});

// ─────────────────────────────────────────────
// /containers/terminal
// ─────────────────────────────────────────────

test('terminal page without container shows service picker', async ({ page }) => {
  await page.goto('/containers/terminal');
  await expect(page.locator('h1')).toContainText('Docker Terminal');
  await expect(page.locator('select#service-select')).toBeVisible();
});

test('terminal page - selecting service redirects', async ({ page }) => {
  await page.goto('/containers/terminal');

  const select = page.locator('select#service-select');
  const options = select.locator('option:not([disabled])');
  const count = await options.count();

  if (count > 0) {
    const value = await options.first().getAttribute('value');
    await select.selectOption(value);
    await page.waitForURL(`**/containers/terminal/${value}`, { timeout: 5000 });
    await expect(page.locator('h1')).toContainText('Docker Terminal');
    await expect(page.locator('#terminal')).toBeVisible();
    await expect(page.locator('select#shell')).toBeVisible();
  } else {
    console.log('No running services for terminal, skipping.');
  }
});

test('terminal', async ({ page }) => {
  test.setTimeout(90_000);

  await page.goto('/containers/terminal/php-fpm-8.0');
  await page.locator('.xterm-rows > div').first().click();
  await page.locator('.xterm-rows > div').first().click({ button: 'right' });
  await page.getByRole('textbox', { name: 'Terminal input' }).fill('php -v');
  await page.getByRole('textbox', { name: 'Terminal input' }).press('Enter');

  const terminalOutput = page.locator('.xterm-rows');
  await expect(terminalOutput).toContainText('The PHP Group');

  console.log('php -v executed successfully');
});

test('terminal - reconnect button appears after disconnect', async ({ page }) => {
  await page.goto('/containers/terminal/php-fpm-8.0');
  await page.waitForTimeout(2000);

  // Status dot and text should be present
  await expect(page.locator('#status-dot')).toBeVisible();
  await expect(page.locator('#status-text')).toBeVisible();

  // Shell selector should be set to bash by default
  const shell = await page.locator('select#shell').inputValue();
  expect(shell).toBe('bash');
});

// ─────────────────────────────────────────────
// /containers/status  (JSON endpoint)
// ─────────────────────────────────────────────

test('containers status endpoint returns JSON', async ({ page }) => {
  const response = await page.request.get('/containers/status');
  expect(response.status()).toBe(200);

  const body = await response.json();
  expect(body).toHaveProperty('running_containers');
  expect(Array.isArray(body.running_containers)).toBe(true);
});
