import { defineConfig } from '@playwright/test';
import dotenv from 'dotenv';
import path from 'path';

dotenv.config({ path: path.resolve(__dirname, '.env') });

const BASE_URL = process.env.BASE_URL;

if (!BASE_URL) {
  throw new Error('BASE_URL must be set in .env');
}

export default defineConfig({
  testDir: '.',

  use: {
    baseURL: BASE_URL,
    ignoreHTTPSErrors: true,
  },

  projects: [
    {
      name: 'setup',
      testMatch: /auth\.setup\.ts/,
    },
    {
      name: 'tests',
      testDir: './tests',
      testMatch: /.*\.spec\.ts/,
      dependencies: ['setup'],
      use: {
        storageState: path.join(__dirname, '../.auth/session.json'),
      },
    },
  ],
});
