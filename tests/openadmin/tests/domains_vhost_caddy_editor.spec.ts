import { test, expect } from '@playwright/test';

async function firstDomainEditorUrl(page, attr: 'vhost' | 'caddy') {
  await page.goto('/domains');
  const link = page.locator(`a[href^="/domains/${attr}/"]`).first();
  const count = await link.count();
  if (count === 0) return null;
  return link.getAttribute('href');
}

test('vhost editor loads, edits, saves, and reverts', async ({ page }) => {
  const href = await firstDomainEditorUrl(page, 'vhost');
  test.skip(!href, 'No domains available to inspect VirtualHost for');

  await page.goto(href!);
  await expect(page.getByText('Edit Domain VirtualHost')).toBeVisible();

  const textarea = page.locator('#bind_content');
  await expect(textarea).toBeVisible();
  const original = await textarea.inputValue();
  const comment = `\n# test-comment-${Date.now()}`;

  await textarea.fill(original + comment);
  await page.getByRole('button', { name: 'Save' }).click();
  await page.waitForLoadState();
  await expect(page.locator('#bind_content')).toHaveValue(new RegExp(comment.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')));

  // revert
  await page.locator('#bind_content').fill(original);
  await page.getByRole('button', { name: 'Save' }).click();
  await page.waitForLoadState();
  await expect(page.locator('#bind_content')).toHaveValue(original);

  console.log(`vhost editor (${href}) edited, saved, and reverted`);
});

test('vhost editor domain selector page loads when no domain in URL', async ({ page }) => {
  await page.goto('/domains/vhost');
  await expect(page).toHaveURL(/domains\/vhost/);
  await expect(page.getByText('Select domain to edit VirtualHosts file for')).toBeVisible();

  console.log('vhost editor domain selector loaded');
});

test('caddyfile editor loads with warning banner', async ({ page }) => {
  // NOTE: verify-only on save -- the page itself warns that misconfiguring the
  // Caddyfile directly "may cause the entire web server and all hosted websites to go
  // down," so we don't submit edits here (use the VirtualHost editor test for an
  // exercised save/revert round-trip instead).
  const href = await firstDomainEditorUrl(page, 'caddy');
  test.skip(!href, 'No domains available to inspect Caddyfile for');

  await page.goto(href!);
  await expect(page.getByText('Edit Domain Caddyfile')).toBeVisible();
  await expect(page.locator('#bind_content')).not.toHaveValue('');
  await expect(page.getByText('Directly editing the Caddyfile is not recommended')).toBeVisible();

  console.log(`caddyfile editor (${href}) loaded with warning banner, not edited`);
});

test('caddyfile editor domain selector page loads when no domain in URL', async ({ page }) => {
  await page.goto('/domains/caddy');
  await expect(page).toHaveURL(/domains\/caddy/);
  await expect(page.getByText('Select domain to edit Caddyfile for')).toBeVisible();

  console.log('caddyfile editor domain selector loaded');
});
