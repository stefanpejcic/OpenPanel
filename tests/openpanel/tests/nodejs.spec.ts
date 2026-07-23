import { test, expect } from '@playwright/test';

// https://github.com/stefanpejcic/nodejs-helloworld

const DOMAIN = 'nodejs.tests.openpanel.org';
const APP_NAME = 'nodeaplikacija';
const PORT = '3000';
const NODE_VERSION = '25.9.0';
const STARTUP_FILE = `/var/www/html/${DOMAIN}/app.js`;

const PACKAGE_JSON = `{
  "name": "helloworld-node",
  "version": "1.0.0",
  "description": "Simple Node.js Hello World app using Express",
  "main": "app.js",
  "scripts": {
    "start": "node app.js"
  },
  "author": "Stefan Pejcic",
  "license": "MIT",
  "dependencies": {
    "express": "^4.18.2"
  }
}`;

const APP_JS = `const express = require('express');
const app = express();
const port = 3000;
app.get('/', (req, res) => {
  res.send(\`Hello World from Node.js \${process.version} on port \${port}!\`);
});
app.listen(port, () => {
  console.log(\`Server is running at http://localhost:\${port}\`);
});`;

test.describe.configure({ mode: 'serial' });

test.describe('Node.js autoinstaller', () => {

  test('1. create app files', async ({ page }) => {
    // package.json
    await page.goto(`/file-manager/edit-file/${DOMAIN}/package.json?editor=text&new=true`);
    await page.locator('#editor-text').fill(PACKAGE_JSON);
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText(/saved|success/i).first()).toBeVisible();

    // app.js
    await page.goto(`/file-manager/edit-file/${DOMAIN}/app.js?editor=text&new=true`);
    await page.locator('#editor-text').fill(APP_JS);
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByText(/saved|success/i).first()).toBeVisible();
  });

  test('2. install app', async ({ page }) => {
    await page.goto('/nodejs/install');

    await page.locator('#service_name').fill(APP_NAME);
    await page.locator('#port').fill(PORT);

    await page.locator('#domain_id').selectOption({ label: DOMAIN });

    await page.locator('#startup_file').fill(STARTUP_FILE);

    await page.locator('#version').selectOption(NODE_VERSION);

    await page.locator('#installButton').click();

    await expect(page.getByText(/setup completed/i)).toBeVisible({ timeout: 120000 });
  });

  test('3. verify app appears on /sites', async ({ page }) => {
    await page.goto('/sites');

    const row = page.getByRole('row').filter({ hasText: DOMAIN });
    await expect(row).toBeVisible();
    await expect(row.getByText(NODE_VERSION)).toBeVisible();
  
    await row.getByRole('link', { name: 'Manage', exact: true }).click();
    await expect(page).toHaveURL(`/website?domain=${DOMAIN}`);
  });

  test('4. verify app is responding', async ({ page }) => {
    // process.version returns e.g. "v25.9.0", strip the leading "v" if version lacks it
    const nodeVersion = NODE_VERSION.startsWith('v') ? NODE_VERSION : `v${NODE_VERSION}`;
    const expected = `Hello World from Node.js ${nodeVersion} on port ${PORT}!`;
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
    console.log('Node.js autoinstaller is fully working');
  });

  // TODO: cover manager actions

});
