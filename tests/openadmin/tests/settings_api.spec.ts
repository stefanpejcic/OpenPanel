import { test, expect } from '@playwright/test';

test('api settings page renders branch matching current state', async ({ page }) => {
  await page.goto('/settings/api');
  await expect(page).toHaveURL(/settings\/api/);

  const disabledByBasicAuth = page.getByText('API Access is Disabled because BasicAuth is enabled');
  const disabled = page.getByText('API Access Disabled');
  const enabledForm = page.locator('#httpRequestForm');

  if (await disabledByBasicAuth.isVisible().catch(() => false)) {
    console.log('api settings: disabled because basic auth is enabled');
  } else if (await enabledForm.isVisible().catch(() => false)) {
    await expect(page.getByRole('button', { name: 'Disable API access' })).toBeVisible();
    console.log('api settings: API access currently enabled, tester form shown');
  } else {
    await expect(disabled).toBeVisible();
    await expect(page.getByRole('button', { name: 'Enable API access' })).toBeVisible();
    console.log('api settings: API access currently disabled');
  }
});

test('api tester sends a safe GET request and shows response', async ({ page }) => {
  await page.goto('/settings/api');

  const form = page.locator('#httpRequestForm');
  const enabled = await form.isVisible().catch(() => false);
  test.skip(!enabled, 'API access not currently enabled on this environment');

  await page.locator('select[name="method"]').selectOption('GET');
  await page.locator('input[name="url"]').fill('/api/');
  await page.locator('#submit_button').click();

  await expect(page.locator('#responseDisplay')).not.toHaveText('', { timeout: 10_000 });
  console.log('api tester submitted a safe GET /api/ request and displayed a response');
});

test('view examples loads endpoint documentation', async ({ page }) => {
  await page.goto('/settings/api');

  const viewExamplesLink = page.getByText('View Examples');
  const visible = await viewExamplesLink.isVisible().catch(() => false);
  test.skip(!visible, 'API access not currently enabled on this environment');

  await viewExamplesLink.click();
  await expect(page.locator('#api_examples')).not.toHaveText('', { timeout: 10_000 });
  console.log('view examples loaded API endpoint documentation');
});
