import { test, expect } from '@playwright/test';

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

test('edit cpu, ram and toggle container state for all rows', async ({ page }) => {
	test.setTimeout(9000_000);
  await page.goto('/containers'); // adjust URL

  const rows = page.locator('#containers-table tr.domain_row');
  const count = await rows.count();

  for (let i = 0; i < count; i++) {
    const row = rows.nth(i);

    // ensure hover tools appear
    await row.hover();

    // -------------------------
    // 1. EDIT CPU
    // -------------------------
    const cpuCell = row.locator('td[x-show*="cpu"]');

    await cpuCell.hover();

    // click pencil (hover-revealed)
    await cpuCell.locator('a').first().click();

    const cpuInput = cpuCell.locator('input[name="value"]');
    await expect(cpuInput).toBeVisible();

    await cpuInput.fill('1'); // new CPU value

    await Promise.all([
      page.waitForResponse(resp =>
        resp.url().includes('/containers/cpu/') && resp.request().method() === 'POST'
      ),
      cpuCell.locator('button[type="submit"]').click(),
    ]);

    // wait for edit mode to close
    await expect(cpuInput).toBeHidden();
    await expect(page.locator('body')).toContainText(/Max CPU for container/i);
    await expect(page.locator('body')).toContainText(/set to 1/i);
    // -------------------------
    // 2. EDIT RAM
    // -------------------------
    const ramCell = row.locator('td[x-show*="ram"]');

    await ramCell.hover();
    await ramCell.locator('a').first().click();

    const ramInput = ramCell.locator('input[name="value"]');
    await expect(ramInput).toBeVisible();

    await ramInput.fill('1'); // new RAM value

    await Promise.all([
      page.waitForResponse(resp =>
        resp.url().includes('/containers/ram/') && resp.request().method() === 'POST'
      ),
      ramCell.locator('button[type="submit"]').click(),
    ]);

    await expect(ramInput).toBeHidden();
    await expect(page.locator('body')).toContainText(/Max RAM for container/i);
    await expect(page.locator('body')).toContainText(/set to 1G/i);
	// -------------------------
	// 3. ENSURE FINAL STATE = STOPPED
	// -------------------------

	// Re-locate row after CPU/RAM updates
	const freshRow = rows.nth(i);
	await freshRow.hover();

	const actionsCell = freshRow.locator('td[x-show*="actions"]');

	const startBtn = actionsCell.getByRole('button', { name: 'Start', exact: true }).first();
	const stopBtn = actionsCell.getByRole('button', { name: 'Stop', exact: true }).first();

	const hasStartBtn = await startBtn.count();

	if (hasStartBtn > 0 && await startBtn.isVisible()) {
	  await Promise.all([
		page.waitForResponse(resp =>
		  resp.url().includes('/containers/start/') &&
		  resp.request().method() === 'POST'
		),
		startBtn.click(),
	  ]);

	  await expect(page.locator('body')).toContainText(/activated successfully/i);

	  await expect(stopBtn).toBeVisible({
		timeout: 30_000,
	  });
	}

	const hasStopBtn = await stopBtn.count();

	if (hasStopBtn > 0 && await stopBtn.isVisible()) {
	  await Promise.all([
		page.waitForResponse(resp =>
		  resp.url().includes('/containers/stop/') &&
		  resp.request().method() === 'POST'
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




test('terminal', async ({ page }) => {
  test.setTimeout(90_000);


  await page.goto('/containers/terminal/php-fpm-8.0');
  await page.locator('.xterm-rows > div').first().click();
  await page.locator('.xterm-rows > div').first().click({
    button: 'right'
  });
  await page.getByRole('textbox', { name: 'Terminal input' }).fill('php -v');
  await page.getByRole('textbox', { name: 'Terminal input' }).press('Enter');

  const terminalOutput = page.locator('.xterm-rows');

  await expect(terminalOutput).toContainText('The PHP Group');

  console.log('php -v executed successfully');
});
