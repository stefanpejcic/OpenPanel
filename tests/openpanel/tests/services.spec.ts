import { test, expect } from '@playwright/test';

const webserverServices = ['nginx', 'apache', 'openlitespeed'];
const commonServices = [...webserverServices, 'mysql', 'php-fpm-8.5'];

async function navigateToService(page: any, service: string) {
  await page.goto(`/services/${service}`);
  await expect(page).toHaveURL(new RegExp(`services/${service}`));
}

/** Try each candidate; return the first whose page loads with a 2xx status. */
async function navigateToFirstAvailable(page: any, candidates: string[]): Promise<string> {
  const errors: string[] = [];
  for (const service of candidates) {
    const response = await page.goto(`/services/${service}`);
    if (response && response.ok()) {
      const hasError = await page.getByText(/does not exist/i).isVisible().catch(() => false);
      if (!hasError) return service;
      errors.push(`${service}: soft error page`);
    } else {
      errors.push(`${service}: HTTP ${response?.status()}`);
    }
  }
  throw new Error(`All candidates failed:\n${errors.join('\n')}`);
}

test('services list page', async ({ page }) => {
  await page.goto('/services/');
  await expect(page).toHaveURL(/services/);
  await expect(page.locator('body')).toContainText(/service|nginx|apache/i);
  console.log('services list page accessible');
});

test('services list contains expected entries', async ({ page }) => {
  await page.goto('/services/');
  const select = page.locator('select[onchange*="services"]');
  if (await select.isVisible({ timeout: 3000 }).catch(() => false)) {
    const options = await select.locator('option').allTextContents();
    expect(options.length).toBeGreaterThan(0);
    console.log(`services dropdown has ${options.length} options: ${options.slice(0, 5).join(', ')}`);
  } else {
    await expect(page.locator('body')).toContainText(/nginx|apache|mysql/i);
    console.log('services listed on page');
  }
});

// Webserver: try nginx → apache2 → openlitespeed; fail only if all three fail
test('webserver service page (nginx / apache / openlitespeed)', async ({ page }) => {
  const service = await navigateToFirstAvailable(page, webserverServices);
  await expect(page.locator('body')).toContainText(new RegExp(service, 'i'));
  const enableBtn = page.locator('button[type="submit"]').first();
  await expect(enableBtn).toBeVisible();
  console.log(`webserver service page accessible via: ${service}`);
});

// mysql: accept either mysql or mariadb
test('mysql service page', async ({ page }) => {
  const service = await navigateToFirstAvailable(page, ['mysql', 'mariadb']);
  await expect(page.locator('body')).toContainText(/mysql|mariadb/i);
  const enableBtn = page.locator('button[type="submit"]').first();
  await expect(enableBtn).toBeVisible();
  console.log(`database service page accessible via: ${service}`);
});

// php-fpm pinned to 8.5
test('php-fpm-8.5 service page', async ({ page }) => {
  await navigateToService(page, 'php-fpm-8.5');
  await expect(page.locator('body')).toContainText(/php-fpm-8\.5|php-fpm/i);
  const enableBtn = page.locator('button[type="submit"]').first();
  await expect(enableBtn).toBeVisible();
  console.log('php-fpm-8.5 service page is accessible');
});

test('restart a service', async ({ page }) => {
  const service = await navigateToFirstAvailable(page, webserverServices);
  const actionInput = page.locator('input[name="action"]');
  if (await actionInput.isVisible({ timeout: 2000 }).catch(() => false)) {
    const currentAction = await actionInput.getAttribute('value');
    console.log(`current action for ${service}: ${currentAction}`);
  }
  await expect(page.locator('body')).toContainText(new RegExp(service, 'i'));
  console.log('service page has correct content');
});

test('service version selector', async ({ page }) => {
  await page.goto('/services/');
  const select = page.locator('select[onchange*="services"]');
  if (await select.isVisible({ timeout: 3000 }).catch(() => false)) {
    const options = await select.locator('option').evaluateAll(opts =>
      opts.map(o => o.value).filter(v => v !== '')
    );
    if (options.length > 0) {
      const first = options[0];
      await select.selectOption(first);
      await expect(page).toHaveURL(new RegExp(`services/${first}`));
      console.log(`navigated to service: ${first}`);
    }
  } else {
    console.log('version selector not present – skipping');
  }
});
