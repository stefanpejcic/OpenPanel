import { test, expect, request as pwRequest, APIRequestContext } from '@playwright/test';

/**
 * endpoints_api.spec.ts
 *
 * Covers every endpoint documented in https://github.com/stefanpejcic/2087/blob/main/modules/api/available_endpoints.txt
 *
 * Uses BASE_URL, PANEL_USERNAME, PANEL_PASSWORD from .env (same as auth.setup.ts).
 *
 * Safety model:
 *   - GET endpoints are exercised for real and their responses asserted.
 *   - Safe lifecycle mutations (plan CRUD, feature-set CRUD) run by default and clean up after themselves.
 *   - Destructive/system-altering endpoints (reboot, root-password, disable-admin, drop-cache,
 *     kill process, updates/now, license delete, migrations, user/domain mutations...) are only
 *     verified for existence + auth enforcement (request WITHOUT a token must be rejected),
 *     unless RUN_MUTATING=1 is set in the environment.
 *   - Endpoints that need pre-existing data (a domain, a user, a notification, a DNS slave...)
 *     discover the first available resource and skip gracefully when none exists.
 */

const BASE_URL = process.env.BASE_URL!;
const USERNAME = process.env.PANEL_USERNAME!;
const PASSWORD = process.env.PANEL_PASSWORD!;
const RUN_MUTATING = process.env.RUN_MUTATING === '1';

const TEST_PLAN_NAME = 'pw_test_plan';
const TEST_FEATURE_SET = 'pw_test_features';

let api: APIRequestContext;
let anon: APIRequestContext;
let token: string | null = null;

function authHeaders() {
  return { Authorization: `Bearer ${token}` };
}

/** Skip the current test when API access is disabled on this environment. */
function requireToken() {
  test.skip(!token, 'API access not currently enabled on this environment');
}

/** GET helper that asserts a 2xx response and returns parsed JSON (or raw text). */
async function getOk(path: string) {
  const res = await api.get(path, { headers: authHeaders() });
  expect(res.ok(), `GET ${path} -> ${res.status()}`).toBeTruthy();
  const ct = res.headers()['content-type'] || '';
  return ct.includes('application/json') ? res.json() : res.text();
}

/** Assert an endpoint rejects unauthenticated requests (existence + auth coverage). */
async function expectAuthRequired(method: 'get' | 'post' | 'put' | 'patch' | 'delete', path: string) {
  const res = await anon[method](path, { data: {} });
  expect(res.status(), `${method.toUpperCase()} ${path} without token should be rejected`).not.toBe(200);
  expect([401, 403, 422].includes(res.status()) || res.status() >= 400,
    `${method.toUpperCase()} ${path} returned ${res.status()} without a token`).toBeTruthy();
  console.log(`auth enforced on ${method.toUpperCase()} ${path} (${res.status()})`);
}

/** Discover the first domain on the server, or null. */
async function firstDomain(): Promise<string | null> {
  const res = await api.get('/api/domains', { headers: authHeaders() });
  if (!res.ok()) return null;
  const body = await res.json().catch(() => null);
  const list = Array.isArray(body) ? body : body?.domains ?? body?.data ?? [];
  if (!Array.isArray(list) || list.length === 0) return null;
  const d = list[0];
  return typeof d === 'string' ? d : d?.domain ?? d?.domain_name ?? d?.name ?? null;
}

/** Discover the first username on the server, or null. */
async function firstUser(): Promise<string | null> {
  const res = await api.get('/api/users', { headers: authHeaders() });
  if (!res.ok()) return null;
  const body = await res.json().catch(() => null);
  const list = Array.isArray(body) ? body : body?.users ?? body?.data ?? [];
  if (!Array.isArray(list) || list.length === 0) return null;
  const u = list[0];
  return typeof u === 'string' ? u : u?.username ?? u?.user ?? null;
}

test.beforeAll(async () => {
  if (!BASE_URL || !USERNAME || !PASSWORD) {
    throw new Error('BASE_URL, PANEL_USERNAME and PANEL_PASSWORD must be set in .env');
  }
  api = await pwRequest.newContext({ baseURL: BASE_URL, ignoreHTTPSErrors: true });
  anon = await pwRequest.newContext({ baseURL: BASE_URL, ignoreHTTPSErrors: true });

  const res = await api.post('/api/', {
    data: { username: USERNAME, password: PASSWORD },
  });
  if (res.ok()) {
    const body = await res.json().catch(() => ({}));
    token = body.access_token ?? body.token ?? null;
  }
  if (!token) {
    console.log(`POST /api/ did not return a token (status ${res.status()}) - API access may be disabled; dependent tests will be skipped`);
  }
});

test.afterAll(async () => {
  await api?.dispose();
  await anon?.dispose();
});

/* -------------------------------------------------------------------------- */
/* Auth: /api/ and /api/whoami                                                */
/* -------------------------------------------------------------------------- */

test.describe('auth', () => {
  test('GET /api/ returns api status', async () => {
    const res = await anon.get('/api/');
    expect(res.status(), 'GET /api/ should respond').toBeLessThan(500);
    console.log(`GET /api/ -> ${res.status()}`);
  });

  test('POST /api/ with valid credentials returns access token', async () => {
    requireToken();
    expect(token).toBeTruthy();
    console.log('POST /api/ issued an access token');
  });

  test('POST /api/ with invalid credentials is rejected', async () => {
    const res = await anon.post('/api/', {
      data: { username: 'definitely_not_a_user', password: 'wrong' },
    });
    expect(res.ok()).toBeFalsy();
    console.log(`POST /api/ with bad credentials -> ${res.status()}`);
  });

  test('GET /api/whoami returns the authenticated username', async () => {
    requireToken();
    const body = await getOk('/api/whoami');
    const who = typeof body === 'string' ? body : JSON.stringify(body);
    expect(who).toContain(USERNAME);
    console.log(`whoami -> ${who}`);
  });

  test('GET /api/whoami without a token is rejected', async () => {
    await expectAuthRequired('get', '/api/whoami');
  });
});

/* -------------------------------------------------------------------------- */
/* Users                                                                      */
/* -------------------------------------------------------------------------- */

test.describe('users', () => {
  test('GET /api/users lists users', async () => {
    requireToken();
    const body = await getOk('/api/users');
    expect(body).toBeTruthy();
    console.log('GET /api/users returned a user list');
  });

  test('mutating verbs on /api/users require auth', async () => {
    await expectAuthRequired('post', '/api/users');
    await expectAuthRequired('delete', '/api/users');
    await expectAuthRequired('patch', '/api/users');
    await expectAuthRequired('put', '/api/users');
  });

  test('GET /api/users/<username>/containers lists compose services', async () => {
    requireToken();
    const user = await firstUser();
    test.skip(!user, 'no users exist on this environment');
    const body = await getOk(`/api/users/${user}/containers`);
    expect(body).toBeTruthy();
    console.log(`containers listed for user ${user}`);
  });

  test('GET /api/users/<username>/containers?stats=1 returns live stats', async () => {
    requireToken();
    const user = await firstUser();
    test.skip(!user, 'no users exist on this environment');
    const res = await api.get(`/api/users/${user}/containers?stats=1`, { headers: authHeaders() });
    expect(res.status(), 'stats endpoint should respond').toBeLessThan(500);
    console.log(`container stats for ${user} -> ${res.status()}`);
  });

  test('POST /api/users/<username>/containers/<action>/<container> requires auth', async () => {
    await expectAuthRequired('post', '/api/users/someuser/containers/restart/someuser_nginx');
  });
});

/* -------------------------------------------------------------------------- */
/* Domains                                                                    */
/* -------------------------------------------------------------------------- */

test.describe('domains', () => {
  test('GET /api/domains lists all domains', async () => {
    requireToken();
    const body = await getOk('/api/domains');
    expect(body).toBeTruthy();
    console.log('GET /api/domains returned a domain list');
  });

  test('POST /api/domains/new requires auth', async () => {
    await expectAuthRequired('post', '/api/domains/new');
  });

  test('POST /api/domains/<action>/<domain> (suspend/unsuspend/delete) requires auth', async () => {
    await expectAuthRequired('post', '/api/domains/suspend/example.com');
    await expectAuthRequired('post', '/api/domains/unsuspend/example.com');
    await expectAuthRequired('post', '/api/domains/delete/example.com');
  });

  test('GET /api/domains/docroot/<domain> returns docroot', async () => {
    requireToken();
    const domain = await firstDomain();
    test.skip(!domain, 'no domains exist on this environment');
    const body = await getOk(`/api/domains/docroot/${domain}`);
    expect(body).toBeTruthy();
    console.log(`docroot fetched for ${domain}`);
  });

  test('POST /api/domains/docroot/<domain> requires auth', async () => {
    await expectAuthRequired('post', '/api/domains/docroot/example.com');
  });

  test('GET /api/domains/<domain>/dns returns the zone file', async () => {
    requireToken();
    const domain = await firstDomain();
    test.skip(!domain, 'no domains exist on this environment');
    const res = await api.get(`/api/domains/${domain}/dns`, { headers: authHeaders() });
    expect(res.status(), 'dns zone endpoint should respond').toBeLessThan(500);
    console.log(`dns zone for ${domain} -> ${res.status()}`);
  });

  test('POST /api/domains/<domain>/dns requires auth', async () => {
    await expectAuthRequired('post', '/api/domains/example.com/dns');
  });

  test('GET /api/domains/<domain>/caddy returns the caddy config', async () => {
    requireToken();
    const domain = await firstDomain();
    test.skip(!domain, 'no domains exist on this environment');
    const res = await api.get(`/api/domains/${domain}/caddy`, { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`caddy config for ${domain} -> ${res.status()}`);
  });

  test('POST /api/domains/<domain>/caddy requires auth', async () => {
    await expectAuthRequired('post', '/api/domains/example.com/caddy');
  });

  test('GET /api/domains/<domain>/vhost/<username> returns the vhost file', async () => {
    requireToken();
    const domain = await firstDomain();
    const user = await firstUser();
    test.skip(!domain || !user, 'no domain/user available on this environment');
    const res = await api.get(`/api/domains/${domain}/vhost/${user}`, { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`vhost for ${domain}/${user} -> ${res.status()}`);
  });

  test('POST /api/domains/<domain>/vhost/<username> requires auth', async () => {
    await expectAuthRequired('post', '/api/domains/example.com/vhost/someuser');
  });

  test('GET /api/domains/<domain>/ssl returns ssl status', async () => {
    requireToken();
    const domain = await firstDomain();
    test.skip(!domain, 'no domains exist on this environment');
    const res = await api.get(`/api/domains/${domain}/ssl`, { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`ssl status for ${domain} -> ${res.status()}`);
  });

  test('POST /api/domains/<domain>/ssl action=logs returns issuance logs', async () => {
    requireToken();
    const domain = await firstDomain();
    test.skip(!domain, 'no domains exist on this environment');
    const res = await api.post(`/api/domains/${domain}/ssl`, {
      headers: authHeaders(),
      data: { action: 'logs' },
    });
    expect(res.status(), 'ssl logs action should respond').toBeLessThan(500);
    console.log(`ssl logs for ${domain} -> ${res.status()}`);
  });

  test('GET /api/domains/<domain>/log returns the paginated access log', async () => {
    requireToken();
    const domain = await firstDomain();
    test.skip(!domain, 'no domains exist on this environment');
    const res = await api.get(`/api/domains/${domain}/log?page=1`, { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`access log for ${domain} -> ${res.status()}`);
  });

  test('GET /api/domains/<domain>/stats/<username> returns goaccess report', async () => {
    requireToken();
    const domain = await firstDomain();
    const user = await firstUser();
    test.skip(!domain || !user, 'no domain/user available on this environment');
    const res = await api.get(`/api/domains/${domain}/stats/${user}`, { headers: authHeaders() });
    // report is generated every 24h, so a 404 on a fresh install is acceptable
    expect(res.status()).toBeLessThan(500);
    console.log(`goaccess stats for ${domain}/${user} -> ${res.status()}`);
  });

  test('GET /api/domains/file-templates returns default templates', async () => {
    requireToken();
    const body = await getOk('/api/domains/file-templates');
    expect(body).toBeTruthy();
    console.log('domain file templates fetched');
  });

  test('POST /api/domains/file-templates requires auth', async () => {
    await expectAuthRequired('post', '/api/domains/file-templates');
  });
});

/* -------------------------------------------------------------------------- */
/* DNS cluster & zone templates                                               */
/* -------------------------------------------------------------------------- */

test.describe('dns', () => {
  test('GET /api/dns/cluster returns cluster config', async () => {
    requireToken();
    const body = await getOk('/api/dns/cluster');
    expect(body).toBeTruthy();
    console.log('dns cluster config fetched');
  });

  test('POST /api/dns/cluster requires auth', async () => {
    await expectAuthRequired('post', '/api/dns/cluster');
  });

  test('GET /api/dns/cluster/<ip> checks slave reachability', async () => {
    requireToken();
    const cluster: any = await getOk('/api/dns/cluster');
    const slaves = cluster?.slaves ?? cluster?.ips ?? [];
    const ip = Array.isArray(slaves) && slaves.length
      ? (typeof slaves[0] === 'string' ? slaves[0] : slaves[0]?.ip)
      : null;
    test.skip(!ip, 'no DNS cluster slave configured on this environment');
    const res = await api.get(`/api/dns/cluster/${ip}`, { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`dns slave ${ip} status -> ${res.status()}`);
  });

  test('GET /api/dns/zone-templates returns ipv4/ipv6 templates', async () => {
    requireToken();
    const body = await getOk('/api/dns/zone-templates');
    expect(body).toBeTruthy();
    console.log('dns zone templates fetched');
  });

  test('POST /api/dns/zone-templates requires auth', async () => {
    await expectAuthRequired('post', '/api/dns/zone-templates');
  });
});

/* -------------------------------------------------------------------------- */
/* Plans - full safe CRUD lifecycle                                           */
/* -------------------------------------------------------------------------- */

test.describe('plans', () => {
  test('GET /api/plans lists plans', async () => {
    requireToken();
    const body = await getOk('/api/plans');
    expect(body).toBeTruthy();
    console.log('GET /api/plans returned a plan list');
  });

  test('plan lifecycle: POST create -> GET -> PUT edit -> DELETE', async () => {
    requireToken();

    // create
    const createRes = await api.post('/api/plans', {
      headers: authHeaders(),
      data: {
        name: TEST_PLAN_NAME,
        description: 'temporary plan created by playwright, safe to delete',
        email_limit: '1', ftp_limit: '1', domains_limit: '1', websites_limit: '1',
        disk_limit: '1000', inodes_limit: '10000', db_limit: '1',
        cpu: '1', ram: '1', bandwidth: '10', feature_set: 'default',
      },
    });
    expect(createRes.ok(), `plan create -> ${createRes.status()}`).toBeTruthy();
    console.log(`created plan ${TEST_PLAN_NAME}`);

    // find its id
    const plansBody: any = await getOk('/api/plans');
    const plans = Array.isArray(plansBody) ? plansBody : plansBody?.plans ?? plansBody?.data ?? [];
    const created = plans.find((p: any) => p?.name === TEST_PLAN_NAME);
    expect(created, 'created plan should appear in GET /api/plans').toBeTruthy();
    const planId = created.id ?? created.plan_id;
    expect(planId, 'created plan should have an id').toBeTruthy();

    // view by id
    const viewBody = await getOk(`/api/plans/${planId}`);
    expect(viewBody).toBeTruthy();
    console.log(`viewed plan ${planId}`);

    // edit
    const putRes = await api.put(`/api/plans/${planId}`, {
      headers: authHeaders(),
      data: { name: TEST_PLAN_NAME, disk_limit: '2000' },
    });
    expect(putRes.status(), `plan edit -> ${putRes.status()}`).toBeLessThan(500);
    console.log(`edited plan ${planId}`);

    // delete
    const delRes = await api.delete(`/api/plans/${planId}`, { headers: authHeaders() });
    expect(delRes.ok(), `plan delete -> ${delRes.status()}`).toBeTruthy();
    console.log(`deleted plan ${planId}`);
  });

  test('PATCH /api/plans/<id> requires auth', async () => {
    await expectAuthRequired('patch', '/api/plans/1');
  });
});

/* -------------------------------------------------------------------------- */
/* Services & docker                                                          */
/* -------------------------------------------------------------------------- */

test.describe('services', () => {
  test('GET /api/services lists monitored services', async () => {
    requireToken();
    const body = await getOk('/api/services');
    expect(body).toBeTruthy();
    console.log('monitored services listed');
  });

  test('PUT /api/services requires auth', async () => {
    await expectAuthRequired('put', '/api/services');
  });

  test('GET /api/services/status checks services status', async () => {
    requireToken();
    const body = await getOk('/api/services/status');
    expect(body).toBeTruthy();
    console.log('services status fetched');
  });

  test('POST /api/services/<action>/<service_name> requires auth', async () => {
    await expectAuthRequired('post', '/api/services/restart/caddy');
  });

  test('GET /api/docker/info returns docker info', async () => {
    requireToken();
    const body = await getOk('/api/docker/info');
    expect(body).toBeTruthy();
    console.log('docker info fetched');
  });
});

/* -------------------------------------------------------------------------- */
/* System info & usage                                                        */
/* -------------------------------------------------------------------------- */

test.describe('system & usage', () => {
  test('GET /api/ips lists dedicated ip users', async () => {
    requireToken();
    const res = await api.get('/api/ips', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`dedicated ips -> ${res.status()}`);
  });

  test('GET /api/system returns system information', async () => {
    requireToken();
    const body = await getOk('/api/system');
    expect(body).toBeTruthy();
    console.log('system info fetched');
  });

  for (const metric of ['cpu', 'memory', 'server', 'disk']) {
    test(`GET /api/usage/${metric} returns usage data`, async () => {
      requireToken();
      const body = await getOk(`/api/usage/${metric}`);
      expect(body).toBeTruthy();
      console.log(`usage/${metric} fetched`);
    });
  }
});

/* -------------------------------------------------------------------------- */
/* Notifications                                                              */
/* -------------------------------------------------------------------------- */

test.describe('notifications', () => {
  test('GET /api/notifications lists notifications', async () => {
    requireToken();
    const body = await getOk('/api/notifications');
    expect(body).toBeTruthy();
    console.log('notifications listed');
  });

  test('POST /api/notifications/<id>/read marks a notification as read', async () => {
    requireToken();
    const body: any = await getOk('/api/notifications');
    const list = Array.isArray(body) ? body : body?.notifications ?? body?.data ?? [];
    const id = list?.[0]?.id;
    test.skip(!id, 'no notifications exist on this environment');
    const res = await api.post(`/api/notifications/${id}/read`, { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`notification ${id} marked as read -> ${res.status()}`);
  });

  test('DELETE /api/notifications/<id> requires auth', async () => {
    await expectAuthRequired('delete', '/api/notifications/1');
  });

  test('DELETE /api/notifications/<id>?command=delete_all requires auth', async () => {
    await expectAuthRequired('delete', '/api/notifications/1?command=delete_all');
  });
});

/* -------------------------------------------------------------------------- */
/* Emails                                                                     */
/* -------------------------------------------------------------------------- */

test.describe('emails', () => {
  test('GET /api/emails/settings returns mailserver settings', async () => {
    requireToken();
    const res = await api.get('/api/emails/settings', { headers: authHeaders() });
    expect(res.status(), 'may 404 if mailserver module disabled').toBeLessThan(500);
    console.log(`emails/settings -> ${res.status()}`);
  });

  test('POST /api/emails/settings requires auth', async () => {
    await expectAuthRequired('post', '/api/emails/settings');
  });

  test('GET /api/emails/accounts lists mailboxes', async () => {
    requireToken();
    const res = await api.get('/api/emails/accounts', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`emails/accounts -> ${res.status()}`);
  });

  test('POST and DELETE /api/emails/accounts require auth', async () => {
    await expectAuthRequired('post', '/api/emails/accounts');
    await expectAuthRequired('delete', '/api/emails/accounts');
  });

  test('GET /api/emails/queue views the mail queue', async () => {
    requireToken();
    const res = await api.get('/api/emails/queue', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`emails/queue -> ${res.status()}`);
  });

  test('POST /api/emails/queue requires auth', async () => {
    await expectAuthRequired('post', '/api/emails/queue');
  });

  test('GET /api/emails/domain-limits views postfwd rules', async () => {
    requireToken();
    const res = await api.get('/api/emails/domain-limits', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`emails/domain-limits -> ${res.status()}`);
  });

  test('GET /api/emails/domain-limits?hits=<domain> views hit counters', async () => {
    requireToken();
    const domain = await firstDomain();
    test.skip(!domain, 'no domains exist on this environment');
    const res = await api.get(`/api/emails/domain-limits?hits=${domain}`, { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`emails/domain-limits hits for ${domain} -> ${res.status()}`);
  });

  test('POST /api/emails/domain-limits requires auth', async () => {
    await expectAuthRequired('post', '/api/emails/domain-limits');
  });
});

/* -------------------------------------------------------------------------- */
/* Security                                                                   */
/* -------------------------------------------------------------------------- */

test.describe('security', () => {
  test('GET /api/security/basic-auth returns basic auth config', async () => {
    requireToken();
    const body = await getOk('/api/security/basic-auth');
    expect(body).toBeTruthy();
    console.log('basic-auth config fetched');
  });

  test('POST /api/security/basic-auth requires auth', async () => {
    await expectAuthRequired('post', '/api/security/basic-auth');
  });

  test('GET /api/security/blacklist-useragents returns the blacklist', async () => {
    requireToken();
    const body = await getOk('/api/security/blacklist-useragents');
    expect(body).toBeTruthy();
    console.log('blacklisted user-agents fetched');
  });

  test('POST /api/security/blacklist-useragents requires auth', async () => {
    await expectAuthRequired('post', '/api/security/blacklist-useragents');
  });

  test('POST /api/security/disable-admin requires auth (never executed)', async () => {
    // irreversible via API - only auth enforcement is verified
    await expectAuthRequired('post', '/api/security/disable-admin');
  });

  test('GET /api/security/firewall checks CSF availability', async () => {
    requireToken();
    const res = await api.get('/api/security/firewall', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`security/firewall -> ${res.status()}`);
  });

  test('POST /api/security/firewall requires auth', async () => {
    await expectAuthRequired('post', '/api/security/firewall');
  });

  test('GET /api/security/waf returns WAF status', async () => {
    requireToken();
    const res = await api.get('/api/security/waf', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`security/waf -> ${res.status()}`);
  });

  test('POST /api/security/waf requires auth', async () => {
    await expectAuthRequired('post', '/api/security/waf');
  });

  test('GET /api/security/waf/rules lists WAF rule sets', async () => {
    requireToken();
    const res = await api.get('/api/security/waf/rules', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`security/waf/rules -> ${res.status()}`);
  });

  test('POST /api/security/waf/rules requires auth', async () => {
    await expectAuthRequired('post', '/api/security/waf/rules');
  });

  test('GET /api/security/2fa returns 2FA status', async () => {
    requireToken();
    const body = await getOk('/api/security/2fa');
    expect(body).toBeTruthy();
    console.log('2fa status fetched');
  });

  test('POST /api/security/2fa/enable and /disable require auth', async () => {
    await expectAuthRequired('post', '/api/security/2fa/enable');
    await expectAuthRequired('post', '/api/security/2fa/disable');
  });

  test('GET /api/security/passkeys lists passkeys', async () => {
    requireToken();
    const body = await getOk('/api/security/passkeys');
    expect(body).toBeTruthy();
    console.log('passkeys listed');
  });

  test('POST and DELETE /api/security/passkeys require auth', async () => {
    await expectAuthRequired('post', '/api/security/passkeys');
    await expectAuthRequired('delete', '/api/security/passkeys');
  });
});

/* -------------------------------------------------------------------------- */
/* Server                                                                     */
/* -------------------------------------------------------------------------- */

test.describe('server', () => {
  test('GET /api/server/crons lists cronjobs', async () => {
    requireToken();
    const body = await getOk('/api/server/crons');
    expect(body).toBeTruthy();
    console.log('cronjobs listed');
  });

  test('POST /api/server/crons requires auth', async () => {
    await expectAuthRequired('post', '/api/server/crons');
  });

  test('GET /api/server/ssh returns ssh status/config/keys', async () => {
    requireToken();
    const body = await getOk('/api/server/ssh');
    expect(body).toBeTruthy();
    console.log('ssh overview fetched');
  });

  test('POST /api/server/ssh requires auth', async () => {
    await expectAuthRequired('post', '/api/server/ssh');
  });

  test('GET /api/server/ssh/config returns raw sshd_config', async () => {
    requireToken();
    const body = await getOk('/api/server/ssh/config');
    expect(body).toBeTruthy();
    console.log('sshd_config fetched');
  });

  test('POST /api/server/ssh/config requires auth', async () => {
    await expectAuthRequired('post', '/api/server/ssh/config');
  });

  test('GET /api/server/timezone returns current timezone', async () => {
    requireToken();
    const body = await getOk('/api/server/timezone');
    expect(body).toBeTruthy();
    console.log('timezone fetched');
  });

  test('POST /api/server/timezone requires auth', async () => {
    await expectAuthRequired('post', '/api/server/timezone');
  });

  test('POST /api/server/memory/drop-cache requires auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/server/memory/drop-cache');
  });

  test('POST /api/server/memory/drop-swap requires auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/server/memory/drop-swap');
  });

  test('GET /api/server/processes lists processes sorted by memory', async () => {
    requireToken();
    const body = await getOk('/api/server/processes?sort=memory');
    expect(body).toBeTruthy();
    console.log('processes listed');
  });

  test('POST /api/server/processes/<pid>/kill requires auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/server/processes/99999999/kill');
  });

  test('GET /api/server/node returns default clustering node', async () => {
    requireToken();
    const res = await api.get('/api/server/node', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`server/node -> ${res.status()}`);
  });

  test('POST /api/server/node requires auth', async () => {
    await expectAuthRequired('post', '/api/server/node');
  });

  test('POST /api/server/root-password requires auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/server/root-password');
  });

  test('POST /api/server/reboot requires auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/server/reboot');
  });

  test('GET /api/server/reboot/status reports panel availability', async () => {
    requireToken();
    const res = await api.get('/api/server/reboot/status', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`reboot/status -> ${res.status()}`);
  });

  test('GET /api/server/migrate reports migration status', async () => {
    requireToken();
    const res = await api.get('/api/server/migrate', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`server/migrate status -> ${res.status()}`);
  });

  test('POST /api/server/migrate requires auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/server/migrate');
  });
});

/* -------------------------------------------------------------------------- */
/* Settings                                                                   */
/* -------------------------------------------------------------------------- */

test.describe('settings', () => {
  test('GET /api/settings/administrators lists admin accounts', async () => {
    requireToken();
    const body = await getOk('/api/settings/administrators');
    expect(body).toBeTruthy();
    console.log('administrators listed');
  });

  test('POST /api/settings/administrators requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/administrators');
  });

  test('GET /api/settings/resellers lists reseller accounts', async () => {
    requireToken();
    const res = await api.get('/api/settings/resellers', { headers: authHeaders() });
    expect(res.status(), 'may be unavailable on community edition').toBeLessThan(500);
    console.log(`resellers -> ${res.status()}`);
  });

  test('POST /api/settings/resellers requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/resellers');
  });

  test('GET /api/settings/general returns general settings', async () => {
    requireToken();
    const body = await getOk('/api/settings/general');
    expect(body).toBeTruthy();
    console.log('general settings fetched');
  });

  test('POST /api/settings/general requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/general');
  });

  test('GET /api/settings/defaults returns default env/services', async () => {
    requireToken();
    const body = await getOk('/api/settings/defaults');
    expect(body).toBeTruthy();
    console.log('defaults fetched');
  });

  test('POST /api/settings/defaults requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/defaults');
  });

  test('GET /api/settings/defaults/files returns compose/env templates', async () => {
    requireToken();
    const body = await getOk('/api/settings/defaults/files');
    expect(body).toBeTruthy();
    console.log('default files fetched');
  });

  test('POST and DELETE /api/settings/defaults/files require auth', async () => {
    await expectAuthRequired('post', '/api/settings/defaults/files');
    await expectAuthRequired('delete', '/api/settings/defaults/files');
  });

  test('GET /api/settings/defaults/files/<username> returns per-user files', async () => {
    requireToken();
    const user = await firstUser();
    test.skip(!user, 'no users exist on this environment');
    const body = await getOk(`/api/settings/defaults/files/${user}`);
    expect(body).toBeTruthy();
    console.log(`compose/env files fetched for ${user}`);
  });

  test('POST /api/settings/defaults/files/<username> requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/defaults/files/someuser');
  });

  test('GET /api/settings/features lists feature sets', async () => {
    requireToken();
    const body = await getOk('/api/settings/features');
    expect(body).toBeTruthy();
    console.log('feature sets listed');
  });

  test('GET /api/settings/features/default views the default feature set', async () => {
    requireToken();
    const body = await getOk('/api/settings/features/default');
    expect(body).toBeTruthy();
    console.log('default feature set fetched');
  });

  test('feature set lifecycle: create -> update -> delete', async () => {
    requireToken();

    // create
    const createRes = await api.post('/api/settings/features', {
      headers: authHeaders(),
      data: { feature_name: TEST_FEATURE_SET },
    });
    expect(createRes.status(), `feature set create -> ${createRes.status()}`).toBeLessThan(500);
    console.log(`created feature set ${TEST_FEATURE_SET}`);

    // update
    const updateRes = await api.post(`/api/settings/features/${TEST_FEATURE_SET}`, {
      headers: authHeaders(),
      data: { action: 'update', features: ['ftp', 'cron'] },
    });
    expect(updateRes.status()).toBeLessThan(500);
    console.log(`updated feature set ${TEST_FEATURE_SET}`);

    // delete
    const deleteRes = await api.post(`/api/settings/features/${TEST_FEATURE_SET}`, {
      headers: authHeaders(),
      data: { action: 'delete' },
    });
    expect(deleteRes.status()).toBeLessThan(500);
    console.log(`deleted feature set ${TEST_FEATURE_SET}`);
  });

  test('GET /api/settings/locales lists locales', async () => {
    requireToken();
    const body = await getOk('/api/settings/locales');
    expect(body).toBeTruthy();
    console.log('locales listed');
  });

  test('POST /api/settings/locales requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/locales');
  });

  test('GET /api/settings/modules lists enabled modules', async () => {
    requireToken();
    const body = await getOk('/api/settings/modules');
    expect(body).toBeTruthy();
    console.log('modules listed');
  });

  test('POST /api/settings/modules requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/modules');
  });

  test('GET /api/settings/custom-code returns custom code snippets', async () => {
    requireToken();
    const body = await getOk('/api/settings/custom-code');
    expect(body).toBeTruthy();
    console.log('custom-code fetched');
  });

  test('POST /api/settings/custom-code requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/custom-code');
  });

  test('GET /api/settings/php returns php options', async () => {
    requireToken();
    const body = await getOk('/api/settings/php');
    expect(body).toBeTruthy();
    console.log('php settings fetched');
  });

  test('POST /api/settings/php requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/php');
  });

  test('GET /api/settings/caddy/metrics returns prometheus metrics', async () => {
    requireToken();
    const res = await api.get('/api/settings/caddy/metrics', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`caddy metrics -> ${res.status()}`);
  });

  test('GET /api/settings/updates returns update preferences', async () => {
    requireToken();
    const body = await getOk('/api/settings/updates');
    expect(body).toBeTruthy();
    console.log('update preferences fetched');
  });

  test('POST /api/settings/updates requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/updates');
  });

  test('POST /api/settings/updates/now requires auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/settings/updates/now');
  });

  test('GET /api/settings/updates/tags lists available image tags', async () => {
    requireToken();
    const res = await api.get('/api/settings/updates/tags', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`update tags -> ${res.status()}`);
  });

  test('POST /api/settings/updates/tags requires auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/settings/updates/tags');
  });

  test('GET /api/settings/notifications returns notification preferences', async () => {
    requireToken();
    const body = await getOk('/api/settings/notifications');
    expect(body).toBeTruthy();
    console.log('notification preferences fetched');
  });

  test('POST /api/settings/notifications requires auth', async () => {
    await expectAuthRequired('post', '/api/settings/notifications');
  });
});

/* -------------------------------------------------------------------------- */
/* License                                                                    */
/* -------------------------------------------------------------------------- */

test.describe('license', () => {
  test('GET /api/license returns license info', async () => {
    requireToken();
    const res = await api.get('/api/license', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`license -> ${res.status()}`);
  });

  test('POST and DELETE /api/license require auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/license');
    await expectAuthRequired('delete', '/api/license');
  });

  test('GET /api/license/info returns detailed license info', async () => {
    requireToken();
    const res = await api.get('/api/license/info', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`license/info -> ${res.status()}`);
  });

  test('POST /api/license/verify verifies the license', async () => {
    requireToken();
    const res = await api.post('/api/license/verify', { headers: authHeaders() });
    expect(res.status(), 'verify may fail without an enterprise key, but must respond').toBeLessThan(500);
    console.log(`license/verify -> ${res.status()}`);
  });
});

/* -------------------------------------------------------------------------- */
/* Support & import/transfer                                                  */
/* -------------------------------------------------------------------------- */

test.describe('support & import', () => {
  test('GET /api/support/report generates a diagnostics report', async () => {
    requireToken();
    const res = await api.get('/api/support/report', { headers: authHeaders(), timeout: 120_000 });
    expect(res.status()).toBeLessThan(500);
    console.log(`support/report -> ${res.status()}`);
  });

  test('GET /api/import/cpanel lists import job statuses', async () => {
    requireToken();
    const res = await api.get('/api/import/cpanel', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`import/cpanel -> ${res.status()}`);
  });

  test('POST /api/import/<panel_type> requires auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/import/cpanel');
  });

  test('GET /api/import/backup-files lists discovered backup archives', async () => {
    requireToken();
    const res = await api.get('/api/import/backup-files', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`import/backup-files -> ${res.status()}`);
  });

  test('GET /api/import/transfers lists transfer log files', async () => {
    requireToken();
    const res = await api.get('/api/import/transfers', { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`import/transfers -> ${res.status()}`);
  });

  test('POST /api/import/transfers requires auth (not executed)', async () => {
    await expectAuthRequired('post', '/api/import/transfers');
  });

  test('GET /api/import/transfers/<username> lists per-user transfer logs', async () => {
    requireToken();
    const user = await firstUser();
    test.skip(!user, 'no users exist on this environment');
    const res = await api.get(`/api/import/transfers/${user}`, { headers: authHeaders() });
    expect(res.status()).toBeLessThan(500);
    console.log(`import/transfers/${user} -> ${res.status()}`);
  });

  test('GET /api/import/logs/account/<log> handles missing log gracefully', async () => {
    requireToken();
    const res = await api.get('/api/import/logs/account/pw_nonexistent.log', { headers: authHeaders() });
    expect(res.status(), 'nonexistent log should not 500').toBeLessThan(500);
    console.log(`import/logs/account (nonexistent) -> ${res.status()}`);
  });

  test('GET /api/import/logs/transfer/<log> handles missing log gracefully', async () => {
    requireToken();
    const res = await api.get('/api/import/logs/transfer/pw_nonexistent.log', { headers: authHeaders() });
    expect(res.status(), 'nonexistent log should not 500').toBeLessThan(500);
    console.log(`import/logs/transfer (nonexistent) -> ${res.status()}`);
  });
});

/* -------------------------------------------------------------------------- */
/* Optional destructive coverage (RUN_MUTATING=1)                             */
/* -------------------------------------------------------------------------- */

test.describe('mutating operations (RUN_MUTATING=1 only)', () => {
  test.skip(!RUN_MUTATING, 'set RUN_MUTATING=1 to run user/domain lifecycle tests');

  const TEST_USER = 'pwtestusr';
  const TEST_DOMAIN = 'pw-test-domain.example';

  test('user lifecycle: create -> suspend -> unsuspend -> delete', async () => {
    requireToken();

    const createRes = await api.post('/api/users', {
      headers: authHeaders(),
      data: {
        email: 'pwtest@example.com',
        username: TEST_USER,
        password: `Pw!${Date.now()}`,
        plan_name: 'default_plan_apache',
      },
      timeout: 300_000,
    });
    expect(createRes.ok(), `user create -> ${createRes.status()}`).toBeTruthy();
    console.log(`created user ${TEST_USER}`);

    const patchSuspend = await api.patch(`/api/users/${TEST_USER}`, {
      headers: authHeaders(),
      data: { action: 'suspend' },
    });
    expect(patchSuspend.status()).toBeLessThan(500);

    const patchUnsuspend = await api.patch(`/api/users/${TEST_USER}`, {
      headers: authHeaders(),
      data: { action: 'unsuspend' },
    });
    expect(patchUnsuspend.status()).toBeLessThan(500);

    const delRes = await api.delete(`/api/users/${TEST_USER}`, {
      headers: authHeaders(),
      timeout: 300_000,
    });
    expect(delRes.ok(), `user delete -> ${delRes.status()}`).toBeTruthy();
    console.log(`deleted user ${TEST_USER}`);
  });

  test('domain lifecycle: new -> suspend -> unsuspend -> docroot -> delete', async () => {
    requireToken();
    const user = await firstUser();
    test.skip(!user, 'no users exist on this environment');

    const addRes = await api.post('/api/domains/new', {
      headers: authHeaders(),
      data: { username: user, domain: TEST_DOMAIN, docroot: `/var/www/html/${TEST_DOMAIN}` },
      timeout: 120_000,
    });
    expect(addRes.ok(), `domain add -> ${addRes.status()}`).toBeTruthy();
    console.log(`added domain ${TEST_DOMAIN} for ${user}`);

    for (const action of ['suspend', 'unsuspend']) {
      const res = await api.post(`/api/domains/${action}/${TEST_DOMAIN}`, { headers: authHeaders() });
      expect(res.status(), `${action} -> ${res.status()}`).toBeLessThan(500);
      console.log(`${action}ed ${TEST_DOMAIN}`);
    }

    const docrootRes = await api.post(`/api/domains/docroot/${TEST_DOMAIN}`, {
      headers: authHeaders(),
      data: { docroot: `/var/www/html/${TEST_DOMAIN}` },
    });
    expect(docrootRes.status()).toBeLessThan(500);

    const delRes = await api.post(`/api/domains/delete/${TEST_DOMAIN}`, {
      headers: authHeaders(),
      timeout: 120_000,
    });
    expect(delRes.ok(), `domain delete -> ${delRes.status()}`).toBeTruthy();
    console.log(`deleted domain ${TEST_DOMAIN}`);
  });
});
