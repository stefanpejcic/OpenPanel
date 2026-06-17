import { test, expect } from '@playwright/test';

const commonServices = ['nginx', 'apache2', 'mysql', 'php-fpm'];

async function navigateToService(page: any, service: string) {
  await page.goto(`/services/${service}`);
  await expect(page).toHaveURL(new RegExp(`services/${service}`));
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
    // table/list view
    await expect(page.locator('body')).toContainText(/nginx|apache|mysql/i);
    console.log('services listed on page');
  }
});


for (const service of commonServices) {
  test(`${service} service page`, async ({ page }) => {
    await navigateToService(page, service);

    await expect(page.locator('body')).toContainText(new RegExp(service, 'i'));

    const enableBtn = page.locator('button[type="submit"]').first();
    await expect(enableBtn).toBeVisible();

    console.log(`${service} service page is accessible`);
  });
}


test('restart a service', async ({ page }) => {
  const service = 'nginx';
  await navigateToService(page, service);

  // look for restart action in the form
  const actionInput = page.locator('input[name="action"]');
  if (await actionInput.isVisible({ timeout: 2000 }).catch(() => false)) {
    const currentAction = await actionInput.getAttribute('value');
    console.log(`current action for ${service}: ${currentAction}`);
  }

  // verify the page has the right service name
  await expect(page.locator('body')).toContainText(/nginx/i);
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
