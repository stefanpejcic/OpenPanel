// TODO: /json/combined_activity

import { test, expect } from '@playwright/test';

async function navigateToDashboardPage(page: any) {
  await page.goto(`/dashboard`);
  await expect(page).toHaveURL(/dashboard/);
}

test('access dashboard', async ({ page }) => {
  await navigateToDashboardPage(page);
  await expect(page.getByText(/welcome/i)).toBeVisible();
  console.log('Dashboard page is accessible');
});



test('sidebar', async ({ page }) => {
  await navigateToDashboardPage(page);

  const html = page.locator('html');
  const sidebar = page.locator('#sidebar');

  await expect(html).toHaveClass(/sidebar-open/);
  await expect(sidebar).toBeVisible();

  await page.locator('#sidebar_toggle_button').click();
  await expect(html).not.toHaveClass(/sidebar-open/);
  await expect(sidebar).not.toBeVisible();

  await page.locator('#sidebar_toggle_button').click();
  await expect(html).toHaveClass(/sidebar-open/);
  await expect(sidebar).toBeVisible();

  console.log('sidebar toggle working');
});


test('dark mode', async ({ page }) => {
  await navigateToDashboardPage(page);

  await page.locator('#user-btn-info').click();

  await expect(page.locator('#theme-toggle-dark-icon')).toBeVisible();

  await page.locator('#theme-toggle-dark-icon').click();
  await expect(page.locator('#theme-toggle-light-icon')).toBeVisible();
  await expect(page.locator('html')).toHaveClass(/dark/);

  await page.locator('#theme-toggle-light-icon').click();
  await expect(page.locator('#theme-toggle-dark-icon')).toBeVisible();
  await expect(page.locator('html')).not.toHaveClass(/dark/);

  console.log('Dark mode switch is working');
});



test('/json/system', async ({ page }) => {
  await page.goto(`/json/system`);
  await expect(page).toHaveURL(/json\/system/);

  const raw = await page.locator('body').innerText();
  const data = JSON.parse(raw);

  const requiredKeys = [
    'cpu', 'hostname', 'kernel', 'openpanel_version',
    'os', 'package_updates', 'running_processes', 'time', 'uptime'
  ];
  for (const key of requiredKeys) {
    expect(data).toHaveProperty(key);
  }

  expect(typeof data.cpu).toBe('string');
  expect(typeof data.hostname).toBe('string');
  expect(typeof data.kernel).toBe('string');
  expect(typeof data.openpanel_version).toBe('string');
  expect(typeof data.os).toBe('string');
  expect(typeof data.package_updates).toBe('string');
  expect(typeof data.running_processes).toBe('string');
  expect(typeof data.time).toBe('string');
  expect(typeof data.uptime).toBe('string'); // or number

  expect(data.cpu.length).toBeGreaterThan(0);
  expect(data.hostname.length).toBeGreaterThan(0);
  expect(Number(data.running_processes)).toBeGreaterThan(0);
  expect(Number(data.uptime)).toBeGreaterThanOrEqual(0);
  expect(Number(data.package_updates)).toBeGreaterThanOrEqual(0);
  expect(data.openpanel_version).toMatch(/^\d+\.\d+\.\d+(-[a-z0-9.]+)?$/i); //beta
  expect(data.time).toMatch(/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/);
  expect(data.os.toLowerCase()).toMatch(/ubuntu|debian|almalinux|rockylinux|centos/);

  console.log('system info is available:', data);
});


test('/json/cpu', async ({ page }) => {
  await page.goto(`/json/cpu`);
  await expect(page).toHaveURL(/json\/cpu/);

  const data = JSON.parse(await page.locator('body').innerText());

  const coreKeys = Object.keys(data);
  expect(coreKeys.length).toBeGreaterThan(0);

  for (const key of coreKeys) {
    expect(key).toMatch(/^core_\d+$/);
  }

  for (const [key, value] of Object.entries(data)) {
    expect(typeof value).toBe('number');
    expect(value).toBeGreaterThanOrEqual(0);
    expect(value).toBeLessThanOrEqual(100);
  }

  const indices = coreKeys.map(k => parseInt(k.split('_')[1])).sort((a, b) => a - b);
  indices.forEach((idx, i) => expect(idx).toBe(i));

  console.log(`cpu info available: ${coreKeys.length} cores`);
});


test('/json/memory', async ({ page }) => {
  await page.goto(`/json/memory`);
  await expect(page).toHaveURL(/json\/memory/);

  const data = JSON.parse(await page.locator('body').innerText());

  expect(data).toHaveProperty('human_readable');
  expect(data).toHaveProperty('ram_info');
  expect(data).toHaveProperty('swap_info');

  const { ram, swap } = data.human_readable;
  for (const key of ['available', 'percent', 'total', 'used']) {
    expect(typeof ram[key]).toBe('string');
    expect(ram[key].length).toBeGreaterThan(0);
  }
  expect(ram.percent).toMatch(/^\d+(\.\d+)?%$/);
  expect(ram.available).toMatch(/[\d.]+ (B|KB|MB|GB|TB)/);
  expect(ram.total).toMatch(/[\d.]+ (B|KB|MB|GB|TB)/);
  expect(ram.used).toMatch(/[\d.]+ (B|KB|MB|GB|TB)/);

  for (const key of ['free', 'percent', 'total', 'used']) {
    expect(typeof swap[key]).toBe('string');
    expect(swap[key].length).toBeGreaterThan(0);
  }
  expect(swap.percent).toMatch(/^\d+(\.\d+)?%$/);

  const { ram_info } = data;
  for (const key of ['available', 'total', 'used']) {
    expect(typeof ram_info[key]).toBe('number');
    expect(ram_info[key]).toBeGreaterThan(0);
  }
  expect(ram_info.percent).toBeGreaterThanOrEqual(0);
  expect(ram_info.percent).toBeLessThanOrEqual(100);

  expect(ram_info.used + ram_info.available).toBeLessThanOrEqual(ram_info.total * 1.01);

  const { swap_info } = data;
  for (const key of ['free', 'total', 'used', 'sin', 'sout']) {
    expect(typeof swap_info[key]).toBe('number');
    expect(swap_info[key]).toBeGreaterThanOrEqual(0);
  }
  expect(swap_info.percent).toBeGreaterThanOrEqual(0);
  expect(swap_info.percent).toBeLessThanOrEqual(100);

  expect(swap_info.used + swap_info.free).toBeLessThanOrEqual(swap_info.total * 1.01);

  console.log(`memory info available — RAM: ${ram.used}/${ram.total} (${ram.percent})`);
});


test('/json/load', async ({ page }) => {
  await page.goto(`/json/load`);
  await expect(page).toHaveURL(/json\/load/);

  const data = JSON.parse(await page.locator('body').innerText());

  for (const key of ['load1min', 'load5min', 'load15min']) {
    expect(data).toHaveProperty(key);
  }

  for (const key of ['load1min', 'load5min', 'load15min']) {
    expect(typeof data[key]).toBe('number');
    expect(data[key]).toBeGreaterThanOrEqual(0);
  }

  const MAX_REASONABLE_LOAD = 256;
  for (const key of ['load1min', 'load5min', 'load15min']) {
    expect(data[key]).toBeLessThan(MAX_REASONABLE_LOAD);
  }

  console.log(`load average — 1m: ${data.load1min}, 5m: ${data.load5min}, 15m: ${data.load15min}`);
});


test('/json/disk', async ({ page }) => {
  await page.goto(`/json/disk`);
  await expect(page).toHaveURL(/json\/disk/);

  const data = JSON.parse(await page.locator('body').innerText());

  expect(Array.isArray(data)).toBe(true);
  expect(data.length).toBeGreaterThan(0);

  for (const disk of data) {
    for (const key of ['device', 'free', 'fstype', 'mountpoint', 'percent', 'total', 'used']) {
      expect(disk).toHaveProperty(key);
    }

    expect(typeof disk.device).toBe('string');
    expect(disk.device.length).toBeGreaterThan(0);
    expect(typeof disk.fstype).toBe('string');
    expect(disk.fstype.length).toBeGreaterThan(0);
    expect(typeof disk.mountpoint).toBe('string');
    expect(disk.mountpoint.length).toBeGreaterThan(0);

    for (const key of ['free', 'total', 'used']) {
      expect(typeof disk[key]).toBe('number');
      expect(disk[key]).toBeGreaterThanOrEqual(0);
    }

    expect(typeof disk.percent).toBe('number');
    expect(disk.percent).toBeGreaterThanOrEqual(0);
    expect(disk.percent).toBeLessThanOrEqual(100);

    expect(disk.used + disk.free).toBeLessThanOrEqual(disk.total * 1.01);

    const mountpoints = data.map(d => d.mountpoint);
    expect(mountpoints).toContain('/');
  }

  console.log(`disk info available: ${data.length} partition(s)`);
});


