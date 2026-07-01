import { test, expect } from '@playwright/test';

function randomSuffix() {
  return Math.random().toString(36).slice(2, 8);
}

const suffix = randomSuffix();
const FILE_NAME = `radovanfajl_${suffix}`;
const FOLDER_NAME = `radovanfolder_${suffix}`;
const TXT_FILE = `petarfajl_${suffix}.txt`;
const TXT_FILE_BAK = `petarfajl_${suffix}.txt_bak`;
const ZIP_FILE = `radozip_${suffix}`;
const ZIP_FOLDER = `radofol_${suffix}`;
const ZIP_ARCHIVE = `/rasizip_${suffix}`;
const ZIP_ARCHIVE_NAME = `rasizip_${suffix}.zip`;

// Subdirectory used for tests that need cleanup, to avoid deleting docroots
const TEST_SUBDIR = `test_subdir_${suffix}`;
const TEST_FILE = `test_file_${suffix}`;
const TEST_DIR = `test_dir_${suffix}`;


async function navigateToFiles(page: any) {
  await page.goto(`/files`);
}

async function navigateToSubdir(page: any) {
  	await page.goto(`/files/${TEST_SUBDIR}`);
}

async function enableOwnerColumn(page: any) {
    await page.locator('#dropdownToggleButton').click();
    const ownerToggle = page.locator('label', { hasText: 'Owner' });
    const isChecked = await ownerToggle.locator('input').isChecked();
    if (!isChecked) await ownerToggle.click();
    await page.locator('#dropdownToggleButton').click();
}

// https://github.com/stefanpejcic/OpenPanel/issues/976
async function verifyOwnerUids(page: any) {
    await enableOwnerColumn(page);
    const ownerCells = page.locator('#filemanager_table .owner-cell');
    const count = await ownerCells.count();
    expect(count).toBeGreaterThan(0);
    for (let i = 0; i < count; i++) {
        const text = (await ownerCells.nth(i).textContent())?.trim();
        if (!text) continue;
        const uid = parseInt(text, 10);
        expect(uid).toBeGreaterThanOrEqual(1000);
    }
}


async function createFileInRoot(page: any, fileName: string, openAfterCreate = false) {
  await navigateToFiles(page);
  await page.getByRole('button', { name: ' New File' }).click();
  await page.getByRole('textbox', { name: 'File Name*' }).fill(fileName);
  if (openAfterCreate) {
    await page.locator('#open').check();
  }
  await page.getByRole('button', { name: 'Create' }).click();
}

async function createFile(page: any, fileName: string, openAfterCreate = false) {
  await page.getByRole('button', { name: ' New File' }).click();
  await page.getByRole('textbox', { name: 'File Name*' }).fill(fileName);
  if (openAfterCreate) {
    await page.locator('#open').check();
  }
  await page.getByRole('button', { name: 'Create' }).click();
}


async function createFolderInRoot(page: any, folderName: string) {
  await navigateToFiles(page);
  await page.getByRole('button', { name: ' New Folder' }).click();
  await page.locator('#foldername').fill(folderName);
  await page.getByRole('button', { name: 'Create' }).click();
}

async function createFolder(page: any, folderName: string) {
  await page.getByRole('button', { name: ' New Folder' }).click();
  await page.locator('#foldername').fill(folderName);
  await page.getByRole('button', { name: 'Create' }).click();
}

async function selectItem(page: any, name: string, multiSelect = false) {
  await page.locator('#filemanager_table div').filter({ hasText: name }).click(
    multiSelect ? { modifiers: ['ControlOrMeta'] } : undefined
  );
}

async function deleteSelected(page: any, skipTrash = false) {
  await page.locator('#deleteButton').click();
  if (skipTrash) {
    await page.getByRole('checkbox', { name: 'Skip the trash and' }).check();
  }
  await page.getByRole('button', { name: 'Delete', exact: true }).click();
}

// Cleanup scoped to TEST_SUBDIR only, to avoid deleting docroots
async function cleanupSubdir(page: any) {
	await createFolderInRoot(page, TEST_SUBDIR);
	await page.waitForTimeout(5000);
	await navigateToSubdir(page);
	await createFile(page, TEST_FILE);
	await page.waitForTimeout(5000);
	await createFolder(page, TEST_DIR);
	await page.getByRole('button', { name: ' Select all' }).click();
	await page.locator('#deleteButton').click();
	await page.getByText('Skip the trash and').click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(page.locator('body')).toContainText(/No items found/i);
}
// TODO: test toggle column names makes them visible in the table
	
// TODO: test for folder search and file search
	
// TODO: copy path button test
	
// TODO: looooong breadcrumbs test
	
// TODO: test upload drag and drop
	
// TODO: test upload multiple files
	


test('create file', async ({ page }) => {
  await navigateToFiles(page);

  await createFileInRoot(page, FILE_NAME);
  await expect(page.locator('body')).toContainText(/File created successfully/i);
  await expect(page.locator('body')).toContainText(new RegExp(FILE_NAME, 'i'));
	await verifyOwnerUids(page);
  console.log('File created successfully');
});

test('create folder', async ({ page }) => {
  await navigateToFiles(page);

  await createFolderInRoot(page, FOLDER_NAME);
  await expect(page.locator('body')).toContainText(/Folder created successfully/i);
  await expect(page.locator('body')).toContainText(new RegExp(FOLDER_NAME, 'i'));
  await verifyOwnerUids(page);
  console.log('Folder created successfully');
});

test('copy file to folder', async ({ page }) => {
  await navigateToFiles(page);

  await selectItem(page, FILE_NAME);
  await page.getByRole('button', { name: ' Copy' }).click();
  await page.locator('#copyfiletree').getByText(FOLDER_NAME).click();
  await page.getByRole('button', { name: 'Copy', exact: true }).click();

  await expect(page.locator('body')).toContainText(/Copy complete/i);
  await expect(page.locator('body')).toContainText(new RegExp(FILE_NAME, 'i'));

  console.log('File copied into folder successfully');
});


test('move file', async ({ page }) => {
  await page.goto(`/files`);

  await page.getByRole('link', { name: FOLDER_NAME }).click();
  await expect(page).toHaveURL(/files\/radovanfolder/);
  await expect(page.locator('body')).toContainText(new RegExp(FILE_NAME, 'i'));
  await selectItem(page, FILE_NAME);
  await page.getByRole('button', { name: ' Move' }).click();
  await page.getByRole('textbox', { name: 'Where to:*' }).click();
  await page.getByRole('textbox', { name: 'Where to:*' }).fill('/');
  await page.getByRole('button', { name: 'Move', exact: true }).click();

  await expect(page.locator('body')).toContainText(/Move complete/i);
  await expect(page.locator('body')).not.toContainText(new RegExp(FILE_NAME, 'i'));
  await expect(page.locator('body')).toContainText(/No items found/i);

  await page.getByRole('link', { name: '/var/www/html/' }).click();
  await expect(page.locator('body')).toContainText(new RegExp(FILE_NAME, 'i'));

  console.log('File moved out of folder successfully');
});


test('delete file to trash', async ({ page }) => {
  await navigateToFiles(page);

  await selectItem(page, FILE_NAME);
  await page.locator('#deleteButton').click();
  await page.getByRole('button', { name: 'Delete', exact: true }).click();
  await expect(page.locator('body')).not.toContainText(new RegExp(FILE_NAME, 'i'));

  console.log('File moved to trash successfully');
});

test('restore file from trash', async ({ page }) => {
  await navigateToFiles(page);
  await page.getByRole('link', { name: 'Trash' }).click();
  await expect(page.locator('body')).toContainText(new RegExp(FILE_NAME, 'i'));
  await selectItem(page, FILE_NAME);
  await page.click('#restoreButton');
  await page.getByRole('button', { name: 'Restore', exact: true }).click();

  // Refresh before verifying file is no longer in trash
  await page.waitForTimeout(5000);
  await page.reload();
  await expect(page.locator('body')).not.toContainText(new RegExp(FILE_NAME, 'i'));

  await page.getByRole('link', { name: 'File Manager' }).click();
  await expect(page.locator('body')).toContainText(new RegExp(FILE_NAME, 'i'));
  console.log('File restored from trash successfully');
});

test('delete multiple items permanently', async ({ page }) => {
  await navigateToFiles(page);

  await selectItem(page, FOLDER_NAME);
  await selectItem(page, FILE_NAME, true);
  await deleteSelected(page, true);

  await expect(page.locator('body')).not.toContainText(new RegExp(FILE_NAME, 'i'));
  await expect(page.locator('body')).not.toContainText(new RegExp(FOLDER_NAME, 'i'));

  console.log('Multiple items permanently deleted successfully');
});


async function createFileWithEditor(page: any, fileName: string) {
  await navigateToFiles(page);
  await createFileInRoot(page, fileName, true);
  await page.locator('.view-lines').click();
  await page.getByRole('textbox', { name: 'Editor content;Press Alt+F1' }).fill('nekitext');
  await page.getByRole('button', { name: 'Save' }).click();
  await page.locator('#fullscreenButton').click();
  await expect(page.locator('body')).toContainText(/File saved successfully/i);
  console.log('File created with editor successfully');
}

async function viewFile(page: any, fileName: string, expectedContent: string) {
  await navigateToFiles(page);
  await expect(page.locator('body')).toContainText(new RegExp(fileName, 'i'));

  await selectItem(page, fileName);
  const popupPromise = page.waitForEvent('popup');
  await page.getByRole('button', { name: ' View' }).click();
  const popup = await popupPromise;
  await expect(popup.locator('body')).toContainText(new RegExp(expectedContent, 'i'));
  console.log('File viewed successfully');
}

async function editFile(page: any, fileName: string, newContent: string) {
  await navigateToFiles(page);
  await selectItem(page, fileName);
  const popupPromise = page.waitForEvent('popup');
  await page.getByRole('button', { name: ' Edit' }).click();
  const popup = await popupPromise;
  await popup.locator('div').filter({ hasText: /^nekitext$/ }).nth(2).click();
  await popup.getByRole('textbox', { name: 'Editor content;Press Alt+F1' }).fill(newContent);
  await popup.getByRole('button', { name: 'Save' }).click();
  console.log('File edited successfully');
}

test('create file with editor', async ({ page }) => {
  await createFileWithEditor(page, TXT_FILE);
});

test('view file content', async ({ page }) => {
  await viewFile(page, TXT_FILE, 'nekitext');
});

test('edit file content', async ({ page }) => {
  await editFile(page, TXT_FILE, 'nekitext2');

  // Verify updated content is visible after edit
  const verifyPopupPromise = page.waitForEvent('popup');
  await page.getByRole('button', { name: ' View' }).click();
  const verifyPopup = await verifyPopupPromise;
  await expect(verifyPopup.locator('body')).toContainText(/nekitext2/i);

  console.log('File content updated and verified successfully');
});


test('rename file', async ({ page }) => {
  await navigateToFiles(page);

  await selectItem(page, TXT_FILE);
  await page.getByRole('button', { name: ' Rename' }).click();
  await page.locator('#renameInput').click();
  await page.locator('#renameInput').fill(TXT_FILE_BAK);
  await page.getByRole('button', { name: 'Rename', exact: true }).click();

  // Verify old filename is gone and new filename is present
  await expect(page.locator('body')).not.toContainText(new RegExp(`^${TXT_FILE}$`, 'i'));
  await expect(page.locator('body')).toContainText(new RegExp(TXT_FILE_BAK, 'i'));
  await expect(page.locator('body')).toContainText(/File renamed successfully/i);

  console.log('File renamed successfully');
});


test('change file permissions', async ({ page }) => {
  await navigateToFiles(page);

  await selectItem(page, TXT_FILE_BAK);
  await page.getByRole('button', { name: ' Permissions' }).click();
  await page.getByPlaceholder('775').fill('755');
  await page.getByRole('button', { name: 'Confirm' }).click();

  await expect(page.locator('body')).toContainText(/Permissions changed/i);
  await expect(page.locator('body')).toContainText(/-rwxr-xr-x/i);
  await verifyOwnerUids(page);

  await selectItem(page, TXT_FILE_BAK);
  await page.getByRole('button', { name: ' Permissions' }).click();
  await expect(page.getByPlaceholder('775')).toHaveValue('755');

  console.log('File permissions changed successfully');
});

test('upload file from URL', async ({ page }) => {
  test.setTimeout(180_000);

  await navigateToFiles(page);

  const page5Promise = page.waitForEvent('popup');
  await page.getByRole('button', { name: ' Upload' }).click();
  const page5 = await page5Promise;

  await page5.getByRole('button', { name: 'Download from URL instead' }).click();
  await page5.getByRole('textbox', { name: 'https://' }).fill('http://ipv4.download.thinkbroadband.com/20MB.zip');
  await page5.getByRole('button', { name: 'Download' }).click();
  await expect(page5.locator('body')).toContainText(/downloaded from URL successfully/i, { timeout: 20_000 });

  await page5.getByRole('link', { name: 'File Manager' }).click();
  await page5.getByRole('heading', { name: '20MB.zip', exact: true }).click();

  await navigateToFiles(page);
  await expect(page.locator('body')).toContainText(/20MB.zip/i);
  await verifyOwnerUids(page);
  console.log('File uploaded from URL successfully');
});


async function compressFiles(page: any) {
  await navigateToFiles(page);

  await createFileInRoot(page, ZIP_FILE);
  await createFolderInRoot(page, ZIP_FOLDER);

  await selectItem(page, ZIP_FOLDER);
  await selectItem(page, ZIP_FILE, true);
  await page.getByRole('button', { name: ' Compress' }).click();
  await page.getByRole('textbox', { name: 'Archive path*' }).fill(ZIP_ARCHIVE);
  await page.getByRole('button', { name: 'Compress', exact: true }).click();
  await expect(page.locator('body')).toContainText(new RegExp(ZIP_ARCHIVE_NAME, 'i'));
  console.log('Files compressed successfully');
}

async function extractFiles(page: any) {
  await navigateToFiles(page);

  await selectItem(page, ZIP_FILE);
  await selectItem(page, ZIP_FOLDER, true);
  await deleteSelected(page);
  await expect(page.locator('body')).not.toContainText(new RegExp(ZIP_FILE, 'i'));
  await expect(page.locator('body')).not.toContainText(new RegExp(ZIP_FOLDER, 'i'));

  await selectItem(page, ZIP_ARCHIVE_NAME);
  await page.getByRole('button', { name: ' Extract' }).click();
  await page.getByRole('button', { name: 'Extract', exact: true }).click();
  await expect(page.locator('body')).toContainText(/File extracted successfully/i);
  await expect(page.locator('body')).toContainText(new RegExp(ZIP_FILE, 'i'));
  await expect(page.locator('body')).toContainText(new RegExp(ZIP_FOLDER, 'i'));
  console.log('Files extracted successfully');
}

test('compress files', async ({ page }) => {
  await compressFiles(page);
  await verifyOwnerUids(page);
});

test('extract files', async ({ page }) => {
  await extractFiles(page);
  await verifyOwnerUids(page);
  // TODO: cover all 3 supported archive extensions
  // Cleanup scoped to test subdir to avoid deleting docroots
});

test('cleanup subdir', async ({ page }) => {
  await cleanupSubdir(page);
  // TODO: cover all 3 supported archive extensions
  // Cleanup scoped to test subdir to avoid deleting docroots
});


// INODES AND DISK USAGE EXPLORERS
const explorerTests = [
  {
    route: '/disk-usage',
    columnHeader: 'Size',
    valueRegex: /^\d+(\.\d+)?\s?[KMGT]?B?$/ 
  },
  {
    route: '/inodes-explorer',
    columnHeader: 'INodes',
    valueRegex: /^\d+$/ 
  }
];

for (const { route, columnHeader, valueRegex } of explorerTests) {
  test(route, async ({ page }) => {
    // 1. Initial
    await page.goto(route);
    const table = page.locator('#folders_to_navigate');
    
    await expect(table).toBeVisible();
    await expect(table).toContainText('docker-data');
    await expect(table.locator('th', { hasText: columnHeader })).toBeVisible();
    
    const chart = page.locator('#folderChart');
    await expect(chart).toBeVisible();

    let box = await chart.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.width).toBeGreaterThan(0);

    const valueCells = table.locator('tbody tr td:nth-child(2)');
    const count = await valueCells.count();
    for (let i = 0; i < count; i++) {
      const text = (await valueCells.nth(i).textContent())?.trim();
      expect(text).toMatch(valueRegex);
    }

    // 2. enter "docker-data'
    const dockerLink = page.locator(`a[href*="${route}/docker-data"]`);
    await dockerLink.click();

    await expect(page).toHaveURL(new RegExp(`${route}/docker-data/?`));
    
    const table2 = page.locator('#folders_to_navigate');
    await expect(table2).toContainText('containerd');
    await expect(table2.locator('th', { hasText: columnHeader })).toBeVisible();
    
    const containerdRow = table2.locator('tr', { hasText: 'containerd' });
    const valueCell = containerdRow.locator('td').nth(1);
    
    await expect(valueCell).toBeVisible();
    const cellText = (await valueCell.textContent())?.trim();
    expect(cellText).toMatch(valueRegex);

    const isRendered = await page.evaluate(() => {
      const canvas = document.getElementById('folderChart') as HTMLCanvasElement;
      if (!canvas) return false;
      const ctx = canvas.getContext('2d');
      if (!ctx) return false;
      const data = ctx.getImageData(0, 0, canvas.width, canvas.height).data;
      return Array.from(data).some((v) => v !== 0);
    });
    expect(isRendered).toBeTruthy();

    // 3. "Up One Level"
    const upOneLevelLink = page.locator('a', { hasText: 'Up One Level' });
    await upOneLevelLink.click();
    
    await expect(page).toHaveURL(new RegExp(`${route}/?$`));
    const table3 = page.locator('#folders_to_navigate');
    await expect(table3).toContainText('docker-data');
    await expect(table3).not.toContainText('containerd');
    
    console.log(`${route} is functional`);
  });
}
