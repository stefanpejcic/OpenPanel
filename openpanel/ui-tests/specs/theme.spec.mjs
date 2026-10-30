import { test, expect } from '@playwright/test';
import { readdirSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { watch, overflow, shoot } from './checks.mjs';

const dir = fileURLToPath(new URL('../../themes/modern/', import.meta.url));
const pages = readdirSync(dir).filter((f) => f.endsWith('.html'));

// the theme pages are the design sign-off, so every page has to pass at every size
for (const file of pages) {
  test(`theme ${file}`, async ({ page }, testInfo) => {
    const seen = watch(page);
    await page.goto(`/themes/modern/${file}`);
    await page.waitForLoadState('networkidle');
    expect(await overflow(page), 'page scrolls sideways').toBeNull();
    expect(seen.errors, 'console errors').toEqual([]);
    expect(seen.failedRequests, 'missing files').toEqual([]);
    await shoot(page, testInfo, 'theme', file.replace('.html', ''));
  });
}

test('theme phone drawer opens and closes', async ({ page }, testInfo) => {
  test.skip(!testInfo.project.name.startsWith('phone'), 'drawer only exists on phones and tablets');
  await page.goto('/themes/modern/domains.html');
  const sidebar = page.locator('.op-sidebar');
  await page.getByRole('button', { name: 'Open menu' }).click();
  await expect(sidebar).toBeInViewport();
  await page.keyboard.press('Escape');
  await expect(sidebar).not.toBeInViewport();
});

test('theme tool search filters the Tools view', async ({ page }) => {
  await page.goto('/themes/modern/dashboard.html?view=tools');
  await page.getByRole('searchbox', { name: 'Find a tool' }).fill('php');
  // PHP Versions, PHP Options, PHP Extensions, php.ini Editor and phpMyAdmin
  await expect(page.locator('.op-tool:visible')).toHaveCount(5);
  await page.getByRole('searchbox', { name: 'Find a tool' }).fill('zzz');
  await expect(page.getByText('No tool matches')).toBeVisible();
});
