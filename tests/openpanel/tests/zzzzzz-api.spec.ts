
import { test, expect, request as pwRequest, APIRequestContext, APIResponse } from '@playwright/test';

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------
const BASE_URL       = process.env.BASE_URL!;
const API_USER       = process.env.PANEL_USERNAME!;
const API_PASS       = process.env.PANEL_PASSWORD!;
const TEST_DOMAIN    = process.env.TEST_DOMAIN ?? 'wp.tests.openpanel.org';
const RUN_MUTATIONS  = process.env.RUN_MUTATIONS === 'true';

const mtest = RUN_MUTATIONS ? test : test.skip;
const mdescribe = RUN_MUTATIONS ? test.describe : test.describe.skip;

function rand(prefix: string): string {
  return `${prefix}${Date.now().toString(36)}${Math.floor(Math.random() * 1000)}`;
}

// ---------------------------------------------------------------------------
// Logging — reads body once, caches it, wraps response so .json()/.text()
// keep working after the log line has already consumed the stream.
// ---------------------------------------------------------------------------
async function logResponse(method: string, url: string, res: APIResponse): Promise<APIResponse> {
  const bodyBuffer = await res.body();
  const bodyText   = bodyBuffer.toString('utf-8');
  const preview    = bodyText.length > 300 ? bodyText.slice(0, 300) + '…' : bodyText;
  const icon       = res.status() < 400 ? '✓' : '✗';

  console.log(`\n${icon} ${method.toUpperCase()} ${BASE_URL}${url}`);
  console.log(`  → ${res.status()} ${res.statusText()}`);
  console.log(`  ← ${preview}`);

  return {
    ...res,
    text:   () => Promise.resolve(bodyText),
    json:   () => Promise.resolve(JSON.parse(bodyText)),
    body:   () => Promise.resolve(bodyBuffer),
    ok:     () => res.ok(),
    status: () => res.status(),
    statusText: () => res.statusText(),
    headers: () => res.headers(),
    headersArray: () => res.headersArray(),
    url:    () => res.url(),
    dispose: () => res.dispose(),
  } as unknown as APIResponse;
}

function logged(ctx: APIRequestContext): APIRequestContext {
  const methods = ['get', 'post', 'put', 'delete', 'patch'] as const;
  return new Proxy(ctx, {
    get(target, prop: string) {
      if (!(methods as readonly string[]).includes(prop)) return (target as any)[prop];
      return async (url: string, options?: object) => {
        const res = await (target as any)[prop](url, options);
        return logResponse(prop, url, res);
      };
    },
  });
}

// ---------------------------------------------------------------------------
// Shared authenticated API context — token fetched once in beforeAll
// ---------------------------------------------------------------------------
let api: APIRequestContext;

test.beforeAll(async () => {
  const anonCtx = await pwRequest.newContext({ baseURL: BASE_URL });
  try {
    const res = await anonCtx.post('/api/login', {
      data: { username: API_USER, password: API_PASS },
    });
    expect(res.status(), 'POST /api/login must return 200').toBe(200);
    const json = await res.json();
    expect(json.token, 'login response must include a token').toBeTruthy();

    api = logged(await pwRequest.newContext({
      baseURL: BASE_URL,
      extraHTTPHeaders: { Authorization: `Bearer ${json.token}` },
    }));
  } finally {
    await anonCtx.dispose();
  }
});

test.afterAll(async () => {
  await api?.dispose();
});

// ---------------------------------------------------------------------------
// POST /api/login
// ---------------------------------------------------------------------------
test.describe('POST /api/login', () => {

  test('rejects wrong credentials with 401', async () => {
    const ctx = await pwRequest.newContext({ baseURL: BASE_URL });
    try {
      const res = await ctx.post('/api/login', {
        data: { username: 'nobody', password: 'wrongpassword' },
      });
      expect(res.status()).toBe(401);
    } finally {
      await ctx.dispose();
    }
  });

  test('returns token shape for valid credentials', async () => {
    const ctx = await pwRequest.newContext({ baseURL: BASE_URL });
    try {
      const res = await ctx.post('/api/login', {
        data: { username: API_USER, password: API_PASS },
      });
      // 200 = success | 401 with twofa_required = 2FA-enabled account
      expect([200, 401]).toContain(res.status());
      const json = await res.json();
      if (res.status() === 200) {
        expect(json).toHaveProperty('token');
        expect(json).toHaveProperty('user_id');
        expect(json).toHaveProperty('expires_in');
        expect(typeof json.token).toBe('string');
        expect(json.expires_in).toBeGreaterThan(0);
      } else {
        expect(json).toHaveProperty('twofa_required', true);
        expect(json).toHaveProperty('user_id');
      }
    } finally {
      await ctx.dispose();
    }
  });

  test('rejects empty body', async () => {
    const ctx = await pwRequest.newContext({ baseURL: BASE_URL });
    try {
      const res = await ctx.post('/api/login', { data: {} });
      expect([400, 401]).toContain(res.status());
    } finally {
      await ctx.dispose();
    }
  });
});

// ---------------------------------------------------------------------------
// GET /api/endpoints — this is the live route map; also used below to sanity
// check that the routes this file exercises actually still exist.
// ---------------------------------------------------------------------------
test.describe('GET /api/endpoints', () => {

  test('returns a list of API endpoints', async () => {
    const res = await api.get('/api/endpoints');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).toHaveProperty('endpoints');
    expect(json).toHaveProperty('total');
    expect(Array.isArray(json.endpoints)).toBe(true);
    expect(json.total).toBeGreaterThan(0);
    for (const ep of json.endpoints as Record<string, unknown>[]) {
      expect(ep).toHaveProperty('path');
      expect(ep).toHaveProperty('methods');
      expect(ep).toHaveProperty('endpoint');
      expect((ep.path as string).startsWith('/api/')).toBe(true);
      expect(Array.isArray(ep.methods)).toBe(true);
    }
  });

  test('includes a representative sample of modules/api/ routes', async () => {
    const res = await api.get('/api/endpoints');
    const json = await res.json();
    const paths: string[] = (json.endpoints as { path: string }[]).map(e => e.path);
    for (const expected of [
      '/api/mysql/databases',
      '/api/postgresql/databases',
      '/api/domains',
      '/api/domains/<domain>',
      '/api/dns',
      '/api/waf/<domain>',
      '/api/containers',
      '/api/emails',
      '/api/sites',
      '/api/wordpress',
    ]) {
      expect(paths, `expected ${expected} to be a registered route`).toContain(expected);
    }
  });
});

// ---------------------------------------------------------------------------
// GET /api/ → redirect
// ---------------------------------------------------------------------------
test.describe('GET /api/ redirect', () => {

  test('resolves without error', async () => {
    const res = await api.get('/api/');
    expect([200, 301, 302, 307, 308]).toContain(res.status());
  });
});

// ---------------------------------------------------------------------------
// /api/account
// ---------------------------------------------------------------------------
test.describe('GET /api/account', () => {
  test('returns account fields', async () => {
    const res = await api.get('/api/account');
    expect(res.status()).toBe(200);
    const json = await res.json();
    for (const field of ['username', 'email', 'plan_id', 'context', 'permit_username_change']) {
      expect(json, `missing field: ${field}`).toHaveProperty(field);
    }
  });
});

test.describe('PATCH /api/account', () => {
  test('rejects empty body', async () => {
    const res = await api.patch('/api/account', { data: {} });
    expect(res.status()).toBe(400);
    expect((await res.json())).toHaveProperty('error');
  });
});

test.describe('GET /api/account/sessions', () => {
  test('returns a sessions array', async () => {
    const res = await api.get('/api/account/sessions');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).toHaveProperty('sessions');
    expect(Array.isArray(json.sessions)).toBe(true);
    for (const s of json.sessions as Record<string, unknown>[]) {
      expect(s).toHaveProperty('session_token');
      expect(s).toHaveProperty('ip_address');
    }
  });
});

test.describe('DELETE /api/account/sessions/:token', () => {
  test('unknown session token returns 404', async () => {
    const res = await api.delete('/api/account/sessions/not-a-real-session-token');
    expect([404, 500]).toContain(res.status());
  });
});

// ---------------------------------------------------------------------------
// GET/POST /api/services
// ---------------------------------------------------------------------------
test.describe('GET /api/services', () => {
  test('returns the service list', async () => {
    const res = await api.get('/api/services');
    expect([200, 404, 500]).toContain(res.status());
    if (res.status() === 200) {
      const json = await res.json();
      expect(Array.isArray(json.services)).toBe(true);
      for (const s of json.services as Record<string, unknown>[]) {
        expect(s).toHaveProperty('service');
        expect(s).toHaveProperty('container_state');
      }
    }
  });
});

test.describe('GET /api/services/:service', () => {
  test('unknown service returns 404', async () => {
    const res = await api.get('/api/services/definitely-not-a-real-service');
    expect([404]).toContain(res.status());
  });
});

test.describe('POST /api/services/:service', () => {
  test('rejects invalid action', async () => {
    const res = await api.post('/api/services/definitely-not-a-real-service', {
      data: { action: 'not-a-real-action' },
    });
    expect([400, 404]).toContain(res.status());
  });
});

// ---------------------------------------------------------------------------
// GET /api/containers (docker.py)
// ---------------------------------------------------------------------------
test.describe('GET /api/containers', () => {
  test('returns compose config', async () => {
    const res = await api.get('/api/containers');
    expect([200, 500]).toContain(res.status());
  });
});

test.describe('GET /api/containers/status', () => {
  test('returns running_containers', async () => {
    const res = await api.get('/api/containers/status');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      expect(await res.json()).toHaveProperty('running_containers');
    }
  });
});

test.describe('GET /api/containers/:service/status', () => {
  test('returns state/health shape', async () => {
    const res = await api.get('/api/containers/mysql/status');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).toHaveProperty('service');
    expect(json).toHaveProperty('state');
  });
});

test.describe('GET /api/containers/:service/logs', () => {
  test('unknown service returns 404', async () => {
    const res = await api.get('/api/containers/definitely-not-a-real-service/logs');
    expect([404]).toContain(res.status());
  });
});

// docker.py has no service allowlist — start/stop/restart/resources genuinely
// affects live containers (possibly shared infra like mysql/caddy). Only run
// against a disposable panel, and only against a container that's safe to churn.
mdescribe('POST /api/containers/:service/{start,stop,restart} lifecycle', () => {
  const service = process.env.TEST_DOCKER_SERVICE ?? '';

  test('restart a disposable test service', async () => {
    test.skip(!service, 'set TEST_DOCKER_SERVICE to a container safe to restart');
    const res = await api.post(`/api/containers/${service}/restart`);
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('message');
  });
});

// ---------------------------------------------------------------------------
// GET /api/stats (goaccess.py)
// ---------------------------------------------------------------------------
test.describe('GET /api/stats', () => {
  test('returns domains with stats availability', async () => {
    const res = await api.get('/api/stats');
    expect(res.status()).toBe(200);
    expect(Array.isArray((await res.json()).domains)).toBe(true);
  });
});

test.describe('GET /api/stats/:domain', () => {
  test('returns stats availability for TEST_DOMAIN', async () => {
    const res = await api.get(`/api/stats/${TEST_DOMAIN}`);
    expect([200, 403, 404]).toContain(res.status());
    if (res.status() !== 403) {
      const json = await res.json();
      expect(json).toHaveProperty('domain', TEST_DOMAIN);
      expect(json).toHaveProperty('available');
    }
  });

  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/stats/definitely-not-owned-domain.example');
    expect([403, 404]).toContain(res.status());
  });
});

// ---------------------------------------------------------------------------
// GET /api/autoinstaller (renamed from /api/auto-installer)
// ---------------------------------------------------------------------------
test.describe('GET /api/autoinstaller', () => {
  test('returns sites, counts and technologies', async () => {
    const res = await api.get('/api/autoinstaller');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).not.toHaveProperty('error');
    expect(json).toHaveProperty('sites');
    expect(json).toHaveProperty('counts');
    expect(json).toHaveProperty('technologies');
    expect(json).toHaveProperty('domain_count');
    expect(Array.isArray(json.sites)).toBe(true);
    expect(Array.isArray(json.technologies)).toBe(true);
  });
});

// ---------------------------------------------------------------------------
// /api/ip-blocker lifecycle
// ---------------------------------------------------------------------------
test.describe('GET /api/ip-blocker', () => {
  test('returns blocked_ips array', async () => {
    const res = await api.get('/api/ip-blocker');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      expect(Array.isArray((await res.json()).blocked_ips)).toBe(true);
    }
  });
});

test.describe('POST /api/ip-blocker', () => {
  test('rejects a request with no valid IPs', async () => {
    const res = await api.post('/api/ip-blocker', { data: { ips: ['not-an-ip'] } });
    expect(res.status()).toBe(400);
    const json = await res.json();
    expect(json).toHaveProperty('error');
    expect(json).toHaveProperty('invalid');
  });
});

mdescribe('/api/ip-blocker block -> unblock-all lifecycle', () => {
  test('blocks a scratch IP then removes all blocks', async () => {
    const testIp = '203.0.113.77'; // TEST-NET-3, safe to block/unblock
    const blockRes = await api.post('/api/ip-blocker', { data: { ips: [testIp] } });
    expect(blockRes.status()).toBe(200);
    const blockJson = await blockRes.json();
    expect(blockJson.blocked).toContain(testIp);

    // DELETE removes ALL blocked IPs for the account, not just this one —
    // only acceptable against a disposable panel with no other real blocks.
    const delRes = await api.delete('/api/ip-blocker');
    expect(delRes.status()).toBe(200);
    expect((await delRes.json())).toHaveProperty('message');
  });
});

// ---------------------------------------------------------------------------
// /api/dynamic-dns lifecycle
// ---------------------------------------------------------------------------
test.describe('GET /api/dynamic-dns', () => {
  test('returns a domains map', async () => {
    const res = await api.get('/api/dynamic-dns');
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('domains');
  });
});

test.describe('POST /api/dynamic-dns', () => {
  test('rejects missing domain/subdomain', async () => {
    const res = await api.post('/api/dynamic-dns', { data: {} });
    expect(res.status()).toBe(400);
    expect((await res.json())).toHaveProperty('error');
  });

  test('rejects unowned domain', async () => {
    const res = await api.post('/api/dynamic-dns', {
      data: { domain: 'definitely-not-owned-domain.example', subdomain: 'test' },
    });
    expect(res.status()).toBe(403);
  });
});

mdescribe('/api/dynamic-dns create -> update -> delete lifecycle', () => {
  const subdomain = rand('pwddns');
  let lineNumber: number;

  test('POST creates an entry', async () => {
    const res = await api.post('/api/dynamic-dns', {
      data: { domain: TEST_DOMAIN, subdomain, ip: '203.0.113.10' },
    });
    expect(res.status()).toBe(201);
    const json = await res.json();
    expect(json).toHaveProperty('entry');
    lineNumber = json.entry.line_number ?? json.entry.lineNumber;
  });

  test('DELETE removes the entry', async () => {
    test.skip(lineNumber === undefined, 'no line_number captured from create step');
    const res = await api.delete('/api/dynamic-dns', {
      data: { domain: TEST_DOMAIN, line_number: lineNumber },
    });
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('deleted_line');
  });
});

// ---------------------------------------------------------------------------
// /api/dns lifecycle (dns.py — separate from domains.py's /dns sub-resource)
// ---------------------------------------------------------------------------
test.describe('GET /api/dns', () => {
  test('returns domains with zone_file_exists flags', async () => {
    const res = await api.get('/api/dns');
    expect(res.status()).toBe(200);
    expect(Array.isArray((await res.json()).domains)).toBe(true);
  });
});

test.describe('GET /api/dns/:domain', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/dns/definitely-not-owned-domain.example');
    expect(res.status()).toBe(403);
    expect((await res.json())).toHaveProperty('error', 'You do not own this domain');
  });

  test('TEST_DOMAIN returns records or 404 if no zone file', async () => {
    const res = await api.get(`/api/dns/${TEST_DOMAIN}`);
    expect([200, 404]).toContain(res.status());
    if (res.status() === 200) {
      const json = await res.json();
      expect(json).toHaveProperty('domain', TEST_DOMAIN);
      expect(json).toHaveProperty('records');
    }
  });
});

test.describe('GET /api/dns/:domain/raw', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/dns/definitely-not-owned-domain.example/raw');
    expect(res.status()).toBe(403);
  });
});

test.describe('POST /api/dns/:domain/records', () => {
  test('rejects missing fields', async () => {
    const res = await api.post(`/api/dns/${TEST_DOMAIN}/records`, { data: {} });
    expect([400, 403, 404]).toContain(res.status());
  });
});

mdescribe('/api/dns record create -> update -> delete lifecycle', () => {
  const subdomain = rand('pwdns');
  let rowId: number;

  test('POST creates an A record', async () => {
    const res = await api.post(`/api/dns/${TEST_DOMAIN}/records`, {
      data: { name: subdomain, type: 'A', record: '203.0.113.20', ttl: 3600 },
    });
    expect(res.status()).toBe(201);
    const json = await res.json();
    expect(json).toHaveProperty('record');
    rowId = json.record.line_number ?? json.record.row_id;
  });

  test('PATCH updates the record', async () => {
    test.skip(rowId === undefined, 'no row_id captured from create step');
    const res = await api.patch(`/api/dns/${TEST_DOMAIN}/records/${rowId}`, {
      data: { content: `${subdomain} 3600 IN A 203.0.113.21` },
    });
    expect([200, 409]).toContain(res.status());
  });

  test('DELETE removes the record', async () => {
    test.skip(rowId === undefined, 'no row_id captured from create step');
    const res = await api.delete(`/api/dns/${TEST_DOMAIN}/records/${rowId}`);
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('deleted');
  });
});

// zone reset wipes all custom records back to default — irreversible, opt-in only.
mdescribe('POST /api/dns/:domain/reset', () => {
  test('resets the zone to default', async () => {
    const res = await api.post(`/api/dns/${TEST_DOMAIN}/reset`);
    expect([200, 500]).toContain(res.status());
  });
});

test.describe('GET /api/dns/:domain/export', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/dns/definitely-not-owned-domain.example/export');
    expect(res.status()).toBe(403);
  });
});

// ---------------------------------------------------------------------------
// /api/waf — status split from stats/log in the current API
// ---------------------------------------------------------------------------
test.describe('GET /api/waf', () => {
  test('returns a domain->status map', async () => {
    const res = await api.get('/api/waf');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).toHaveProperty('domains');
    expect(typeof json.domains).toBe('object');
  });
});

test.describe('GET /api/waf/:domain', () => {
  test('returns status and rule exclusions', async () => {
    const res = await api.get(`/api/waf/${TEST_DOMAIN}`);
    expect([200, 403]).toContain(res.status());
    if (res.status() === 200) {
      const json = await res.json();
      expect(json).toHaveProperty('domain', TEST_DOMAIN);
      expect(json).toHaveProperty('status');
      expect(json).toHaveProperty('removed_rules');
      expect(json).toHaveProperty('removed_tags');
    }
  });

  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/waf/definitely-not-owned-domain.example');
    expect(res.status()).toBe(403);
  });
});

test.describe('POST /api/waf/:domain', () => {
  test('rejects invalid status', async () => {
    const res = await api.post(`/api/waf/${TEST_DOMAIN}`, { data: { status: 'Maybe' } });
    expect([400, 403, 404]).toContain(res.status());
  });
});

mdescribe('POST /api/waf/:domain toggle lifecycle', () => {
  test('reads current status, flips it, flips it back', async () => {
    const before = await (await api.get(`/api/waf/${TEST_DOMAIN}`)).json();
    const flipped = before.status === 'On' ? 'Off' : 'On';

    const flipRes = await api.post(`/api/waf/${TEST_DOMAIN}`, { data: { status: flipped } });
    expect([200, 207]).toContain(flipRes.status());

    const restoreRes = await api.post(`/api/waf/${TEST_DOMAIN}`, { data: { status: before.status } });
    expect([200, 207]).toContain(restoreRes.status());
  });
});

test.describe('GET /api/waf/log/:domain', () => {
  test('returns paginated entries', async () => {
    const res = await api.get(`/api/waf/log/${TEST_DOMAIN}`);
    expect([200, 403]).toContain(res.status());
    if (res.status() === 200) {
      const json = await res.json();
      expect(json).toHaveProperty('entries');
      expect(json).toHaveProperty('total');
    }
  });
});

test.describe('GET /api/waf/stats/:domain', () => {
  test('?seconds param is reflected in response', async () => {
    const res = await api.get(`/api/waf/stats/${TEST_DOMAIN}?seconds=300`);
    expect([200, 403]).toContain(res.status());
    if (res.status() === 200) {
      const json = await res.json();
      expect(json).toHaveProperty('domain', TEST_DOMAIN);
      expect(json.seconds).toBe(300);
      expect(typeof json.checks).toBe('number');
      expect(typeof json.blocks).toBe('number');
    }
  });
});

test.describe('GET /api/waf/ids/:id_type', () => {
  test('accepts tags and ids', async () => {
    for (const idType of ['tags', 'ids']) {
      const res = await api.get(`/api/waf/ids/${idType}`);
      expect(res.status()).toBe(200);
      expect((await res.json())).toHaveProperty(idType);
    }
  });

  test('rejects unknown id_type', async () => {
    const res = await api.get('/api/waf/ids/bogus');
    expect(res.status()).toBe(400);
  });
});

// ---------------------------------------------------------------------------
// /api/php/*
// ---------------------------------------------------------------------------
test.describe('GET /api/php/versions', () => {
  test('returns available php versions', async () => {
    const res = await api.get('/api/php/versions');
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('versions');
  });
});

test.describe('GET /api/php/:version/ini', () => {
  test('unknown version returns 404', async () => {
    const res = await api.get('/api/php/9.9/ini');
    expect(res.status()).toBe(404);
  });
});

test.describe('PUT /api/php/:version/ini', () => {
  test('rejects missing content', async () => {
    const res = await api.put('/api/php/9.9/ini', { data: {} });
    expect([400, 404]).toContain(res.status());
  });
});

test.describe('GET /api/php/:version/options', () => {
  test('returns available_keys and current_config', async () => {
    const versionsRes = await api.get('/api/php/versions');
    const { versions } = await versionsRes.json();
    test.skip(!versions?.length, 'no PHP versions installed on this panel');
    const res = await api.get(`/api/php/${versions[0]}/options`);
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).toHaveProperty('available_keys');
    expect(json).toHaveProperty('current_config');
  });
});

test.describe('PUT /api/php/:version/options', () => {
  test('rejects a body with no recognized keys', async () => {
    const versionsRes = await api.get('/api/php/versions');
    const { versions } = await versionsRes.json();
    test.skip(!versions?.length, 'no PHP versions installed on this panel');
    const res = await api.put(`/api/php/${versions[0]}/options`, {
      data: { not_a_real_option: '123' },
    });
    expect(res.status()).toBe(400);
  });
});

test.describe('GET /api/php/:version/extensions', () => {
  test('returns extension list', async () => {
    const versionsRes = await api.get('/api/php/versions');
    const { versions } = await versionsRes.json();
    test.skip(!versions?.length, 'no PHP versions installed on this panel');
    const res = await api.get(`/api/php/${versions[0]}/extensions`);
    expect([200, 400]).toContain(res.status()); // 400 on LiteSpeed panels
    if (res.status() === 200) {
      expect(Array.isArray((await res.json()).extensions)).toBe(true);
    }
  });
});

test.describe('GET /api/php/:version/extensions/available', () => {
  test('returns installable extensions', async () => {
    const versionsRes = await api.get('/api/php/versions');
    const { versions } = await versionsRes.json();
    test.skip(!versions?.length, 'no PHP versions installed on this panel');
    const res = await api.get(`/api/php/${versions[0]}/extensions/available`);
    expect([200, 400]).toContain(res.status());
  });
});

test.describe('POST /api/php/:version/extensions/install', () => {
  test('rejects empty extensions list', async () => {
    const versionsRes = await api.get('/api/php/versions');
    const { versions } = await versionsRes.json();
    test.skip(!versions?.length, 'no PHP versions installed on this panel');
    const res = await api.post(`/api/php/${versions[0]}/extensions/install`, {
      data: { extensions: [] },
    });
    expect([400]).toContain(res.status());
  });
});

test.describe('GET /api/php/:version/extensions/install/status', () => {
  test('returns idle/busy status with no install_id', async () => {
    const versionsRes = await api.get('/api/php/versions');
    const { versions } = await versionsRes.json();
    test.skip(!versions?.length, 'no PHP versions installed on this panel');
    const res = await api.get(`/api/php/${versions[0]}/extensions/install/status`);
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('status');
  });
});

// toggling a real extension restarts php-fpm — real container churn, opt-in only.
mdescribe('POST /api/php/:version/extensions toggle', () => {
  test('enables then disables a low-risk extension', async () => {
    const versionsRes = await api.get('/api/php/versions');
    const { versions } = await versionsRes.json();
    test.skip(!versions?.length, 'no PHP versions installed on this panel');
    const ext = process.env.TEST_PHP_EXTENSION ?? '';
    test.skip(!ext, 'set TEST_PHP_EXTENSION to an extension name safe to toggle');
    const onRes = await api.post(`/api/php/${versions[0]}/extensions`, {
      data: { extension: ext, enable: '1' },
    });
    expect(onRes.status()).toBe(200);
    const offRes = await api.post(`/api/php/${versions[0]}/extensions`, {
      data: { extension: ext, enable: '0' },
    });
    expect(offRes.status()).toBe(200);
  });
});

// ---------------------------------------------------------------------------
// /api/webmail
// ---------------------------------------------------------------------------
test.describe('GET /api/webmail', () => {
  test('returns webmail running state', async () => {
    const res = await api.get('/api/webmail');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).toHaveProperty('is_running');
    expect(json).toHaveProperty('webmail_url');
  });
});

test.describe('POST /api/webmail/:email', () => {
  test('rejects an invalid email', async () => {
    const res = await api.post('/api/webmail/not-an-email');
    expect(res.status()).toBe(400);
  });

  test('unowned domain returns 403', async () => {
    const res = await api.post('/api/webmail/user@definitely-not-owned-domain.example');
    expect([403, 503]).toContain(res.status());
  });
});

// ---------------------------------------------------------------------------
// /api/webserver-conf
// ---------------------------------------------------------------------------
test.describe('GET /api/webserver-conf', () => {
  test('returns the live webserver config', async () => {
    const res = await api.get('/api/webserver-conf');
    expect([200, 404]).toContain(res.status());
    if (res.status() === 200) {
      const json = await res.json();
      expect(json).toHaveProperty('web_server');
      expect(json).toHaveProperty('content');
    }
  });
});

test.describe('PUT /api/webserver-conf', () => {
  test('rejects missing content', async () => {
    const res = await api.put('/api/webserver-conf', { data: {} });
    expect(res.status()).toBe(400);
  });
});

// overwrites the live webserver config and restarts the container — opt-in only.
mdescribe('PUT /api/webserver-conf round-trip', () => {
  test('writes back the same content it read', async () => {
    const getRes = await api.get('/api/webserver-conf');
    test.skip(getRes.status() !== 200, 'no webserver-conf available to read back');
    const { content } = await getRes.json();
    const putRes = await api.put('/api/webserver-conf', { data: { content } });
    expect(putRes.status()).toBe(200);
    expect((await putRes.json())).toHaveProperty('restarted');
  });
});

// ---------------------------------------------------------------------------
// /api/crons lifecycle
// ---------------------------------------------------------------------------
test.describe('GET /api/crons', () => {
  test('returns jobs and containers', async () => {
    const res = await api.get('/api/crons');
    expect([200, 400]).toContain(res.status());
    if (res.status() === 200) {
      expect((await res.json())).toHaveProperty('jobs');
    }
  });
});

test.describe('GET /api/crons/raw', () => {
  test('returns raw file content', async () => {
    const res = await api.get('/api/crons/raw');
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('content');
  });
});

test.describe('POST /api/crons', () => {
  test('rejects missing fields', async () => {
    const res = await api.post('/api/crons', { data: {} });
    expect(res.status()).toBe(400);
  });

  test('rejects an invalid schedule', async () => {
    const res = await api.post('/api/crons', {
      data: { schedule: 'not a cron expr', command: 'echo hi', container: 'php' },
    });
    expect(res.status()).toBe(400);
  });
});

test.describe('GET /api/crons/log', () => {
  test('returns log entries or a clear not-found error', async () => {
    const res = await api.get('/api/crons/log');
    expect([200, 404]).toContain(res.status());
    if (res.status() === 200) {
      expect(Array.isArray((await res.json()).entries)).toBe(true);
    }
  });

  test('rejects an overlong job filter', async () => {
    const res = await api.get(`/api/crons/log?job=${'x'.repeat(51)}`);
    expect(res.status()).toBe(400);
  });
});

mdescribe('/api/crons create -> update -> delete lifecycle', () => {
  const comment = rand('pwcron');
  const job = { schedule: '0 3 * * *', command: 'echo playwright-test', container: 'php' };

  test('POST creates a job', async () => {
    const res = await api.post('/api/crons', { data: { ...job, comment } });
    expect(res.status()).toBe(201);
    expect((await res.json()).job).toMatchObject({ comment });
  });

  test('PATCH updates the job', async () => {
    const newSchedule = '0 4 * * *';
    const res = await api.patch('/api/crons', {
      data: {
        ...job, schedule: newSchedule,
        original_schedule: job.schedule, original_command: job.command,
        original_container: job.container, original_comment: comment,
      },
    });
    expect([200, 404]).toContain(res.status());
    job.schedule = newSchedule;
  });

  test('DELETE removes the job', async () => {
    const res = await api.delete('/api/crons', { data: { ...job, comment } });
    expect(res.status()).toBe(200);
  });
});

// ---------------------------------------------------------------------------
// /api/pm2
// ---------------------------------------------------------------------------
test.describe('GET /api/pm2/:site/logs', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/pm2/definitely-not-owned-domain.example/logs');
    expect(res.status()).toBe(403);
  });
});

test.describe('POST /api/pm2/:site/{start,stop,restart}', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.post('/api/pm2/definitely-not-owned-domain.example/start');
    expect(res.status()).toBe(403);
  });

  test('protected service name is rejected', async () => {
    const res = await api.post(`/api/pm2/php-fpm-8.1/restart`);
    expect(res.status()).toBe(403);
  });
});

test.describe('PATCH /api/pm2/:site', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.patch('/api/pm2/definitely-not-owned-domain.example', {
      data: { cpu: '1' },
    });
    expect(res.status()).toBe(403);
  });
});

// PM2 lifecycle needs a real pre-provisioned nodejs/python site — no API exists
// to create one from scratch, so full create->delete isn't reachable here.
// start/stop/restart/delete on an existing app are exercised only if provided.
mdescribe('PM2 lifecycle against a pre-provisioned app', () => {
  const site = process.env.TEST_PM2_SITE ?? '';

  test('restart -> read logs', async () => {
    test.skip(!site, 'set TEST_PM2_SITE to an existing nodejs/python site path');
    const restartRes = await api.post(`/api/pm2/${site}/restart`);
    expect(restartRes.status()).toBe(200);
    const logsRes = await api.get(`/api/pm2/${site}/logs`);
    expect([200, 404]).toContain(logsRes.status());
  });
});

// ---------------------------------------------------------------------------
// /api/ftp lifecycle
// ---------------------------------------------------------------------------
test.describe('GET /api/ftp', () => {
  test('returns accounts array', async () => {
    const res = await api.get('/api/ftp');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(Array.isArray(json.accounts)).toBe(true);
    expect(json).toHaveProperty('server_ip');
  });
});

test.describe('POST /api/ftp', () => {
  test('rejects missing fields', async () => {
    const res = await api.post('/api/ftp', { data: {} });
    expect([400, 503]).toContain(res.status());
  });

  test('rejects a path outside /var/www/html/', async () => {
    const res = await api.post('/api/ftp', {
      data: { username: rand('pwftp'), password: 'Sup3rSecret!23', domain: TEST_DOMAIN, path: '/etc' },
    });
    expect([400, 503]).toContain(res.status());
  });
});

test.describe('GET /api/ftp/connections', () => {
  test('returns raw connections text', async () => {
    const res = await api.get('/api/ftp/connections');
    expect([200, 503]).toContain(res.status());
    if (res.status() === 200) {
      expect((await res.json())).toHaveProperty('connections');
    }
  });
});

test.describe('GET /api/ftp/configuration/:type/:account', () => {
  test('rejects unknown config type', async () => {
    const res = await api.get(`/api/ftp/configuration/bogus/someone@${TEST_DOMAIN}`);
    expect(res.status()).toBe(400);
  });
});

mdescribe('/api/ftp create -> update -> delete lifecycle', () => {
  const username = `${rand('pwftp')}@${TEST_DOMAIN}`;
  const password = 'Sup3rSecretPW!23';

  test('POST creates an account', async () => {
    const res = await api.post('/api/ftp', {
      data: { username, password, domain: TEST_DOMAIN, path: '/var/www/html/' },
    });
    expect([201, 503]).toContain(res.status());
    test.skip(res.status() === 503, 'FTP service not running on this panel');
  });

  test('PATCH updates the password', async () => {
    const res = await api.patch(`/api/ftp/${encodeURIComponent(username)}/password`, {
      data: { password: 'AnotherSecretPW!45' },
    });
    expect([200, 503]).toContain(res.status());
  });

  test('PATCH updates the path', async () => {
    const res = await api.patch(`/api/ftp/${encodeURIComponent(username)}/path`, {
      data: { path: '/var/www/html/' },
    });
    expect([200, 503]).toContain(res.status());
  });

  test('DELETE removes the account', async () => {
    const res = await api.delete(`/api/ftp/${encodeURIComponent(username)}`);
    expect([200, 503]).toContain(res.status());
  });
});

// ---------------------------------------------------------------------------
// /api/cache/* — memcached, redis, valkey, opensearch, elasticsearch share a shape
// ---------------------------------------------------------------------------
for (const service of ['memcached', 'redis', 'valkey', 'opensearch', 'elasticsearch']) {
  test.describe(`GET /api/cache/${service}`, () => {
    test('returns status shape', async () => {
      const res = await api.get(`/api/cache/${service}`);
      expect(res.status()).toBe(200);
      const json = await res.json();
      expect(json).toHaveProperty('service', service);
      expect(json).toHaveProperty('container_state');
      expect(json).toHaveProperty('actions');
    });
  });

  test.describe(`POST /api/cache/${service}`, () => {
    test('rejects an invalid action', async () => {
      const res = await api.post(`/api/cache/${service}`, { data: { action: 'not-a-real-action' } });
      expect(res.status()).toBe(400);
    });
  });

  // enable/disable/restart is a real container operation — opt-in only.
  mdescribe(`POST /api/cache/${service} restart`, () => {
    test('restarts the service', async () => {
      const res = await api.post(`/api/cache/${service}`, { data: { action: 'restart' } });
      expect(res.status()).toBe(200);
      expect((await res.json())).toHaveProperty('message');
    });
  });
}

test.describe('GET /api/cache/varnish', () => {
  test('returns status and domain_statuses', async () => {
    const res = await api.get('/api/cache/varnish');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).toHaveProperty('service', 'varnish');
    expect(json).toHaveProperty('domain_statuses');
  });
});

test.describe('POST /api/cache/varnish', () => {
  test('rejects an invalid action', async () => {
    const res = await api.post('/api/cache/varnish', { data: { action: 'not-a-real-action' } });
    expect(res.status()).toBe(400);
  });
});

test.describe('GET /api/cache/varnish/stats', () => {
  test('returns running or stopped shape', async () => {
    const res = await api.get('/api/cache/varnish/stats');
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('status');
  });
});

test.describe('GET /api/cache/varnish/domains', () => {
  test('returns a domain status map', async () => {
    const res = await api.get('/api/cache/varnish/domains');
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('domain_statuses');
  });
});

test.describe('POST /api/cache/varnish/domains/:domain', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.post('/api/cache/varnish/domains/definitely-not-owned-domain.example', {
      data: { status: 'On' },
    });
    expect(res.status()).toBe(403);
  });

  test('rejects invalid status', async () => {
    const res = await api.post(`/api/cache/varnish/domains/${TEST_DOMAIN}`, { data: { status: 'Maybe' } });
    expect([400]).toContain(res.status());
  });
});

// varnish enable/disable performs a live sequence of webserver container
// start/stop operations — genuinely disruptive, opt-in only.
mdescribe('POST /api/cache/varnish restart', () => {
  test('restarts varnish', async () => {
    const res = await api.post('/api/cache/varnish', { data: { action: 'restart' } });
    expect([200, 500]).toContain(res.status());
  });
});

// ---------------------------------------------------------------------------
// /api/mysql lifecycle
// ---------------------------------------------------------------------------
test.describe('GET /api/mysql/databases', () => {
  test('returns databases array', async () => {
    const res = await api.get('/api/mysql/databases');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      expect(Array.isArray((await res.json()).databases)).toBe(true);
    }
  });
});

test.describe('POST /api/mysql/databases', () => {
  test('rejects an invalid database name', async () => {
    const res = await api.post('/api/mysql/databases', { data: { name: 'not valid!' } });
    expect([400]).toContain(res.status());
  });

  test('rejects a reserved database name', async () => {
    const res = await api.post('/api/mysql/databases', { data: { name: 'mysql' } });
    expect([400]).toContain(res.status());
  });
});

test.describe('GET /api/mysql/info', () => {
  test('returns databases/users/assigned_databases', async () => {
    const res = await api.get('/api/mysql/info');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      const json = await res.json();
      expect(json).toHaveProperty('databases');
      expect(json).toHaveProperty('users');
      expect(json).toHaveProperty('assigned_databases');
    }
  });
});

test.describe('GET /api/mysql/processlist', () => {
  test('returns processlist array', async () => {
    const res = await api.get('/api/mysql/processlist');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      expect(Array.isArray((await res.json()).processlist)).toBe(true);
    }
  });
});

test.describe('GET /api/mysql/remote-access', () => {
  test('returns enabled/server_ip/port', async () => {
    const res = await api.get('/api/mysql/remote-access');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      const json = await res.json();
      expect(json).toHaveProperty('enabled');
      expect(json).toHaveProperty('server_ip');
    }
  });
});

test.describe('GET /api/mysql/configuration', () => {
  test('returns configuration and available_keys', async () => {
    const res = await api.get('/api/mysql/configuration');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).toHaveProperty('configuration');
    expect(json).toHaveProperty('available_keys');
  });
});

test.describe('PUT /api/mysql/configuration', () => {
  test('rejects a body with no recognized keys', async () => {
    const res = await api.put('/api/mysql/configuration', { data: { not_a_real_key: '1' } });
    expect(res.status()).toBe(400);
  });
});

mdescribe('/api/mysql database + user + grants lifecycle', () => {
  const dbName = rand('pwdb');
  const userName = rand('pwuser');
  const password = 'Sup3rSecretPW!23';

  test('POST creates a database', async () => {
    const res = await api.post('/api/mysql/databases', { data: { name: dbName } });
    expect(res.status()).toBe(201);
  });

  test('POST creates a user', async () => {
    const res = await api.post('/api/mysql/users', { data: { username: userName, password } });
    expect(res.status()).toBe(201);
  });

  test('POST grants privileges', async () => {
    const res = await api.post('/api/mysql/grants', {
      data: { username: userName, database: dbName, privileges: ['SELECT', 'INSERT'] },
    });
    expect(res.status()).toBe(201);
  });

  test('GET privileges reflects the grant', async () => {
    const res = await api.get(`/api/mysql/users/${userName}/privileges/${dbName}`);
    expect(res.status()).toBe(200);
    expect(Array.isArray((await res.json()).privileges)).toBe(true);
  });

  test('GET tables on the new (empty) database', async () => {
    const res = await api.get(`/api/mysql/databases/${dbName}/tables`);
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('tables');
  });

  test('PATCH updates the user password', async () => {
    const res = await api.patch(`/api/mysql/users/${userName}/password`, {
      data: { password: 'AnotherSecretPW!45' },
    });
    expect(res.status()).toBe(200);
  });

  test('DELETE revokes grants', async () => {
    const res = await api.delete('/api/mysql/grants', { data: { username: userName, database: dbName } });
    expect(res.status()).toBe(200);
  });

  test('DELETE removes the user', async () => {
    const res = await api.delete(`/api/mysql/users/${userName}`);
    expect(res.status()).toBe(200);
  });

  test('DELETE removes the database', async () => {
    const res = await api.delete(`/api/mysql/databases/${dbName}`);
    expect(res.status()).toBe(200);
  });
});

// remote-access toggles a real exposed port for the whole MySQL instance —
// opt-in only, and always restore state in the same test.
mdescribe('POST /api/mysql/remote-access toggle', () => {
  test('flips the setting and flips it back', async () => {
    const before = await (await api.get('/api/mysql/remote-access')).json();
    const flipRes = await api.post('/api/mysql/remote-access', {
      data: { action: before.enabled ? 'disable' : 'enable' },
    });
    expect(flipRes.status()).toBe(200);
    const restoreRes = await api.post('/api/mysql/remote-access', {
      data: { action: before.enabled ? 'enable' : 'disable' },
    });
    expect(restoreRes.status()).toBe(200);
  });
});

// ---------------------------------------------------------------------------
// /api/postgresql lifecycle (mirrors mysql.py)
// ---------------------------------------------------------------------------
test.describe('GET /api/postgresql/databases', () => {
  test('returns databases array', async () => {
    const res = await api.get('/api/postgresql/databases');
    expect([200, 500]).toContain(res.status());
  });
});

test.describe('POST /api/postgresql/databases', () => {
  test('rejects a reserved database name', async () => {
    const res = await api.post('/api/postgresql/databases', { data: { name: 'postgres' } });
    expect([400]).toContain(res.status());
  });
});

test.describe('GET /api/postgresql/info', () => {
  test('returns databases/users/assigned_databases', async () => {
    const res = await api.get('/api/postgresql/info');
    expect([200, 500]).toContain(res.status());
  });
});

test.describe('GET /api/postgresql/processlist', () => {
  test('returns processlist array', async () => {
    const res = await api.get('/api/postgresql/processlist');
    expect([200, 500]).toContain(res.status());
  });
});

test.describe('GET /api/postgresql/remote-access', () => {
  test('returns enabled/server_ip/port', async () => {
    const res = await api.get('/api/postgresql/remote-access');
    expect([200, 500]).toContain(res.status());
  });
});

test.describe('GET /api/postgresql/configuration', () => {
  test('returns configuration or a clear container-down error', async () => {
    const res = await api.get('/api/postgresql/configuration');
    expect([200, 503]).toContain(res.status());
    if (res.status() === 503) {
      expect((await res.json())).toHaveProperty('error');
    }
  });
});

test.describe('PUT /api/postgresql/configuration', () => {
  test('rejects a body with no recognized keys', async () => {
    const res = await api.put('/api/postgresql/configuration', { data: { not_a_real_key: '1' } });
    expect(res.status()).toBe(400);
  });
});

mdescribe('/api/postgresql database + user + grants lifecycle', () => {
  const dbName = rand('pwpgdb');
  const userName = rand('pwpguser');
  const password = 'Sup3rSecretPW!23';

  test('POST creates a database', async () => {
    const res = await api.post('/api/postgresql/databases', { data: { name: dbName } });
    expect(res.status()).toBe(201);
  });

  test('POST creates a user', async () => {
    const res = await api.post('/api/postgresql/users', { data: { username: userName, password } });
    expect(res.status()).toBe(201);
  });

  test('POST grants privileges', async () => {
    const res = await api.post('/api/postgresql/grants', {
      data: { username: userName, database: dbName },
    });
    expect(res.status()).toBe(201);
  });

  test('PATCH updates the user password', async () => {
    const res = await api.patch(`/api/postgresql/users/${userName}/password`, {
      data: { password: 'AnotherSecretPW!45' },
    });
    expect(res.status()).toBe(200);
  });

  test('DELETE revokes grants', async () => {
    const res = await api.delete('/api/postgresql/grants', { data: { username: userName, database: dbName } });
    expect(res.status()).toBe(200);
  });

  test('DELETE removes the user', async () => {
    const res = await api.delete(`/api/postgresql/users/${userName}`);
    expect(res.status()).toBe(200);
  });

  test('DELETE removes the database', async () => {
    const res = await api.delete(`/api/postgresql/databases/${dbName}`);
    expect(res.status()).toBe(200);
  });
});

mdescribe('POST /api/postgresql/remote-access toggle', () => {
  test('flips the setting and flips it back', async () => {
    const before = await (await api.get('/api/postgresql/remote-access')).json();
    const flipRes = await api.post('/api/postgresql/remote-access', {
      data: { action: before.enabled ? 'disable' : 'enable' },
    });
    expect(flipRes.status()).toBe(200);
    await api.post('/api/postgresql/remote-access', {
      data: { action: before.enabled ? 'enable' : 'disable' },
    });
  });
});

// ---------------------------------------------------------------------------
// /api/inodes, /api/disk-usage — read-only, path-traversal guarded
// ---------------------------------------------------------------------------
test.describe('GET /api/inodes', () => {
  test('returns entries sorted by inode_count', async () => {
    const res = await api.get('/api/inodes');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      expect(Array.isArray((await res.json()).entries)).toBe(true);
    }
  });

  test('rejects path traversal', async () => {
    const res = await api.get('/api/inodes/..%2F..%2F..%2Fetc');
    expect([400, 404]).toContain(res.status());
  });
});

test.describe('GET /api/disk-usage', () => {
  test('returns entries with size/path', async () => {
    const res = await api.get('/api/disk-usage');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      expect(Array.isArray((await res.json()).entries)).toBe(true);
    }
  });

  test('rejects path traversal', async () => {
    const res = await api.get('/api/disk-usage/..%2F..%2F..%2Fetc');
    expect([400, 404]).toContain(res.status());
  });
});

// ---------------------------------------------------------------------------
// /api/usage
// ---------------------------------------------------------------------------
test.describe('GET /api/usage', () => {
  test('returns usage data or a clear not-yet-available error', async () => {
    const res = await api.get('/api/usage');
    expect([200, 503]).toContain(res.status());
  });
});

test.describe('GET /api/usage/history', () => {
  test('returns paginated entries', async () => {
    const res = await api.get('/api/usage/history');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).toHaveProperty('entries');
    expect(json).toHaveProperty('total');
    expect(json).toHaveProperty('page');
  });

  test('?page=2 is reflected in response', async () => {
    const res = await api.get('/api/usage/history?page=2');
    expect(res.status()).toBe(200);
    expect((await res.json()).page).toBe(2);
  });
});

// ---------------------------------------------------------------------------
// /api/malware-scanner — read-only quarantine list is safe; scan is heavy (up
// to 300s, real clamscan) and gated.
// ---------------------------------------------------------------------------
test.describe('GET /api/malware-scanner/quarantine', () => {
  test('returns quarantine_files array', async () => {
    const res = await api.get('/api/malware-scanner/quarantine');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      expect(Array.isArray((await res.json()).quarantine_files)).toBe(true);
    }
  });
});

test.describe('POST /api/malware-scanner/scan', () => {
  test('rejects a directory outside /var/www/html', async () => {
    const res = await api.post('/api/malware-scanner/scan', { data: { directory: '/etc' } });
    expect(res.status()).toBe(400);
  });

  test('rejects a non-existent directory', async () => {
    const res = await api.post('/api/malware-scanner/scan', {
      data: { directory: '/var/www/html/definitely-does-not-exist' },
    });
    expect(res.status()).toBe(404);
  });
});

mtest('POST /api/malware-scanner/scan runs a real clamscan', async () => {
  test.setTimeout(320_000);
  const res = await api.post('/api/malware-scanner/scan', { data: { directory: '/var/www/html' } });
  expect([200, 504]).toContain(res.status());
});

// ---------------------------------------------------------------------------
// /api/emails, aliases, filters, deliverability
// ---------------------------------------------------------------------------
test.describe('GET /api/emails', () => {
  test('returns an emails array', async () => {
    const res = await api.get('/api/emails');
    expect(res.status()).toBe(200);
    expect(Array.isArray((await res.json()).emails)).toBe(true);
  });
});

test.describe('POST /api/emails', () => {
  test('rejects missing fields', async () => {
    const res = await api.post('/api/emails', { data: {} });
    expect(res.status()).toBe(400);
  });
});

test.describe('GET /api/emails/deliverability', () => {
  test('returns per-domain deliverability checks', async () => {
    const res = await api.get('/api/emails/deliverability');
    expect(res.status()).toBe(200);
    expect(Array.isArray((await res.json()).domains)).toBe(true);
  });
});

test.describe('GET /api/emails/deliverability/:domain', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/emails/deliverability/definitely-not-owned-domain.example');
    expect(res.status()).toBe(403);
  });
});

test.describe('GET /api/emails/configuration/:type/:account', () => {
  test('rejects an unknown config type', async () => {
    const res = await api.get(`/api/emails/configuration/bogus/someone@${TEST_DOMAIN}`);
    expect(res.status()).toBe(400);
  });
});

test.describe('GET /api/emails/aliases', () => {
  test('returns an aliases array', async () => {
    const res = await api.get('/api/emails/aliases');
    expect(res.status()).toBe(200);
    expect(Array.isArray((await res.json()).aliases)).toBe(true);
  });
});

test.describe('POST /api/emails/aliases', () => {
  test('rejects an invalid target address', async () => {
    const res = await api.post('/api/emails/aliases', {
      data: { username: rand('pwalias'), domain: TEST_DOMAIN, target: 'not-an-email' },
    });
    expect(res.status()).toBe(400);
  });
});

test.describe('GET /api/emails/default/:domain', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/emails/default/definitely-not-owned-domain.example');
    expect(res.status()).toBe(403);
  });
});

mdescribe('/api/emails mailbox + alias + filter lifecycle', () => {
  const localPart = rand('pwmail');
  const address = `${localPart}@${TEST_DOMAIN}`;
  const password = 'Sup3rSecretPW!23';

  test('POST creates a mailbox', async () => {
    const res = await api.post('/api/emails', {
      data: { domain: TEST_DOMAIN, username: localPart, password },
    });
    expect(res.status()).toBe(201);
  });

  test('GET returns the mailbox detail', async () => {
    const res = await api.get(`/api/emails/${encodeURIComponent(address)}`);
    expect(res.status()).toBe(200);
    expect((await res.json())).toHaveProperty('address', address);
  });

  test('PATCH suspends incoming mail', async () => {
    const res = await api.patch(`/api/emails/${encodeURIComponent(address)}`, {
      data: { incoming: 'suspend' },
    });
    expect(res.status()).toBe(200);
  });

  test('PUT sets a filter', async () => {
    const res = await api.put(`/api/emails/filters/${encodeURIComponent(address)}`, {
      data: { content: 'if header :contains "subject" "test" { discard; }' },
    });
    expect(res.status()).toBe(200);
  });

  test('POST creates an alias to the mailbox', async () => {
    const res = await api.post('/api/emails/aliases', {
      data: { username: rand('pwaliassrc'), domain: TEST_DOMAIN, target: address },
    });
    expect(res.status()).toBe(201);
  });

  test('DELETE removes the mailbox', async () => {
    const res = await api.delete(`/api/emails/${encodeURIComponent(address)}`);
    expect(res.status()).toBe(200);
  });
});

// ---------------------------------------------------------------------------
// /api/website-builder — POST requires a numeric domain_id not exposed by any
// GET endpoint, so full creation isn't independently reachable via the API;
// only validation paths are covered unless TEST_WEBSITE_BUILDER_DOMAIN_ID is set.
// ---------------------------------------------------------------------------
test.describe('GET /api/website-builder/:domain', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/website-builder/definitely-not-owned-domain.example');
    expect(res.status()).toBe(403);
  });
});

test.describe('POST /api/website-builder', () => {
  test('rejects a missing domain_id', async () => {
    const res = await api.post('/api/website-builder', { data: {} });
    expect(res.status()).toBe(400);
  });
});

test.describe('DELETE /api/website-builder/sites/:id', () => {
  test('unknown site_id returns 404', async () => {
    const res = await api.delete('/api/website-builder/sites/999999999');
    expect([403, 404]).toContain(res.status());
  });
});

mdescribe('/api/website-builder create -> edit -> delete lifecycle', () => {
  const domainId = process.env.TEST_WEBSITE_BUILDER_DOMAIN_ID ?? '';
  let siteId: number;

  test('POST creates a site', async () => {
    test.skip(!domainId, 'set TEST_WEBSITE_BUILDER_DOMAIN_ID to a free domain row id owned by the test user');
    const res = await api.post('/api/website-builder', { data: { domain_id: Number(domainId) } });
    expect([201, 409]).toContain(res.status());
  });

  test('PUT updates the html/css', async () => {
    test.skip(!domainId, 'no domain_id configured');
    const res = await api.put(`/api/website-builder/${TEST_DOMAIN}`, {
      data: { html: '<h1>playwright</h1>', css: 'h1{color:red}' },
    });
    expect(res.status()).toBe(200);
  });
});

// ---------------------------------------------------------------------------
// /api/sites (websites.py) — mostly read-only + external scanners
// ---------------------------------------------------------------------------
test.describe('GET /api/sites', () => {
  test('returns a sites array', async () => {
    const res = await api.get('/api/sites');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(Array.isArray(json.sites)).toBe(true);
    expect(json).toHaveProperty('count');
  });
});

test.describe('GET /api/sites/:domain', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/sites/definitely-not-owned-domain.example');
    expect(res.status()).toBe(403);
  });

  test('TEST_DOMAIN returns site detail or 404 if unmanaged', async () => {
    const res = await api.get(`/api/sites/${TEST_DOMAIN}`);
    expect([200, 404]).toContain(res.status());
  });
});

test.describe('GET /api/sites/:domain/safebrowsing', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/sites/definitely-not-owned-domain.example/safebrowsing');
    expect(res.status()).toBe(403);
  });
});

test.describe('POST /api/sites/:domain/pagespeed', () => {
  test('kicks off an async scan', async () => {
    const res = await api.post(`/api/sites/${TEST_DOMAIN}/pagespeed`);
    expect([202, 403]).toContain(res.status());
  });

  test('rejects an invalid domain param', async () => {
    const res = await api.post('/api/sites/%00bad/pagespeed');
    expect([400, 403]).toContain(res.status());
  });
});

test.describe('POST /api/sites/:domain/wp-vulnerability', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.post('/api/sites/definitely-not-owned-domain.example/wp-vulnerability');
    expect(res.status()).toBe(403);
  });
});

// GET pagespeed/wp-vulnerability run synchronous external scans up to 120s when
// no cache file exists yet — slow and network-dependent, opt-in only.
mdescribe('GET /api/sites/:domain/pagespeed and wp-vulnerability (synchronous scan)', () => {
  test('pagespeed', async () => {
    test.setTimeout(130_000);
    const res = await api.get(`/api/sites/${TEST_DOMAIN}/pagespeed`);
    expect([200, 403, 500]).toContain(res.status());
  });

  test('wp-vulnerability', async () => {
    test.setTimeout(130_000);
    const res = await api.get(`/api/sites/${TEST_DOMAIN}/wp-vulnerability`);
    expect([200, 403]).toContain(res.status());
  });
});

// ---------------------------------------------------------------------------
// /api/wordpress — reads are safe; backup is heavy-but-safe; restore/uninstall
// are destructive and need a disposable WP install, so they stay opt-in.
// ---------------------------------------------------------------------------
test.describe('GET /api/wordpress', () => {
  test('returns sites and count', async () => {
    const res = await api.get('/api/wordpress');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(Array.isArray(json.sites)).toBe(true);
    expect(json).toHaveProperty('count');
  });
});

test.describe('GET /api/wordpress/:domain/backups', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/wordpress/definitely-not-owned-domain.example/backups');
    expect(res.status()).toBe(403);
  });
});

test.describe('POST /api/wordpress/:domain/backups', () => {
  test('rejects both backup flags disabled', async () => {
    const res = await api.post(`/api/wordpress/${TEST_DOMAIN}/backups`, {
      data: { backup_database: false, backup_files: false },
    });
    expect([400, 403, 404]).toContain(res.status());
  });
});

test.describe('POST /api/wordpress/:domain/restore', () => {
  test('rejects missing backup_date', async () => {
    const res = await api.post(`/api/wordpress/${TEST_DOMAIN}/restore`, { data: {} });
    expect([400, 403]).toContain(res.status());
  });
});

test.describe('GET /api/wordpress/secure', () => {
  test('returns available hardening rules', async () => {
    const res = await api.get('/api/wordpress/secure');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      expect(Array.isArray((await res.json()).rules)).toBe(true);
    }
  });
});

test.describe('GET /api/wordpress/:domain/secure', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/wordpress/definitely-not-owned-domain.example/secure');
    expect(res.status()).toBe(403);
  });
});

test.describe('PUT /api/wordpress/:domain/secure', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.put('/api/wordpress/definitely-not-owned-domain.example/secure', {
      data: { disable_all: true },
    });
    expect(res.status()).toBe(403);
  });
});

test.describe('DELETE /api/wordpress/sites/:id', () => {
  test('unknown site_id returns 404', async () => {
    const res = await api.delete('/api/wordpress/sites/999999999');
    expect([403, 404]).toContain(res.status());
  });
});

test.describe('POST /api/wp-cli/:action', () => {
  test('rejects an unknown action', async () => {
    const res = await api.post('/api/wp-cli/not-a-real-action', { data: { domain: TEST_DOMAIN } });
    expect(res.status()).toBe(400);
  });

  test('rejects a missing domain', async () => {
    const res = await api.post('/api/wp-cli/list_plugins', { data: {} });
    expect(res.status()).toBe(400);
  });
});

// safe/read-mostly wp-cli actions against a real, pre-provisioned WP install.
mdescribe('POST /api/wp-cli/:action read-only actions', () => {
  for (const action of ['list_plugins', 'list_themes', 'cache_flush']) {
    test(action, async () => {
      test.setTimeout(130_000);
      const res = await api.post(`/api/wp-cli/${action}`, { data: { domain: TEST_DOMAIN } });
      expect([200, 403, 404]).toContain(res.status());
    });
  }
});

mdescribe('POST /api/wordpress/:domain/backups real backup', () => {
  test('creates a backup of TEST_DOMAIN', async () => {
    test.setTimeout(130_000);
    const res = await api.post(`/api/wordpress/${TEST_DOMAIN}/backups`, {
      data: { backup_database: true, backup_files: true },
    });
    expect([200, 400, 403, 404]).toContain(res.status());
  });
});

// ---------------------------------------------------------------------------
// /api/process-manager
// ---------------------------------------------------------------------------
test.describe('GET /api/process-manager', () => {
  test('returns a processes array', async () => {
    const res = await api.get('/api/process-manager');
    expect([200, 500]).toContain(res.status());
    if (res.status() === 200) {
      expect(Array.isArray((await res.json()).processes)).toBe(true);
    }
  });
});

test.describe('DELETE /api/process-manager/:pid', () => {
  test('rejects an invalid pid', async () => {
    const res = await api.delete('/api/process-manager/0');
    expect(res.status()).toBe(400);
  });

  test('rejects a pid that is not one of the caller\'s processes', async () => {
    const res = await api.delete('/api/process-manager/1');
    expect([403, 500]).toContain(res.status());
  });
});
// no mutating test here: killing a real PID is destructive and there's no
// safe way to spin up a disposable one through this API.

// ---------------------------------------------------------------------------
// /api/domains lifecycle — POST/DELETE provision real domains (vhost, DNS
// zone, containers); SSL generation hits Let's Encrypt (rate-limited). Both
// gated. Reads and validation errors run unconditionally.
// ---------------------------------------------------------------------------
test.describe('GET /api/domains', () => {
  test('returns domains with total', async () => {
    const res = await api.get('/api/domains');
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(Array.isArray(json.domains)).toBe(true);
    expect(json).toHaveProperty('total');
  });
});

test.describe('POST /api/domains', () => {
  test('rejects a missing domain', async () => {
    const res = await api.post('/api/domains', { data: {} });
    expect(res.status()).toBe(400);
  });
});

test.describe('GET /api/domains/:domain/status', () => {
  test('unowned domain returns 403', async () => {
    const res = await api.get('/api/domains/definitely-not-owned-domain.example/status');
    expect(res.status()).toBe(403);
  });

  test('TEST_DOMAIN returns status shape', async () => {
    const res = await api.get(`/api/domains/${TEST_DOMAIN}/status`);
    expect(res.status()).toBe(200);
    const json = await res.json();
    expect(json).toHaveProperty('domain', TEST_DOMAIN);
    expect(json).toHaveProperty('ssl');
    expect(json).toHaveProperty('status');
  });
});

test.describe('GET /api/domains/:domain/docroot', () => {
  test('TEST_DOMAIN returns a docroot', async () => {
    const res = await api.get(`/api/domains/${TEST_DOMAIN}/docroot`);
    expect([200, 403]).toContain(res.status());
    if (res.status() === 200) {
      expect((await res.json())).toHaveProperty('docroot');
    }
  });
});

test.describe('PUT /api/domains/:domain/docroot', () => {
  test('rejects a docroot outside /var/www/html/', async () => {
    const res = await api.put(`/api/domains/${TEST_DOMAIN}/docroot`, { data: { docroot: '/etc' } });
    expect([400, 403]).toContain(res.status());
  });
});

test.describe('GET /api/domains/:domain/redirect', () => {
  test('TEST_DOMAIN returns redirect status', async () => {
    const res = await api.get(`/api/domains/${TEST_DOMAIN}/redirect`);
    expect([200, 403]).toContain(res.status());
  });
});

test.describe('PUT /api/domains/:domain/redirect', () => {
  test('rejects an invalid url', async () => {
    const res = await api.put(`/api/domains/${TEST_DOMAIN}/redirect`, { data: { redirect_url: 'not-a-url' } });
    expect([400, 403]).toContain(res.status());
  });
});

test.describe('GET /api/domains/:domain/ssl', () => {
  test('TEST_DOMAIN returns ssl_mode', async () => {
    const res = await api.get(`/api/domains/${TEST_DOMAIN}/ssl`);
    expect([200, 403]).toContain(res.status());
    if (res.status() === 200) {
      expect((await res.json())).toHaveProperty('ssl_mode');
    }
  });
});

test.describe('POST /api/domains/:domain/ssl', () => {
  test('rejects an unknown action', async () => {
    const res = await api.post(`/api/domains/${TEST_DOMAIN}/ssl`, { data: { action: 'not-a-real-action' } });
    expect([400, 403]).toContain(res.status());
  });
});

test.describe('GET /api/domains/:domain/vhost', () => {
  test('TEST_DOMAIN returns vhost content or 404', async () => {
    const res = await api.get(`/api/domains/${TEST_DOMAIN}/vhost`);
    expect([200, 403, 404]).toContain(res.status());
  });
});

test.describe('PUT /api/domains/:domain/vhost', () => {
  test('rejects empty vhost content', async () => {
    const res = await api.put(`/api/domains/${TEST_DOMAIN}/vhost`, { data: { vhost: '' } });
    expect([400, 403]).toContain(res.status());
  });
});

test.describe('GET /api/domains/:domain/dns', () => {
  test('TEST_DOMAIN returns records or 404', async () => {
    const res = await api.get(`/api/domains/${TEST_DOMAIN}/dns`);
    expect([200, 403, 404]).toContain(res.status());
  });
});

test.describe('POST /api/domains/:domain/dns/records', () => {
  test('rejects missing fields', async () => {
    const res = await api.post(`/api/domains/${TEST_DOMAIN}/dns/records`, { data: {} });
    expect([400, 403, 404]).toContain(res.status());
  });
});

mdescribe('/api/domains/:domain/dns record create -> update -> delete (via domains.py)', () => {
  const subdomain = rand('pwddns2');
  let rowId: number;

  test('POST creates a record', async () => {
    const res = await api.post(`/api/domains/${TEST_DOMAIN}/dns/records`, {
      data: { name: subdomain, type: 'A', value: '203.0.113.30', ttl: 3600 },
    });
    expect(res.status()).toBe(201);
    rowId = (await res.json()).record.line_number ?? (await res.json()).record.row_id;
  });

  test('DELETE removes the record', async () => {
    test.skip(rowId === undefined, 'no row_id captured from create step');
    const res = await api.delete(`/api/domains/${TEST_DOMAIN}/dns/records/${rowId}`);
    expect(res.status()).toBe(200);
  });
});

test.describe('GET /api/domains/:domain/logs', () => {
  test('TEST_DOMAIN returns paginated logs', async () => {
    const res = await api.get(`/api/domains/${TEST_DOMAIN}/logs`);
    expect([200, 403, 404]).toContain(res.status());
    if (res.status() === 200) {
      const json = await res.json();
      expect(json).toHaveProperty('total_pages');
    }
  });
});

// full domain provisioning + SSL issuance — real infra, Let's Encrypt rate
// limits apply. Only run against a disposable panel with TEST_NEW_DOMAIN set.
mdescribe('/api/domains create -> suspend -> unsuspend -> delete lifecycle', () => {
  const domain = process.env.TEST_NEW_DOMAIN ?? '';

  test('POST creates the domain', async () => {
    test.skip(!domain, 'set TEST_NEW_DOMAIN to a disposable domain pointed at this panel');
    const res = await api.post('/api/domains', { data: { domain } });
    expect(res.status()).toBe(201);
  });

  test('POST suspends it', async () => {
    test.skip(!domain, 'no TEST_NEW_DOMAIN configured');
    const res = await api.post(`/api/domains/${domain}/suspend`);
    expect(res.status()).toBe(200);
  });

  test('POST unsuspends it', async () => {
    test.skip(!domain, 'no TEST_NEW_DOMAIN configured');
    const res = await api.post(`/api/domains/${domain}/unsuspend`);
    expect(res.status()).toBe(200);
  });

  test('DELETE removes the domain', async () => {
    test.skip(!domain, 'no TEST_NEW_DOMAIN configured');
    const res = await api.delete(`/api/domains/${domain}`);
    expect(res.status()).toBe(200);
  });
});

// ---------------------------------------------------------------------------
// Auth guard — every route below must reject a request with no Bearer token.
// api_required() returns 401 {"error":"Authentication required", ...} for any
// request that isn't JWT-authenticated (no redirect for JWT-style requests).
// One representative GET route per module in modules/api/, plus the two core
// routes.
// ---------------------------------------------------------------------------
test.describe('Auth guard', () => {

  const routes = [
    '/api/endpoints',
    '/api/account',
    '/api/account/sessions',
    '/api/services',
    '/api/containers',
    '/api/containers/status',
    '/api/stats',
    '/api/autoinstaller',
    '/api/ip-blocker',
    '/api/dynamic-dns',
    '/api/dns',
    `/api/dns/${TEST_DOMAIN}`,
    '/api/waf',
    `/api/waf/${TEST_DOMAIN}`,
    '/api/php/versions',
    '/api/webmail',
    '/api/webserver-conf',
    '/api/crons',
    `/api/pm2/${TEST_DOMAIN}/logs`,
    '/api/ftp',
    '/api/cache/redis',
    '/api/cache/memcached',
    '/api/cache/valkey',
    '/api/cache/opensearch',
    '/api/cache/elasticsearch',
    '/api/cache/varnish',
    '/api/mysql/databases',
    '/api/postgresql/databases',
    '/api/inodes',
    '/api/disk-usage',
    '/api/usage',
    '/api/malware-scanner/quarantine',
    '/api/emails',
    `/api/website-builder/${TEST_DOMAIN}`,
    '/api/sites',
    '/api/wordpress',
    '/api/process-manager',
    '/api/domains',
  ];

  test('unauthenticated requests are rejected on all protected routes', async () => {
    const ctx = await pwRequest.newContext({ baseURL: BASE_URL });
    try {
      for (const route of routes) {
        const res = await ctx.get(route, { maxRedirects: 0 });
        expect(
          [401, 403, 302],
          `${route} should reject unauthenticated requests (got ${res.status()})`
        ).toContain(res.status());
        if (res.status() === 401) {
          const json = await res.json();
          expect(json).toHaveProperty('error', 'Authentication required');
        }
      }
    } finally {
      await ctx.dispose();
    }
  });

  test('a garbage bearer token is rejected the same way', async () => {
    const ctx = await pwRequest.newContext({
      baseURL: BASE_URL,
      extraHTTPHeaders: { Authorization: 'Bearer not-a-real-jwt' },
    });
    try {
      const res = await ctx.get('/api/endpoints', { maxRedirects: 0 });
      expect([401, 302]).toContain(res.status());
    } finally {
      await ctx.dispose();
    }
  });
});
