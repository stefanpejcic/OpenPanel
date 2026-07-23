import { test, expect } from '@playwright/test';


test('access webserver configuration page', async ({ page }) => {
  await page.goto('/server/webserver_conf');
  await expect(page).toHaveURL(/webserver_conf/);
  await expect(page.locator('body')).toContainText(/webserver|nginx|apache|configuration/i);
  console.log('webserver configuration page accessible');
});


test('editor contains valid config content', async ({ page }) => {
  await page.goto('/server/webserver_conf');

  // CodeMirror editor or textarea
  const hasCodeMirror = await page.locator('.CodeMirror').isVisible({ timeout: 3000 }).catch(() => false);
  const hasTextarea = await page.locator('textarea[name="file_content"], textarea#editor').isVisible({ timeout: 3000 }).catch(() => false);

  expect(hasCodeMirror || hasTextarea).toBe(true);

  if (hasCodeMirror) {
    const content = await page.evaluate(() => {
      const cm = (document.querySelector('.CodeMirror') as any)?.CodeMirror;
      return cm ? cm.getValue() : '';
    });
    expect(content.length).toBeGreaterThan(10);
    console.log(`editor has ${content.length} chars of config`);
  } else {
    const content = await page.locator('textarea[name="file_content"], textarea#editor').first().inputValue();
    expect(content.length).toBeGreaterThan(10);
    console.log(`textarea has ${content.length} chars of config`);
  }
});


test('save webserver configuration', async ({ page }) => {
  await page.goto('/server/webserver_conf');

  const saveBtn = page.getByRole('button', { name: /save|apply/i });
  await expect(saveBtn).toBeVisible();
  await saveBtn.click();

  await expect(page.locator('body')).toContainText(/saved|updated|restarted|success/i, { timeout: 30000 });
  console.log('webserver configuration saved successfully');
});


test('invalid config is rejected', async ({ page }) => {
  await page.goto('/server/webserver_conf');

  const hasCodeMirror = await page.locator('.CodeMirror').isVisible({ timeout: 3000 }).catch(() => false);

  if (hasCodeMirror) {
    await page.evaluate(() => {
      const cm = (document.querySelector('.CodeMirror') as any)?.CodeMirror;
      if (cm) cm.setValue('THIS IS COMPLETELY INVALID NGINX/APACHE CONFIG $$$');
    });
  } else {
    const textarea = page.locator('textarea[name="file_content"], textarea#editor').first();
    await textarea.fill('THIS IS COMPLETELY INVALID NGINX/APACHE CONFIG $$$');
  }

  await page.getByRole('button', { name: /save|apply/i }).click();

  await expect(page.locator('body')).toContainText(/error|invalid|failed|syntax/i, { timeout: 30000 });
  console.log('invalid config correctly rejected');
});
