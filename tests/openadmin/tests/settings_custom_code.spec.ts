import { test, expect } from '@playwright/test';

test('custom code page loads with save button', async ({ page }) => {
  await page.goto('/settings/custom-code');
  await expect(page).toHaveURL(/settings\/custom-code/);
  await expect(page.locator('#tour-custom-code')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Save Files' })).toBeVisible();

  console.log('custom code page loaded');
});

test('pagespeed api key field edits and saves', async ({ page }) => {
  await page.goto('/settings/custom-code');

  const input = page.locator('#pagespeed_api_key');
  const original = await input.inputValue();
  const testValue = `test-key-${Date.now()}`;

  await input.fill(testValue);
  await page.getByRole('button', { name: 'Save Files' }).click();
  await page.waitForLoadState();
  await expect(page.locator('#pagespeed_api_key')).toHaveValue(testValue);

  // revert
  await page.locator('#pagespeed_api_key').fill(original);
  await page.getByRole('button', { name: 'Save Files' }).click();
  await page.waitForLoadState();
  await expect(page.locator('#pagespeed_api_key')).toHaveValue(original);

  console.log('pagespeed api key edited, saved, and reverted');
});

// Enterprise-only sections (custom_css/js/header/footer/howto_guides/custom_section)
const enterpriseTextareas = ['custom_css', 'custom_js', 'in_header', 'in_footer', 'howto_guides', 'custom_section'];
// available regardless of license
const commonTextareas = ['wp_plugins', 'wp_themes', 'forbidden_usernames', 'restricted_domains', 'post_update', 'pre_startup'];

for (const name of commonTextareas) {
  test(`edit, save, and revert "${name}" textarea`, async ({ page }) => {
    await page.goto('/settings/custom-code');

    const textarea = page.locator(`textarea[name="${name}"]`);
    await expect(textarea).toBeVisible();
    const original = await textarea.inputValue();
    const comment = `\n# test-comment-${Date.now()}`;

    await textarea.fill(original + comment);
    await page.getByRole('button', { name: 'Save Files' }).click();
    await page.waitForLoadState();
    await expect(page.locator(`textarea[name="${name}"]`)).toHaveValue(new RegExp(comment.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')));

    // revert
    await page.locator(`textarea[name="${name}"]`).fill(original);
    await page.getByRole('button', { name: 'Save Files' }).click();
    await page.waitForLoadState();
    await expect(page.locator(`textarea[name="${name}"]`)).toHaveValue(original);

    console.log(`"${name}" textarea edited, saved, and reverted`);
  });
}

for (const name of enterpriseTextareas) {
  test(`edit, save, and revert "${name}" textarea (Enterprise only)`, async ({ page }) => {
    await page.goto('/settings/custom-code');

    const textarea = page.locator(`#${name}`);
    const visible = await textarea.isVisible().catch(() => false);
    test.skip(!visible, `"${name}" section requires an Enterprise license`);

    const original = await textarea.inputValue();
    const comment = `\n/* test-comment-${Date.now()} */`;

    await textarea.fill(original + comment);
    await page.getByRole('button', { name: 'Save Files' }).click();
    await page.waitForLoadState();
    await expect(page.locator(`#${name}`)).toHaveValue(new RegExp(comment.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')));

    // revert
    await page.locator(`#${name}`).fill(original);
    await page.getByRole('button', { name: 'Save Files' }).click();
    await page.waitForLoadState();
    await expect(page.locator(`#${name}`)).toHaveValue(original);

    console.log(`"${name}" textarea edited, saved, and reverted`);
  });
}

test('restore default fetches content into a textarea', async ({ page }) => {
  await page.goto('/settings/custom-code');

  const textarea = page.locator('#forbidden_usernames');
  const original = await textarea.inputValue();
  await textarea.fill('placeholder-to-be-replaced');

  await textarea.locator('xpath=ancestor::div[@x-data][1]').getByRole('button', { name: 'Restore Default' }).click();
  await expect(textarea).not.toHaveValue('placeholder-to-be-replaced', { timeout: 15_000 });

  // revert to original without saving the fetched defaults
  await textarea.fill(original);
  console.log('restore default fetched content into forbidden_usernames textarea');
});
