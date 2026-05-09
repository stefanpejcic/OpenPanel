import { test, expect } from '@playwright/test';



test('list', async ({ page }) => {
  await page.goto('/cronjobs');
  await expect(page.getByText(/no cronjobs yet/i)).toBeVisible();
  await expect(page.getByRole('link', { name: /create new/i })).toBeVisible();
  await expect(page.getByRole('link', { name: /switch to file editor/i })).toBeVisible();
  console.log(`cronjobs functional`);
});



test('create job', async ({ page }) => {
  await page.goto('/cronjobs/new');
  await expect(page).toHaveURL(/\/cronjobs\/new/);

  await page.selectOption('#container', 'php-fpm-8.5');
  await page.fill('#schedule', '@every 5s');
  const testCommand = 'curl https://google.com > /var/www/html/cron-test.txt';
  await page.fill('#command', testCommand);
  await page.fill('#comment', 'curl job');

  await page.getByRole('button', { name: 'Schedule CronJob' }).click();

  await expect(page.getByText('Cron job created and saved successfully!')).toBeVisible();

  const tableRow = page.locator('tr', { hasText: 'curl job' });
  await expect(tableRow).toBeVisible();
  await expect(tableRow).toContainText(testCommand);

  // TODO: check if service auto-started by fetching /api/services?name=cron

  console.log(`cronjob created`);
});



test('view logs', async ({ page }) => {
  await page.waitForTimeout(15000); // wait for contianer to start and created job to run
  await page.goto('/cronjobs');

  const tableRow = page.locator('tr', { hasText: 'curl job' });
  await expect(tableRow).toBeVisible();

  const logsButton = tableRow.locator('button', {
    hasText: /log|logs/i,
  });

  await logsButton.click();

  const logsPanel = page.locator('[x-show="logsOpen"]');
  await expect(logsPanel).toBeVisible();

  const logRows = logsPanel.locator('table tbody tr');
  await expect(logRows.first()).toBeVisible();
  expect(await logRows.count()).toBeGreaterThan(0);

  console.log('cronjob logs working');
});



test('edit as file', async ({ page }) => {
  await page.goto('/cronjobs?view=code');
  await expect(page).toHaveURL(/\/cronjobs\?view=code/);

  const expectedCron = `[job-exec "curl job"]
schedule = @every 5s
container = php-fpm-8.5
command = curl https://google.com > /var/www/html/cron-test.txt`;

  const actualContent = await page.evaluate(() => {
    return document.querySelector('.CodeMirror').CodeMirror.getValue();
  });

  expect(actualContent.trim()).toBe(expectedCron);

  const updatedContent = expectedCron.replace('@every 5s', '* * * * * *');
  
  await page.evaluate((val) => {
    const cm = document.querySelector('.CodeMirror').CodeMirror;
    cm.setValue(val);
    cm.save();
  }, updatedContent);

  await page.getByRole('button', { name: 'Save Changes' }).click();
  
  await expect(page.getByText('Crontab file saved successfully!')).toBeVisible();

  const postSaveContent = await page.evaluate(() => {
    return document.querySelector('.CodeMirror').CodeMirror.getValue();
  });
  expect(postSaveContent).toContain('schedule = * * * * * *');

  await page.goto('/cronjobs?view=table');
  await expect(page).toHaveURL(/\/cronjobs\?view=table/);
  const tableRow = page.locator('tr', { hasText: 'curl job' });
  await expect(tableRow).toBeVisible();
  await expect(tableRow).toContainText(`* * * * * *`);

  console.log(`cronjob file editor working`);
});



test('edit job', async ({ page }) => {
  await page.goto('/cronjobs?view=table');

  let tableRow = page.locator('tr', { hasText: 'curl job' });
  await expect(tableRow).toBeVisible();

  const edits = [
    { label: 'schedule', newValue: '0 0 * * *' },
    { label: 'container', newValue: 'php-fpm-8.4', isSelect: true },
    { label: 'command', newValue: 'curl https://google.com' },
    { label: 'comment', newValue: 'updated description' },
  ];

  for (const edit of edits) {
    await tableRow.getByRole('button', { name: /Edit/i }).click();

    if (edit.isSelect) {
      await tableRow.locator('select[name="container"]').selectOption({ label: edit.newValue });
    } else {
      const input = tableRow.locator(`input[name="${edit.label}"]:visible`);
      await input.fill(edit.newValue);
    }

    await tableRow.getByRole('button', { name: /Save/i }).click();
    await expect(page.getByText(/successfully edited/i)).toBeVisible();

    if (edit.label === 'comment') {
       tableRow = page.locator('tr', { hasText: edit.newValue });
    }

    await expect(tableRow).toContainText(edit.newValue);
  }
});



test('delete job', async ({ page }) => {
  await page.goto('/cronjobs?view=table');
  await expect(page).toHaveURL(/\/cronjobs\?view=table/);

  const tableRow = page.locator('tr', { hasText: /curl job|updated description/ });
  await expect(tableRow).toBeVisible();

  await tableRow.getByRole('button', { name: /Delete/i }).click();
  
  await expect(page.getByText('Cron job was successfully deleted.')).toBeVisible();

  const remainingRows = page.locator('tbody tr');
  await expect(remainingRows).toHaveCount(0);

  console.log(`delete working`);
});
