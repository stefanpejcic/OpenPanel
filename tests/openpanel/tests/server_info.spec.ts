import { test, expect } from '@playwright/test';

test('server info page', async ({ page }) => {
  const apiResponses: Record<string, unknown> = {};

  const captureResponse = async (url: string, key: string) => {
    const response = await page.waitForResponse(
      (res) => res.url().includes(url) && res.status() === 200
    );
    apiResponses[key] = await response.json();
  };

  await page.goto(`/server/info`);
  await expect(page).toHaveURL(/server\/info/);

  const [planData, portsData, infoData] = await Promise.all([
    page.waitForResponse((res) => res.url().includes('/json/system/hosting/plan') && res.status() === 200).then((res) => res.json()),
    page.waitForResponse((res) => res.url().includes('/json/system/hosting/ports') && res.status() === 200).then((res) => res.json()),
    page.waitForResponse((res) => res.url().includes('/json/system/hosting/info') && res.status() === 200).then((res) => res.json()),
  ]);

  // 1. /json/system/hosting/plan

  for (const field of ['context', 'ns1', 'ns2', 'ns3', 'ns4', 'plan_description',
    'plan_disk_limit', 'plan_max_email_quota', 'plan_mysql', 'plan_ram_limit', 'plan_webserver']) {
    expect(typeof planData[field], `plan.${field} should be a string`).toBe('string');
  }

  for (const field of ['plan_bandwidth', 'plan_db_limit', 'plan_domains_limit',
    'plan_email_limit', 'plan_ftp_limit', 'plan_inodes_limit', 'plan_websites_limit']) {
    expect(typeof planData[field], `plan.${field} should be a number`).toBe('number');
    expect(planData[field], `plan.${field} should be >= 0`).toBeGreaterThanOrEqual(0);
  }

  expect(typeof planData['plan_cpu_limit'], 'plan.plan_cpu_limit should be a string').toBe('string');
  expect(Number(planData['plan_cpu_limit']), 'plan_cpu_limit should be parseable as a positive number')
    .toBeGreaterThan(0);

  expect(['apache', 'nginx', 'litespeed', 'openlitespeed']).toContain(planData.plan_webserver);
  expect(['mariadb', 'mysql', 'percona']).toContain(planData.plan_mysql);

  // 2. /json/system/hosting/ports

  for (const field of ['pgadmin_port', 'phpmyadmin_port', 'remote_mysql_port', 'remote_postgres_port']) {
    expect(typeof portsData[field], `ports.${field} should be a string`).toBe('string');
    const port = Number(portsData[field]);
    expect(Number.isInteger(port), `ports.${field} should be parseable as an integer`).toBe(true);
    expect(port, `ports.${field} should be a valid port (1–65535)`).toBeGreaterThanOrEqual(1);
    expect(port, `ports.${field} should be a valid port (1–65535)`).toBeLessThanOrEqual(65535);
  }

  // 3. /json/system/hosting/info

  for (const field of ['ip', 'load_avg', 'machine', 'node', 'processor',
    'release', 'system', 'uptime', 'version']) {
    expect(typeof infoData[field], `info.${field} should be a string`).toBe('string');
  }

  expect(infoData.ip, 'info.ip should be a valid IPv4 address')
    .toMatch(/^\d{1,3}(\.\d{1,3}){3}$/);

  expect(infoData.load_avg, 'info.load_avg should be 3 comma-separated numbers')
    .toMatch(/^[\d.]+,\s*[\d.]+,\s*[\d.]+$/);

  expect(infoData.uptime.trim().length, 'info.uptime should not be empty').toBeGreaterThan(0);

  expect(['Linux']).toContain(infoData.system);

  // TODO: compare ui with api
  await expect(page.getByText(infoData.ip)).toBeVisible();
  await expect(page.getByText(planData.plan_description)).toBeVisible();

  console.log('server info has data', { planData, portsData, infoData });
});
