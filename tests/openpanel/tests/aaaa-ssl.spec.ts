import { test, expect } from '@playwright/test';
import tls from 'tls';

test.use({ ignoreHTTPSErrors: true });

const DOMAIN = 'wp.tests.openpanel.org';


const CERT_PEM = `-----BEGIN CERTIFICATE-----
MIIEnjCCA4agAwIBAgIUOfPOctSsB62Ls9PNnuvn7QCwu9AwDQYJKoZIhvcNAQEL
BQAwgYsxCzAJBgNVBAYTAlVTMRkwFwYDVQQKExBDbG91ZEZsYXJlLCBJbmMuMTQw
MgYDVQQLEytDbG91ZEZsYXJlIE9yaWdpbiBTU0wgQ2VydGlmaWNhdGUgQXV0aG9y
aXR5MRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRMwEQYDVQQIEwpDYWxpZm9ybmlh
MB4XDTI2MDQyODExNTAwMFoXDTQxMDQyNDExNTAwMFowYjEZMBcGA1UEChMQQ2xv
dWRGbGFyZSwgSW5jLjEdMBsGA1UECxMUQ2xvdWRGbGFyZSBPcmlnaW4gQ0ExJjAk
BgNVBAMTHUNsb3VkRmxhcmUgT3JpZ2luIENlcnRpZmljYXRlMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEApWRs6cxShnfWhncY6vqp9IElLYLTlW0SN1GY
nsawMIakavvlNB8ikPMuVc1NJwE2UnlFXH9gvHAIv8Wjc1WMGFnk5W3aB5rNP6dQ
NleNEt6egvlUwEnhti0YGXBq7XU/+MoDBXcMRgAWoVShQgt3ydF4EPYtRieHRh1R
9rEwsP284EUmfAt9kdnqcl71SqbRvzeiw5jahH8WLbpxrmMWWO3WHGJeaMXutcJ7
6G/q5Hz3nWDdCjpB6ZlsQ49HtHm/8AQwYIH5nj13IjvqYszZ5aNhxM+9topg55fc
B3a6LO9xpD6T+K83p/6xsU7yLqgo8kMQw6tsXKvOHR1ZYkBKEwIDAQABo4IBIDCC
ARwwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcD
ATAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBRX4fLT0Abd3ULSfFjnqM6BugEgxzAf
BgNVHSMEGDAWgBQk6FNXXXw0QIep65TbuuEWePwppDBABggrBgEFBQcBAQQ0MDIw
MAYIKwYBBQUHMAGGJGh0dHA6Ly9vY3NwLmNsb3VkZmxhcmUuY29tL29yaWdpbl9j
YTAhBgNVHREEGjAYghZ3cC50ZXN0cy5vcGVucGFuZWwub3JnMDgGA1UdHwQxMC8w
LaAroCmGJ2h0dHA6Ly9jcmwuY2xvdWRmbGFyZS5jb20vb3JpZ2luX2NhLmNybDAN
BgkqhkiG9w0BAQsFAAOCAQEAW5oOJHAxVaZTATxLN/9jSR8aZ/b8Z78TCZ8RQJwB
YAjug4SlrHnyka77VTFdD43GSQ3jI+iWemEJ8IyFyJSzpNe5HcTlXDahgjSSJDRo
lv8SYi/Y3fcTHemeX8hE4oXRDW60VUUjvCnxH1ncjJgcv5a5mnVs8POA3lGYaGQR
lr1Cze8vxaQSelVB247Eygc7WGjnJDcaYzswj5StR4nefXSWdMefiMkW57h0rps/
uypcrUy76JPIcepMQUBxWxkZvpCrzEtZfEGpMiMONZXc4CM+KvEo0Ap7hnvqVefi
CRtTQsSYwlpK+ouQGl2+hhseFP3paNe9JUsuq5hhUV+gmg==
-----END CERTIFICATE-----
`;

const KEY_PEM = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQClZGzpzFKGd9aG
dxjq+qn0gSUtgtOVbRI3UZiexrAwhqRq++U0HyKQ8y5VzU0nATZSeUVcf2C8cAi/
xaNzVYwYWeTlbdoHms0/p1A2V40S3p6C+VTASeG2LRgZcGrtdT/4ygMFdwxGABah
VKFCC3fJ0XgQ9i1GJ4dGHVH2sTCw/bzgRSZ8C32R2epyXvVKptG/N6LDmNqEfxYt
unGuYxZY7dYcYl5oxe61wnvob+rkfPedYN0KOkHpmWxDj0e0eb/wBDBggfmePXci
O+pizNnlo2HEz722imDnl9wHdros73GkPpP4rzen/rGxTvIuqCjyQxDDq2xcq84d
HVliQEoTAgMBAAECggEADi5/xkZeVyhkbTA+IzvuIciHGxFqOhhZpQOqnga0adzJ
sWC7BQ6cZKhtcy8A7BTHByhd4bIMZewHXAZC3ytZMWdX4LJcLSXBbrFWh+pW7uTG
270sXraXE4tnUxsYGBdjLl6IBspv83qjdh7vGt4n3dbHwFCTjj3qdAEkm44S+kIM
R3dvOLYU0F258sUJKbUmhmGDZj3oJV9gMQOE++swO2IabTYl1+mdFG9W6iaif/la
ETa1DABPEcRNZaKxoB/mLWomxSYXsVdl0o++PkC2Bf8GXW+TO+h6JyvMn194siIa
/V7gSpchogrFU8Ec3sGdStVlXIvUYCPfwgxFlOt8LQKBgQDUP5vMR1txMaU+oJMa
9/l4JDSG2NcbPiB6SYohtyoTqls50+wHcI0p1encfM6WX+XnTJ0771sLugovWZJW
oapJyHEqYaO4w2S/eyGY1uHTinlGt49m3N8984kFAG/MUQ20/wVDFN8oH7/13CBk
N0xBUtQ2wvi+SBaJ4nplSajSTQKBgQDHfDUQICyBV6Oy4X2usTyFVmdRl72PqBAL
mU7y2k9NmE2xhTro7VvTEBINcBHSrepzlAmRcnWQnD0GelO2nG7TAWgOV2OaGfW0
wwaCvzFrI3hx58jWy0k8Eg78a7k7GABj53H1AXNA7e3+BiJXCDcdynOmGD3VU/Zf
f2bEwUf93wKBgFfBqRAwXM2Tgkg/qjMXXm1fQtySYXYhHNqS92rzSZFx+WASkF+P
GL64dIY2kFA6fFtDISu7zoAtvrJPLaNmGnuBRdEJJ+Fn4IsPRRflmN+XPIeRs9gK
8L6zp+6KfK8UwD8axjkzMwVrAzqLdlUZTA0iSx4NRT2fnroKCyM/7m5tAoGAPQLU
B9aPRg/T1UX59o/mfrFqcB4EsAcqwSFmcAgs8QJ/4Kdq2QqfZvInU0zPZqwiZK8G
LiHfqxbd4zlOmS9HBeoMNTatE9iUuXBccWigaLA0ikHlvyv1fhXX14Pq5xP0KpoC
1HhZE6axf1vI7O1qTgY5ULdhUfmYBKUmfU7QAekCgYEAqSYnzTB4vcuDx9k6Ym2O
F/rqYnI665SUdAidYWKg9HGrFgzkuftii+Ep/6QP83nBr2Tm76kb4K0S1cg3xLrr
DAzx8uoCxjp5DLjaDIhruiacvjHfn21hmC6oB/pTwnwmt5F5P1Cq66j8N23vRgr2
Q1KxOtc7x30jSEzV4+veNew=
-----END PRIVATE KEY-----
`;

function getCert(domain: string) {
  return new Promise<any>((resolve, reject) => {
    const socket = tls.connect(
      {
        host: domain,
        port: 443,
        servername: domain, // SNI REQUIRED
        rejectUnauthorized: false, // Error: unable to verify the first certificate
      },
      () => {
        const cert = socket.getPeerCertificate(true);
        socket.end();
        resolve(cert);
      }
    );

    socket.setTimeout(5000, () => {
      socket.destroy();
      reject(new Error('TLS timeout'));
    });

    socket.on('error', reject);
  });
}



test('view ssl info', async ({ page }) => {
  await page.goto(`/domains/ssl?domain_name=${DOMAIN}`);
  // Confirm the SSL info panel is actually rendered
  await expect(page.locator('#view-cert')).toBeVisible();

  await page.waitForLoadState('networkidle');
  await expect(page.getByText(/US, CloudFlare, Inc|US, Let's Encrypt/i).first()).toBeVisible();
  
  console.log('ssl info for a domain is visible');
});



test('add custom ssl', async ({ page }) => {

  // Write files
  await page.goto('/file-manager/edit-file/cert.txt?editor=text&new=true');
  await page.locator('#editor-text').fill(CERT_PEM);
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText(/saved|success/i).first()).toBeVisible();

  await page.goto('/file-manager/edit-file/private.txt?editor=text&new=true');
  await page.locator('#editor-text').fill(KEY_PEM);
  await page.getByRole('button', { name: 'Save' }).click();
  await expect(page.getByText(/saved|success/i).first()).toBeVisible();

  // 2. set ssl in UI
  await page.goto(`/domains/ssl?domain_name=${DOMAIN}`);
  await page.locator('input[name="public_path"]').fill('/var/www/html/cert.txt');
  await page.locator('input[name="private_path"]').fill('/var/www/html/private.txt');
  await page.getByRole('button', { name: 'Configure Custom Certificate' }).click();

  await expect(page.getByText(/to use custom ssl/i).first()).toBeVisible();

  await page.waitForLoadState('networkidle');
  await expect(page.getByText(/US, CloudFlare, Inc/i).first()).toBeVisible();

  // 3. verify panel shows it
  await page.locator('a#view-cert').click();
  const certCode = page.locator('pre#certCode');
  await expect(certCode).toBeVisible();

  await expect(certCode).toContainText('BEGIN CERTIFICATE');
  await expect(certCode).toContainText('BEGIN PRIVATE KEY');

  // 4. verify domain uses it
  const domainPage = await page.context().newPage();
  await domainPage.goto(`https://${DOMAIN}`);
  const title = await domainPage.title();
  expect(title).not.toMatch(/privacy error|your connection is not private|err_cert/i);
  await domainPage.close();

  const cert = await getCert(DOMAIN);
  const issuer = cert?.issuer?.O || cert?.issuer?.CN || '';
  console.log('Issuer:', issuer);
  expect(issuer).toMatch(/cloudflare/i);
  console.log('custom ssl is working');
});



test('switch back to Lets Encrypt', async ({ page }) => {

  // 1. set autossl in UI
  await page.goto(`/domains/ssl?domain_name=${DOMAIN}`);
  await page.getByRole('button', { name: "Switch to Let's Encrypt and generate" }).click();

  await expect(page.getByText(/to use AutoSSL/i).first()).toBeVisible();

  await page.waitForLoadState('networkidle');
  await expect(page.getByText(/US, Let's Encrypt/i).first()).toBeVisible();

  // 2. verify panel shows it
  await page.locator('a#view-cert').click();
  const certCode = page.locator('pre#certCode');
  await expect(certCode).toBeVisible();

  await expect(certCode).toContainText('BEGIN CERTIFICATE');
  await expect(certCode).toContainText('BEGIN EC PRIVATE KEY');

  // 3. verify domain uses it
  const domainPage = await page.context().newPage();
  await domainPage.goto(`https://${DOMAIN}`);

  const title = await domainPage.title();
  expect(title).not.toMatch(/privacy error|your connection is not private|err_cert/i);
  await domainPage.close();

  const cert = await getCert(DOMAIN);
  const issuer = cert?.issuer?.O || cert?.issuer?.CN || '';
  console.log('Issuer:', issuer);
  expect(issuer).toMatch(/let'?s encrypt|isrg|r3/i);

  console.log('switch from custom to LE is working!');
});
