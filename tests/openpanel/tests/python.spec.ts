import { test, expect } from '@playwright/test';

// https://github.com/stefanpejcic/python-helloworld

const DOMAIN = 'python.tests.openpanel.org';
const APP_NAME = 'pythonaplikacija';
const PORT = '5000';
const PYTHON_VERSION = '3.13.13';
const STARTUP_FILE = `/var/www/html/${DOMAIN}/app.py`;

const REQUIREMENTS_TXT = `Flask==2.3.3`;

const APP_PY = `from flask import Flask
app = Flask(__name__)
@app.route('/')
def hello():
    return "Hello World from Flask on port 5000!"
if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)`;

test.describe.configure({ mode: 'serial' });

test.describe('Python PM2 autoinstaller', () => {
  test('1. create app files', async ({ page }) => {
    // requirements.txt
    await page.goto(`/file-manager/edit-file/${DOMAIN}/requirements.txt?editor=text&new=true`);
    await page.locator('#editor-text').fill(REQUIREMENTS_TXT);
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText(/saved|success/i).first()).toBeVisible();

    // app.py
    await page.goto(`/file-manager/edit-file/${DOMAIN}/app.py?editor=text&new=true`);
    await page.locator('#editor-text').fill(APP_PY);
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText(/saved|success/i).first()).toBeVisible();
  });

  test('2. install PM2 app', async ({ page }) => {
    await page.goto('/pm2/install#python');
    await page.locator('#service_name').fill(APP_NAME);
    await page.getByRole('spinbutton', { name: 'Port:' }).fill(PORT);
    await page.getByLabel('Domain:').selectOption({ label: DOMAIN });
    await page.getByRole('textbox', { name: /Startup file/i }).fill(STARTUP_FILE);
    await page.getByLabel('Version:').selectOption(PYTHON_VERSION);
    await page.getByRole('button', { name: 'Start Installation' }).click();
    await expect(page.getByText(/setup completed/i)).toBeVisible({ timeout: 60000 });
  });

  test('3. verify app appears on /sites', async ({ page }) => {
    await page.goto('/sites');
    const row = page.locator('tr', { hasText: DOMAIN });
    await expect(row).toBeVisible();
    await expect(row.getByText(PYTHON_VERSION)).toBeVisible();
    await row.getByText(DOMAIN, { exact: true }).click();
  });

  test('4. verify app is responding', async ({ page }) => {
    const expected = `Hello World from Flask on port ${PORT}!`;
    const url = `https://${DOMAIN}/`;
    await page.goto(url);
    const locator = page.getByText(expected);
    const timeout = 30000;
    const start = Date.now();
    while (Date.now() - start < timeout) {
      if (await locator.isVisible()) break;
      await page.waitForTimeout(1000);
      await page.reload();
    }
    await expect(locator).toBeVisible();
    console.log('Python PM2 autoinstaller is fully working');
  });

  // TODO: cover manager actions
});
