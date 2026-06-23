import { test, expect } from '@playwright/test';

const localeMapping = {
    'ne': 'ne-np',
    'en': 'en-us',
    'pt': 'pt-br',
    'uk': 'uk-ua',
    'zh': 'zh-cn',
    'sr': 'sr-rs',
};

async function getTranslation(locale) {
    const folder = localeMapping[locale] || `${locale}-${locale}`;
    const url = `https://raw.githubusercontent.com/stefanpejcic/openpanel-translations/main/${folder}/messages.po`;
    
    try {
        const response = await fetch(url);
        if (!response.ok) return null;
        const text = await response.text();
        
        // Matches the msgstr for the "Change Language" msgid
        const regex = /msgid "Change Language"\s+msgstr "(.*)"/;
        const match = text.match(regex);
        return match ? match[1] : null;
    } catch (e) {
        return null;
    }
}

const localesToTest = [ 'sr', 'bg', 'de', 'es', 'fr', 'hu', 'ne', 'pt', 'ro', 'ru', 'tr', 'uk', 'zh', 'en'];

test.describe('Change and use locale', () => {
    
    for (const locale of localesToTest) {
        
        test(`locale: ${locale}`, async ({ page }) => {                   
            await page.goto('/account/language');
            await page.selectOption('#locale-select', locale);
            let expectedText;
            if (locale === 'en') {
                expectedText = 'Change Language';
            } else {
                expectedText = await getTranslation(locale);
            }
            test.skip(!expectedText, `Translation key for "${locale}" not found on GitHub.`);
            const locator = page.getByText(expectedText, { exact: true }).first();
            await expect(locator).toBeVisible();
        });
    }
});
