import { test, expect, Page } from '@playwright/test';

const domain = 'wp.tests.openpanel.org';

const versions = [ '8.5', '8.4', '8.3', '8.2', '8.1', '8.0', '7.4', '7.3', '7.2', '7.1', '7.0', '5.6' ];

test.setTimeout(10 * 60 * 1000); // 10min so we can also download docker images

async function openPhpPage(page: Page) {
  await page.goto('/php/domains');
  await expect(page.locator('table')).toBeVisible();
}

function domainRows(page: Page) {
  return page.locator('tbody tr');
}

test('list versions', async ({ page }) => {
  await openPhpPage(page);

  // header is ok
  await expect(page.getByRole('heading', { name: /PHP version for domains/i })).toBeVisible();

  // table is ok
  const headers = page.locator('thead th');
  await expect(headers).toHaveCount(3);
  await expect(headers.nth(0)).toContainText(/domain/i);
  await expect(headers.nth(1)).toContainText(/current php version/i);
  await expect(headers.nth(2)).toContainText(/change version/i);

  // versions are shown in the table
  const rows = domainRows(page);
  const rowCount = await rows.count();

  if (rowCount === 1) {
    const empty = rows.first();
    const text = await empty.textContent();
    if (text?.includes('No domains')) {
      test.skip();
      return;
    }
  }

  expect(rowCount).toBeGreaterThan(0);
  for (let i = 0; i < rowCount; i++) {
    const row = rows.nth(i);
    const versionCell = row.locator('td').nth(1);

    // version format is ok
    const text = await versionCell.textContent();
    expect(text?.trim()).toMatch(/\d+\.\d+/);
  
    // status indicators are ok
    const bars = versionCell.locator('div.flex.gap-0\\.5 > div');
    await expect(bars).toHaveCount(3);   
  }

  // summary per version
  const counters = page.locator('dl > div');
  const count = await counters.count();
  if (rowCount > 0) {
    expect(count).toBeGreaterThan(0);
  }  

});



test.describe('search filter', () => {
  test('filter table rows', async ({ page }) => {
    await openPhpPage(page);

    const rows = domainRows(page);
    const totalRows = await rows.count();
    if (totalRows < 2) return;

    const firstDomainText = (await rows.first().locator('td').first().textContent()) ?? '';
    const searchTerm = firstDomainText.trim().split('.')[0];

    await page.locator('input[x-model="searchQuery"]').fill(searchTerm);

    await page.waitForTimeout(300);

    const visibleRows = rows.filter({ hasNot: page.locator('[style*="display: none"]') });
    const visibleCount = await visibleRows.count();
    expect(visibleCount).toBeGreaterThanOrEqual(1);
    expect(visibleCount).toBeLessThanOrEqual(totalRows);
  });

  test('version counter filter', async ({ page }) => {
    await openPhpPage(page);

    const counterLink = page.locator('dl a').first();
    const count = await counterLink.count();
    if (count === 0) return;

    await counterLink.click();

    await page.waitForTimeout(300);

    const searchValue = await page.locator('input[x-model="searchQuery"]').inputValue();
    expect(searchValue).toMatch(/\d+\.\d+/);
  });

  test('clear search', async ({ page }) => {
    await openPhpPage(page);

    const rows = domainRows(page);
    const totalRows = await rows.count();
    if (totalRows < 1) return;

    const searchInput = page.locator('input[x-model="searchQuery"]');
    await searchInput.fill('xyznonexistent999');
    await page.waitForTimeout(300);

    await searchInput.fill('');
    await page.waitForTimeout(300);

    const hiddenRows = page.locator('tbody tr[style*="display: none"]');
    await expect(hiddenRows).toHaveCount(0);
  });
});



test.describe('version change', () => {

  test.beforeAll(async ({ browser }) => {
    const page = await browser.newPage();

    // Ensure info.php exists
    await page.goto(`/file-manager/edit-file/${domain}/info.php?editor=text&new=true`);
    await page.locator('#editor-text').fill('<?php phpinfo();');
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText(/saved|success/i).first()).toBeVisible();

    await page.close();
  });

  for (const version of versions) {
    test(`php ${version}`, async ({ page }) => {
      test.setTimeout(120_000);

      await openPhpPage(page);

      const rows = domainRows(page);
      const rowCount = await rows.count();

      let targetRow = null;
      for (let i = 0; i < rowCount; i++) {
        const row = rows.nth(i);
        const domainCell = await row.locator('td').first().textContent();
        if (domainCell?.trim() && domain.includes(domainCell.trim())) {
          targetRow = row;
          break;
        }
      }

      if (!targetRow) {
        test.skip(true, 'Domain not found');
        return;
      }

      const select = targetRow.locator('select[name="new_php_version"]');

      const currentVersion =
        (await targetRow.locator('td').nth(1).textContent())?.match(/\d+\.\d+/)?.[0] ?? 'unknown';

      await select.selectOption(version);

      await Promise.all([
        page.waitForResponse(
          res => res.request().method() === 'POST' && res.status() === 200,
          { timeout: 90_000 }
        ),
        targetRow.getByRole('button', { name: /change/i }).click(),
      ]);

      await expect(page.getByText(new RegExp(`updated from ${currentVersion} to ${version}`, 'i'))).toBeVisible();
      const versionShort = version.match(/\d+\.\d+/)?.[0] ?? version;

      await expect(async () => {
        await page.goto(`https://${domain}/info.php?nocache=${Date.now()}`);
        await expect(page.locator('body')).toContainText(`PHP Version ${versionShort}`);
      }).toPass({ timeout: 30000, intervals: [500] }); // 30s max, every 0.5s

      console.log(`php ${versionShort} is working`);
    });
  }
});


test('change default php version', async ({ page }) => {
  await page.goto('/php/default');
  await expect(page.getByText(/Current default version/i)).toBeVisible();
  
  const oldVersion = await page.locator('#current_default_version').innerText();
  console.log(`Starting version: ${oldVersion}`);

  const dropdown = page.locator('#new_php_version');
  const values = await dropdown.locator('option').evaluateAll(options => 
    options
      .map(opt => opt.value)
      .filter(val => val !== "" && val !== "oldVersionValueHere") 
  );
  
  if (values.length === 0) {
    throw new Error('No alternative PHP versions available to select.');
  }

  const randomVersion = values[Math.floor(Math.random() * values.length)];
  await dropdown.selectOption(randomVersion);
  await page.click('#change-php-version');

  const successRegex = new RegExp(`PHP version ${randomVersion} set as default for new domains`, 'i');
  await expect(page.getByText(successRegex)).toBeVisible();
  await expect(page.locator('#current_default_version')).toHaveText(randomVersion);

  console.log(`Default PHP version switch to ${randomVersion} is working`);
  // TODO: test if used on a new domain!
});



test('edit php options', async ({ page }) => {
  await page.goto('/php/options');  
  const dropdown = page.locator('#php_version');
  const values = await dropdown.locator('option').evaluateAll(options => 
    options
      .map(opt => opt.value)
      .filter(val => val !== "" && val !== "oldVersionValueHere") 
  );
  
  if (values.length === 0) {
    throw new Error('No alternative PHP versions available to select.');
  }

  const randomVersion = values[Math.floor(Math.random() * values.length)];
  await dropdown.selectOption(randomVersion);
  await page.click('#submit_version');

  const successRegex = new RegExp(`${randomVersion} Options`, 'i');
  await expect(page.getByText(successRegex).first()).toBeVisible();
  await expect(page).toHaveURL(new RegExp(`/php/php${randomVersion}/options\\?php_version=${randomVersion}`));

  // 1. max_execution_time
  const maxExecTime = page.locator('input[name="max_execution_time"]');
  await maxExecTime.fill('600');

  // 2. disable_functions
  const disableFuncs = page.locator('input[name="disable_functions"]');
  await disableFuncs.fill('exec,passthru,shell_exec,system');

  // 3. post_max_size
  const postMaxSizeContainer = page.locator('div[data-key="post_max_size"]');
  await postMaxSizeContainer.locator('input.numeric-part').fill('2');
  await postMaxSizeContainer.locator('select.unit-part').selectOption('G');
  
  await page.click('#save-changes-button');
  await expect(page.getByText(/Configuration edited successfully and PHP-FPM service restarted/i)).toBeVisible();

  await expect(maxExecTime).toHaveValue('600');
  await expect(disableFuncs).toHaveValue('exec,passthru,shell_exec,system');
  await expect(postMaxSizeContainer.locator('input.numeric-part')).toHaveValue('2');
  await expect(postMaxSizeContainer.locator('select.unit-part')).toHaveValue('G');
  
  console.log(`PHP options editor for version ${randomVersion} is working and verified.`);
  // TODO: test on a website and test on page Search to filter table!
});



test('edit php.ini files', async ({ page }) => {
  await page.goto('/php/php_ini_editor');  
  const dropdown = page.locator('#php_version');
  const values = await dropdown.locator('option').evaluateAll(options =>
    options
      .map(opt => opt.value)
      .filter(val => val !== '')
  );

  if (values.length === 0) {
    throw new Error('No PHP versions available to select.');
  }

  const randomVersion = values[Math.floor(Math.random() * values.length)];
  await dropdown.selectOption(randomVersion);
  await page.click('#submit_version');
  await page.waitForURL(new RegExp(`/php/php${randomVersion}\\.ini/editor`));

  const successRegex = new RegExp(`Edit PHP\\.INI file for version ${randomVersion}`, 'i');
  await expect(page.getByRole('heading', { name: successRegex })).toBeVisible();
  await expect(page).toHaveURL(new RegExp(`/php/php${randomVersion}.ini/editor\\?php_version=${randomVersion}`));

  const editorLocator = page.locator('.CodeMirror'); 
  await expect(editorLocator).toBeVisible();

  const updates = {
    'max_input_time': '120',
    'opcache.enable': '1'
  };
  
  await page.evaluate((settings) => {
    const cm = (document.querySelector('.CodeMirror') as any).CodeMirror;
    let content = cm.getValue();
  
    for (const [key, value] of Object.entries(settings)) {
      const escapedKey = key.replace('.', '\\.');
      const regex = new RegExp(`^;?(${escapedKey}\\s*=\\s*).*`, 'm');
      if (regex.test(content)) {
        content = content.replace(regex, `${key} = ${value}`);
      } else {
        throw new Error(`Directive "${key}" not found in php.ini`);
      }
    }
    cm.setValue(content);
  }, updates);

  await page.click('#save-changes-button');
  const feedbackRegex = new RegExp(`PHP.INI file for PHP-FPM version ${randomVersion} edited successfully`, 'i');
  await expect(page.getByText(feedbackRegex)).toBeVisible();

  const updatedContent = await page.evaluate(() => {
    return document.querySelector('.CodeMirror').CodeMirror.getValue(); //or just textarea!
  });
  
  expect(updatedContent).toContain('max_input_time = 120');
  expect(updatedContent).toContain('opcache.enable = 1');

  console.log(`PHP.INI editor for version ${randomVersion} is working and verified.`);
  // TODO: test on a website
});
