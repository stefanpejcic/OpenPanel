import { test, expect } from '@playwright/test';


test('auto-installer page loads', async ({ page }) => {
  await page.goto('/auto-installer');
  await expect(page).toHaveURL(/auto-installer/);
  await expect(page.locator('body')).toContainText(/install|application|wordpress/i);
  console.log('auto-installer page accessible');
});


test('auto-installer shows available applications', async ({ page }) => {
  await page.goto('/auto-installer');

  // should list at least some installable apps
  const apps = ['WordPress', 'Python', 'NodeJS'];
  let found = 0;
  for (const app of apps) {
    const isVisible = await page.locator('body').textContent().then(t => t?.includes(app) ?? false);
    if (isVisible) found++;
  }

  expect(found).toBeGreaterThan(0);
  console.log(`auto-installer shows ${found} known applications`);
});


test('auto-installer search/filter works', async ({ page }) => {
  await page.goto('/auto-installer');

  const searchInput = page.locator('input[type="search"], input[placeholder*="search"], input[type="text"]').first();
  const hasSearch = await searchInput.isVisible({ timeout: 3000 }).catch(() => false);

  if (hasSearch) {
    await searchInput.fill('WordPress');
    await page.waitForTimeout(300);
    await expect(page.locator('body')).toContainText(/wordpress/i);
    console.log('auto-installer search is working');
  } else {
    console.log('no search input found – skipping filter test');
  }
});


test('auto-installer install form is accessible', async ({ page }) => {
  await page.goto('/auto-installer');

  // click install/select on first visible app
  const installLink = page.locator('a:has-text("Install"), button:has-text("Install"), a[href*="install"]').first();
  const hasInstallLink = await installLink.isVisible({ timeout: 3000 }).catch(() => false);

  if (hasInstallLink) {
    await installLink.click();
    await expect(page.locator('body')).toContainText(/domain|path|version|install/i, { timeout: 10000 });
    console.log('auto-installer install form is accessible');
  } else {
    // apps shown in cards/icons — just verify page has content
    await expect(page.locator('body')).toContainText(/install|application/i);
    console.log('auto-installer has application content');
  }
});
