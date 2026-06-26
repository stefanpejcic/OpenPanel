import { test, expect } from '@playwright/test';

// Covers modules/json/favicons.py (/json/favicon/<domain>) and
// modules/json/screenshots.py (/json/screenshot/<domain>) — both used as embedded
// image sources on the domains list and website manager pages.

const domain = 'wp.tests.openpanel.org';

test('favicon endpoint redirects to a favicon image', async ({ page }) => {
  const response = await page.goto(`/json/favicon/${domain}`);
  expect(response).not.toBeNull();
  expect(response!.status()).toBeLessThan(400);
  const contentType = response!.headers()['content-type'] ?? '';
  expect(contentType).toMatch(/image/);
  console.log('favicon endpoint is working');
});

test('favicon endpoint rejects a domain the user does not own', async ({ page }) => {
  const response = await page.goto('/json/favicon/not-owned-by-this-user.example.com');
  expect(response).not.toBeNull();
  expect(response!.status()).toBe(403);
  console.log('favicon endpoint enforces domain ownership');
});

test('screenshot endpoint returns an image', async ({ page }) => {
  test.setTimeout(60_000);
  const response = await page.goto(`/json/screenshot/${domain}`);
  expect(response).not.toBeNull();
  expect(response!.status()).toBeLessThan(400);
  const contentType = response!.headers()['content-type'] ?? '';
  expect(contentType).toMatch(/image/);
  console.log('screenshot endpoint is working');
});

test('screenshot endpoint rejects a domain the user does not own', async ({ page }) => {
  const response = await page.goto('/json/screenshot/not-owned-by-this-user.example.com');
  expect(response).not.toBeNull();
  expect(response!.status()).toBe(403);
  console.log('screenshot endpoint enforces domain ownership');
});

test('screenshot can be force-regenerated via POST', async ({ page }) => {
  test.setTimeout(60_000);
  await page.goto(`/website?domain=${domain}`);

  const status = await page.evaluate(async (d) => {
    const res = await fetch(`/json/screenshot/${d}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-CSRF-TOKEN': (window as any).csrf_token },
    });
    return res.status;
  }, domain);

  expect(status).toBeLessThan(400);
  console.log('screenshot regeneration endpoint is working');
});
