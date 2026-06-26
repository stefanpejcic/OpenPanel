import { test, expect } from '@playwright/test';


const testConfigs = [
  // /domains/zone-templates
  { url: '/domains/zone-templates', id: 'zone_template_ipv4', name: 'IPv4 Zone', restoreIndex: 0 },
  { url: '/domains/zone-templates', id: 'zone_template_ipv6', name: 'IPv6 Zone', restoreIndex: 1 },
  
  // /domains/file-templates
  { url: '/domains/file-templates', id: 'default_page', name: 'Default Page', restoreIndex: 0 },
  { url: '/domains/file-templates', id: 'suspended_website', name: 'Suspended Website', restoreIndex: 1 },
  { url: '/domains/file-templates', id: 'suspended_user', name: 'Suspended User', restoreIndex: 2 },
  { url: '/domains/file-templates', id: 'docker_apache_domain', name: 'Apache VHost', restoreIndex: 3 },
  { url: '/domains/file-templates', id: 'docker_nginx_domain', name: 'Nginx VHost', restoreIndex: 4 },
  { url: '/domains/file-templates', id: 'docker_openresty_domain', name: 'OpenResty VHost', restoreIndex: 5 },
  { url: '/domains/file-templates', id: 'docker_varnish', name: 'Varnish Template', restoreIndex: 6 },
  { url: '/domains/file-templates', id: 'docker_caddy', name: 'Caddy VHosts', restoreIndex: 7 },
];

for (const config of testConfigs) {
  test(`edit template: ${config.name}`, async ({ page }) => {
    await page.goto(config.url);

    const textarea = page.locator(`textarea#${config.id}`);
    const comment = `\n;test comment for ${config.id}`;
    const saveButton = page.getByRole('button', { name: 'Save Files' });
    const successMsg = page.getByText(/Templates? updated successfully/);

    // EDIT AND SAVE
    const originalValue = await textarea.inputValue();
    await textarea.fill(originalValue + comment);
    await saveButton.click();

    await expect(successMsg).toBeVisible();
    await expect(textarea).toHaveValue(new RegExp(comment));

    // RESTORE AND SAVE
    await page.getByRole('button', { name: 'Restore Default' }).nth(config.restoreIndex).click();
    await expect(textarea).not.toHaveValue(new RegExp(comment));
    await saveButton.click();

    await expect(successMsg).toBeVisible();
    await expect(textarea).not.toHaveValue(new RegExp(comment));
  });
}



test('edit zone for a domain', async ({ page }) => {
  await page.goto('/domains/dns');
  const domainSelect = page.locator('#domains');
  const options = domainSelect.locator('option:not([disabled])');
  const firstOption = options.first();
  const domainValue = await firstOption.getAttribute('value');
  const domainName = (await firstOption.textContent())?.trim();
  await domainSelect.selectOption(domainValue!);
  await expect(page).toHaveURL(new RegExp(`/domains/dns/${domainValue}`));
  const textarea = page.locator('[name="bind_content"]');
  await expect(textarea).toBeVisible();
  const existingContent = await textarea.inputValue();
  const updatedContent = `${existingContent}\n;added a comment`;
  await textarea.fill(updatedContent);
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText(`Zone file for ${domainName} saved successfully and DNS service reloaded.`)).toBeVisible();
  await expect(textarea).toHaveValue(/;added a comment/);
  // TODO: check backend added \n at end of file
});
