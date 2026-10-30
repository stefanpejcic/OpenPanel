import { defineConfig } from '@playwright/test';

const port = Number(process.env.PORT || 4173);
// PW_CHANNEL=msedge or chrome reuses an installed browser instead of downloading Chromium
const channel = process.env.PW_CHANNEL || undefined;
const sizes = [
  { name: 'phone', viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true, deviceScaleFactor: 2 },
  { name: 'tablet', viewport: { width: 820, height: 1180 }, hasTouch: true },
  { name: 'desktop', viewport: { width: 1440, height: 900 } },
];

export default defineConfig({
  testDir: './specs',
  outputDir: './test-results',
  fullyParallel: true,
  workers: process.env.CI ? 2 : undefined,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: { baseURL: `http://127.0.0.1:${port}`, channel, reducedMotion: 'reduce' },
  projects: sizes.flatMap(({ name, ...size }) => ['dark', 'light'].map((colorScheme) => ({
    name: `${name}-${colorScheme}`,
    use: { ...size, colorScheme },
  }))),
  webServer: { command: 'node server.mjs', url: `http://127.0.0.1:${port}/themes/modern/index.html`, reuseExistingServer: !process.env.CI },
});
