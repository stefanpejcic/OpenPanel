import { test, expect } from '@playwright/test';
import * as ftp from 'basic-ftp';
import { Readable, Writable } from 'stream';

// npm install basic-ftp

const FTP_USER = 'ftp';
const FTP_DOMAIN = 'files.tests.openpanel.org';
const FTP_PASS = 'b&tK3C9+cncXl%Ut';
const FTP_NEW_PASS = 'N3wP@ssw0rd!xyz';
const FTP_PATH = '/var/www/html/files.tests.openpanel.org';

let ftpHost: string;

async function resolveFtpHost(page: any): Promise<string> {
  if (!ftpHost) {
    await page.goto('/ftp');
    ftpHost = (await page.locator('#ftp_server_address').textContent())?.trim() ?? '';
  }
  return ftpHost;
}

async function ftpConnect(host: string, password: string): Promise<ftp.Client> {
  const client = new ftp.Client();
  client.ftp.verbose = true;
  await client.access({
    host,
    port: 21,
    user: `${FTP_USER}@${FTP_DOMAIN}`,
    password,
    secure: false,
  });
  return client;
}

async function getFTPCount(page: Page): Promise<number> {
  const text = await page.locator('#dashboard_usage_ftp').locator('p').nth(1).textContent();
  if (!text) throw new Error('Cannot read ftp count');

  const match = text.match(/(\d+)\s*\//);
  if (!match) throw new Error(`Cannot parse ftp count from: ${text}`);

  return parseInt(match[1], 10);
}

test('create account', async ({ page }) => {
  // 1. check current user count
  await page.goto('/dashboard');
  const initialCount = await getFTPCount(page);
  let expectedCount = initialCount;

  // 1. create account
  await page.goto('/ftp/new');
  await expect(page.getByRole('heading', { name: 'New FTP Account' })).toBeVisible();
  await page.locator('#new_ftp_username').fill(FTP_USER);
  await page.locator('#domain').selectOption({ label: FTP_DOMAIN });
  await page.locator('#password').fill(FTP_PASS);
  await page.locator('#new_user_path').fill(FTP_PATH);
  await page.getByRole('button', { name: /Create Account/i }).click();
  await expect(page.getByText(/created successfully/i)).toBeVisible();

  const row = page.locator('tbody tr').filter({ hasText: `${FTP_USER}@${FTP_DOMAIN}` });
  await expect(row).toBeVisible();
  await expect(row).toContainText(`${FTP_USER}@${FTP_DOMAIN}`);
  await expect(row).toContainText(FTP_PATH);

  ftpHost = (await page.locator('#ftp_server_address').textContent())?.trim() ?? '';
  const ftpPort = (await page.locator('#ftp_server_port').textContent())?.trim();
  expect(ftpHost).toBeTruthy();
  expect(ftpPort).toBe('21');

  expectedCount++;

  console.log(`Account created, logins:`);
  console.log(`USERNAME: ${FTP_USER}@${FTP_DOMAIN}`);
  console.log(`PASSWORD: ${FTP_PASS}`);
  console.log(`SERVER:   ${ftpHost}`);
  console.log(`PORT:     ${ftpPort}`);

  // 3. check count again
  await page.goto('/dashboard');
  await expect.poll(async () => {return await getFTPCount(page);}).toBe(expectedCount);
});

test('login, upload, list, download, delete', async ({ page }) => {
  const host = await resolveFtpHost(page);
  expect(host).toBeTruthy();

  const client = await ftpConnect(host, FTP_PASS);
  try {
    const testContent = Buffer.from('playwright-ftp-test');
    await client.uploadFrom(Readable.from(testContent), 'pw_test.txt');

    const list = await client.list();
    const found = list.find(f => f.name === 'pw_test.txt');
    expect(found).toBeTruthy();
    expect(found!.size).toBe(testContent.byteLength);
    console.log('ftp upload is working');

    const chunks: Buffer[] = [];
    const sink = new Writable({ write(chunk, _enc, cb) { chunks.push(chunk); cb(); } });
    await client.downloadTo(sink, 'pw_test.txt');
    expect(Buffer.concat(chunks).toString()).toBe('playwright-ftp-test');
    console.log('ftp download is working');

    await client.remove('pw_test.txt');
    const listAfter = await client.list();
    expect(listAfter.find(f => f.name === 'pw_test.txt')).toBeUndefined();
    console.log('ftp delete is working');
  } finally {
    client.close();
  }
});


//test('connection list', async ({ page }) => {
//  const host = await resolveFtpHost(page);
//  expect(host).toBeTruthy();
//
//  const client = await ftpConnect(host, FTP_PASS);
//  const keepAlive = setInterval(async () => {
//    try { await client.pwd(); } catch { /* ignore */ }
//  }, 3000);
//
//  try {
//    await page.goto('/ftp/connections');
//    const pre = page.locator('pre');
//    await expect(pre).toBeVisible();
//    await expect(pre).toContainText(FTP_USER);
//    console.log('ftp connection list is working');
//  } finally {
//    clearInterval(keepAlive);
//    client.close();
//  }
//});



test('password change', async ({ page }) => {
  const host = await resolveFtpHost(page);
  expect(host).toBeTruthy();

  await page.goto(`/ftp/password/${FTP_USER}@${FTP_DOMAIN}`);
  await page.locator('#new_password').fill(FTP_NEW_PASS);
  await page.locator('#confirm_password').fill(FTP_NEW_PASS);
  await page.getByRole('button', { name: /Update/i }).click();
  await expect(page.getByText(/password changed successfully/i)).toBeVisible();
  console.log('ftp password change is working');

  // Verify old password no longer works
  const oldClient = new ftp.Client();
  oldClient.ftp.verbose = false;
  let oldFailed = false;
  try {
    await oldClient.access({ host, port: 21, user: `${FTP_USER}@${FTP_DOMAIN}`, password: FTP_PASS, secure: false });
  } catch {
    oldFailed = true;
  } finally {
    oldClient.close();
  }
  expect(oldFailed, 'old password should be rejected').toBe(true);

  // Verify new password works
  const client = await ftpConnect(host, FTP_NEW_PASS);
  try {
    const list = await client.list();
    expect(Array.isArray(list)).toBe(true);
    console.log('ftp new password login is working');
  } finally {
    client.close();
  }
});

test('path change', async ({ page }) => {
  const host = await resolveFtpHost(page);
  expect(host).toBeTruthy();

  const newPath = '/var/www/html/';

  await page.goto(`/ftp/path/${FTP_USER}@${FTP_DOMAIN}`);
  await page.locator('#new_path').fill(newPath);
  await page.getByRole('button', { name: /Update/i }).click();
  await expect(page.getByText(/path changed successfully/i)).toBeVisible();
  console.log('ftp path change is working');

  // Verify new path shown in table
  await page.goto('/ftp');
  const row = page.locator('tbody tr').filter({ hasText: `${FTP_USER}@${FTP_DOMAIN}` });
  await expect(row).toBeVisible();
  await expect(row).toContainText(newPath);
  await expect(row).not.toContainText(FTP_PATH);
  console.log('ftp path updated in table');

  // Verify FTP connection still works and lands in new path
  const client = await ftpConnect(host, FTP_NEW_PASS);
  try {
      const list = await client.list();
      const targetExists = list.some(item => item.name === 'files.tests.openpanel.org');
      expect(targetExists).toBe(true);
      console.log('ftp new path confirmed via directory contents listing');
    } finally {
      client.close();
    }
});



test('filezilla config', async ({ page }) => {
  const host = await resolveFtpHost(page);
  expect(host).toBeTruthy();

  const response = await page.request.get(`/ftp/configuration/filezilla/${FTP_USER}@${FTP_DOMAIN}`);
  expect(response.ok()).toBeTruthy();

  const xml = await response.text();

  expect(xml).toMatch(/^<\?xml/);
  expect(xml).toContain('<FileZilla3>');
  expect(xml).toContain(`<Host>${host}</Host>`);
  expect(xml).toContain('<Port>21</Port>');
  expect(xml).toContain(`<User>${FTP_USER}@${FTP_DOMAIN}</User>`);
  expect(xml).toContain(`<Name>${FTP_USER}@${FTP_DOMAIN}</Name>`);

  console.log('filezilla config is valid and contains correct connection info');
});

test('cyberduck config', async ({ page }) => {
  const host = await resolveFtpHost(page);
  expect(host).toBeTruthy();

  const response = await page.request.get(`/ftp/configuration/cyberduck/${FTP_USER}@${FTP_DOMAIN}`);
  expect(response.ok()).toBeTruthy();

  const xml = await response.text();

  expect(xml).toMatch(/^<\?xml/);
  expect(xml).toContain('<bookmark>');
  expect(xml).toContain(`<hostname>${host}</hostname>`);
  expect(xml).toContain(`<username>${FTP_USER}@${FTP_DOMAIN}</username>`);
  expect(xml).toContain('<protocol>ftp</protocol>');

  console.log('cyberduck config is valid and contains correct connection info');
});


test('search', async ({ page }) => {
  await page.goto('/ftp');

  const searchInput = page.locator('input[x-model="searchQuery"]');
  const row = page.locator('tbody tr').filter({ hasText: `${FTP_USER}@${FTP_DOMAIN}` });
  await expect(row).toBeVisible();

  // Search for our user
  await searchInput.fill(`${FTP_USER}@${FTP_DOMAIN}`);
  await expect(row).toBeVisible();

  // Only one row should be visible
  const visibleRows = page.locator('tbody tr').filter({ visible: true });
  await expect(visibleRows).toHaveCount(1);

  // Search for something else
  await searchInput.fill('zzznomatchzzz');
  await expect(row).not.toBeVisible();

  console.log('ftp account search is working');
});


test('account delete', async ({ page }) => {
  // 1. check count
  await page.goto('/dashboard');
  const initialCount = await getFTPCount(page);
  let expectedCount = initialCount;

  // 2. delete
  const host = await resolveFtpHost(page);
  expect(host).toBeTruthy();

  await page.goto('/ftp');
  const row = page.locator('tbody tr').filter({ hasText: `${FTP_USER}@${FTP_DOMAIN}` });
  await expect(row).toBeVisible();

  await row.getByRole('button', { name: /delete/i }).click();
  await page.getByRole('button', { name: /confirm/i }).click();
  await expect(page.getByText(/deleted successfully/i)).toBeVisible();
  console.log('ftp account delete is working');

  await expect(row).not.toBeVisible();
  expectedCount--;

  // 3. test connection
  const client = new ftp.Client();
  client.ftp.verbose = false;
  let rejected = false;
  try {
    await client.access({ host, port: 21, user: `${FTP_USER}@${FTP_DOMAIN}`, password: FTP_NEW_PASS, secure: false });
  } catch {
    rejected = true;
  } finally {
    client.close();
  }
  expect(rejected, 'deleted account should be rejected by FTP server').toBe(true);
  console.log('ftp deleted account correctly rejected');

  // 4. check count again
  await page.goto('/dashboard');
  await expect.poll(async () => {return await getFTPCount(page);}).toBe(expectedCount);
});
