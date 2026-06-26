import { test, expect } from '@playwright/test';

// NOTE: verify-only -- enabling/disabling DNS clustering and adding a cluster node are
// real, hard-to-revert-safely actions affecting DNS sync across servers. We only check
// that the UI renders correctly for whichever state the environment is currently in.

test('dns cluster page renders enabled or disabled branch', async ({ page }) => {
  await page.goto('/domains/dns-cluster');
  await expect(page).toHaveURL(/domains\/dns-cluster/);

  await expect(page.locator('#tour-dns-cluster-toggle')).toBeVisible();

  const docsLink = page.getByRole('link', { name: 'View Documentation' });
  if (await docsLink.isVisible().catch(() => false)) {
    await expect(page.locator('#tour-dns-cluster-toggle')).toHaveText('Enable DNS Clustering');
    console.log('dns cluster: currently disabled, docs link and enable button shown');
  } else {
    await expect(page.locator('#tour-dns-cluster-toggle')).toHaveText('Disable DNS Clustering');
    await expect(page.locator('#exiting_users')).toBeVisible();
    console.log('dns cluster: currently enabled, cluster nodes table shown');
  }
});

test('add server form validates IPv4 format client-side', async ({ page }) => {
  await page.goto('/domains/dns-cluster');

  const addServerLink = page.getByText('Add Server', { exact: true });
  const visible = await addServerLink.isVisible().catch(() => false);
  test.skip(!visible, 'DNS clustering not currently enabled on this environment');

  await addServerLink.click();
  const ipInput = page.locator('#ip');
  await expect(ipInput).toBeVisible();

  await ipInput.fill('not-an-ip');
  await expect(page.locator('#ip-message')).toHaveText('Invalid IPv4 address.', { timeout: 5_000 });
  await expect(page.locator('#submit-btn')).toBeDisabled();

  console.log('add server form correctly rejects invalid IPv4 input (not submitted)');
});

test('search filters cluster nodes table', async ({ page }) => {
  await page.goto('/domains/dns-cluster');

  const rows = page.locator('#exiting_users tbody tr');
  const count = await rows.count();
  test.skip(count === 0, 'No cluster nodes to search (clustering disabled or no nodes added)');

  const firstIp = (await rows.first().locator('td').first().innerText()).trim();
  await page.locator('input[x-model="searchQuery"]').fill(firstIp);
  await page.waitForTimeout(150);

  await expect(rows.filter({ hasText: firstIp })).toBeVisible();
  console.log(`search filtered cluster nodes table to "${firstIp}"`);
});
