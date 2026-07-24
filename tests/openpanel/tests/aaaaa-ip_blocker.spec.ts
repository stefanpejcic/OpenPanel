import { test, expect, Page } from '@playwright/test';

test('IP Blocker', async ({ page, context }) => {

    // 1. get ip
    await page.goto('https://ip.openpanel.com/');
    const serverIP = (await page.textContent('body'))?.trim();
    console.log(`Testing with IP: ${serverIP}`);

    // 2. block ip
    await page.goto('/security/ip-blocker');
    await expect(page.getByText('CIDR')).toBeVisible();
    await page.locator('#blocked_ips').fill(serverIP);
    await page.click('#save-ips');
    await expect(page.getByText('IP addresses have been successfully added to blocklist and can no longer access websites')).toBeVisible();

    // 3. test
    await page.waitForTimeout(2000);
    await test.step('Verify website is blocked', async () => {
        await expect.poll(async () => {
            try {
                const response = await page.goto('https://wp.tests.openpanel.org', { timeout: 2000 });
                return response?.status();
            } catch (e) {
                return 'blocked'; // Playwright throws if navigation is aborted/blocked
            }
        }, {
            message: 'Website should be blocked',
            intervals: [1000],
            timeout: 10000,
        }).not.toBe(200);
    });

    // 4. remove ip from the blocklist
    await page.goto('/security/ip-blocker');
    await page.locator('#blocked_ips').clear();
    await page.click('#save-ips');
    await expect(page.getByText('All IP addresses have been successfully removed from blocklist and can now access websites')).toBeVisible();

    // 5. test again
    await page.waitForTimeout(2000);
    await test.step('Verify website is accessible', async () => {
        await expect.poll(async () => {
            const response = await page.goto('https://wp.tests.openpanel.org');
            return response?.status();
        }, {
            message: 'Website should be accessible',
            intervals: [1000],
            timeout: 10000,
        }).toBe(200);
    });

    console.log(`IP blocker is working for IP address ${serverIP}`);
});
