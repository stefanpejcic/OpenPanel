// shared prepare() helpers for shots.mjs

// swap demo text for readable names: text nodes, <option>s and input values under scope
export const rename = (map, scope = 'main') => async page => {
  await page.evaluate(([map, scope]) => {
    const swap = s => {
      for (const [from, to] of Object.entries(map)) s = s.split(from).join(to);
      return s;
    };
    for (const root of document.querySelectorAll(scope)) {
      const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
      for (let n; (n = walker.nextNode()); ) {
        const next = swap(n.textContent);
        if (next !== n.textContent) n.textContent = next;
      }
      root.querySelectorAll('input:not([type=hidden]), textarea').forEach(i => { if (i.value) i.value = swap(i.value); });
    }
  }, [map, scope]);
};

// run several prepare() steps in order
export const all = (...steps) => async page => {
  for (const s of steps) await s(page);
};

export const click = (selector, wait = 400) => async page => {
  await page.locator(selector).first().click();
  await page.waitForTimeout(wait);
};

export const fill = values => async page => {
  for (const [sel, v] of Object.entries(values)) await page.locator(sel).first().fill(v);
};

// fill the visible text/password inputs in main, in page order
export const fillVisible = values => async page => {
  const inputs = page.locator('main input:is([type=text], [type=password], [type=email], [type=url], :not([type])):visible:enabled');
  const n = Math.min(values.length, await inputs.count());
  for (let i = 0; i < n; i++) if (values[i] != null) await inputs.nth(i).fill(values[i]);
  await page.waitForTimeout(300);
};

// wait for main's selects to be populated, then pick the first real option in each
export const selectFirst = (wait = 800) => async page => {
  await page.waitForFunction(() => [...document.querySelectorAll('main select')].every(s => s.options.length > 1), null, { timeout: 8000 }).catch(() => {});
  const selects = page.locator('main select:visible');
  for (let i = 0; i < (await selects.count()); i++) {
    const opts = await selects.nth(i).locator('option').evaluateAll(os => os.map(o => o.value).filter(Boolean));
    if (opts.length) await selects.nth(i).selectOption(opts[0]);
  }
  await page.waitForTimeout(wait);
};

// replace real visitor IPs with documentation ranges (RFC 5737 / RFC 3849), the same IP always maps to the same fake one
export const maskIPs = (scope = 'main') => async page => {
  await page.evaluate(scope => {
    const seen = new Map();
    const fake = (ip, v6) => {
      if (!seen.has(ip)) seen.set(ip, v6 ? `2001:db8::${(seen.size + 1).toString(16)}` : `203.0.113.${10 + seen.size}`);
      return seen.get(ip);
    };
    const v4 = /\b(?:\d{1,3}\.){3}\d{1,3}\b/g;
    const v6 = /\b[0-9a-f]{1,4}(?::[0-9a-f]{0,4}){2,7}\b/gi;
    // digit-only matches like 23:33:18 are times, not addresses
    const isV6 = m => m.includes('::') || /[a-f]/i.test(m) || (m.match(/:/g) || []).length >= 4;
    const swap = s => s.replace(v6, m => (isV6(m) ? fake(m, true) : m)).replace(v4, m => (m.startsWith('127.') || m.startsWith('0.') ? m : fake(m, false)));
    for (const root of document.querySelectorAll(scope)) {
      const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
      for (let n; (n = walker.nextNode()); ) {
        const next = swap(n.textContent);
        if (next !== n.textContent) n.textContent = next;
      }
      // form fields and editors hold their text as values, not text nodes
      root.querySelectorAll('input:not([type=hidden]), textarea').forEach(i => { if (i.value) i.value = swap(i.value); });
    }
  }, scope);
};

// hide the health-issue notices (an "Error"/"Warning" box with a Close button) that some pages pop up
export const hideNotices = async page => {
  await page.evaluate(() => {
    for (const btn of document.querySelectorAll('button')) {
      if (btn.textContent.trim() !== 'Close') continue;
      let box = btn.parentElement;
      for (let i = 0; i < 4 && box && !/^\s*(Error|Warning|Notice)/.test(box.innerText); i++) box = box.parentElement;
      if (box && /^\s*(Error|Warning|Notice)/.test(box.innerText)) box.style.display = 'none';
    }
  });
};

// the onboarding wizard pops up on every dashboard load for the demo user
export const hideOnboarding = async page => {
  await page.evaluate(() => {
    const h = [...document.querySelectorAll('h2')].find(e => e.textContent.includes("Let's set up your account"));
    const overlay = h && h.closest('.fixed.inset-0');
    if (overlay) overlay.style.display = 'none';
  });
};

// remove an element found by its visible text, and hide the closest box around it
export const hideText = (text, levels = 1) => async page => {
  await page.evaluate(([text, levels]) => {
    const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
    for (let n; (n = walker.nextNode()); ) {
      if (!n.textContent.includes(text)) continue;
      let el = n.parentElement;
      for (let i = 1; i < levels && el.parentElement; i++) el = el.parentElement;
      el.style.display = 'none';
      return;
    }
  }, [text, levels]);
};

// tag the card around a heading so a shot can crop to [data-shot=card]
export const markCard = title => async page => {
  await page.evaluate(title => {
    document.querySelectorAll('[data-shot]').forEach(e => e.removeAttribute('data-shot'));
    const walker = document.createTreeWalker(document.querySelector('main'), NodeFilter.SHOW_TEXT);
    const matches = [];
    for (let n; (n = walker.nextNode()); ) if (n.textContent.trim() === title && n.parentElement.offsetParent) matches.push(n);
    // a card title is a heading, so prefer those over icon labels with the same text
    matches.sort((a, b) => /^H\d$/.test(b.parentElement.tagName) - /^H\d$/.test(a.parentElement.tagName));
    for (const n of matches) {
      let el = n.parentElement;
      while (el && !(/\brounded/.test(el.className) && /\bborder\b/.test(el.className) && el.getBoundingClientRect().height > 60)) el = el.parentElement;
      if (el) { el.setAttribute('data-shot', 'card'); return; }
    }
  }, title);
};

// switch a client-side tab by its link text
export const tab = name => async page => {
  // tab labels carry icon markup, so match on innerText in the page
  await page.evaluate(name => {
    const a = [...document.querySelectorAll('a')].find(a => /activeTab/.test(a.getAttribute('@click') || '') && a.innerText.trim() === name);
    a.click();
    // tag the card that holds the tab bar and its panel, so the crop can stop there
    document.querySelectorAll('[data-shot]').forEach(e => e.removeAttribute('data-shot'));
    let el = a.parentElement;
    while (el && !(/\brounded/.test(el.className) && /\bborder\b/.test(el.className) && el.getBoundingClientRect().height > 150)) el = el.parentElement;
    if (el) el.setAttribute('data-shot', 'tabs');
  }, name);
  await page.waitForTimeout(1200);
};

// tag the grid row that holds a section heading, so a shot can crop to [data-shot=section]
export const markSection = title => async page => {
  await page.evaluate(title => {
    document.querySelectorAll('[data-shot=section]').forEach(e => e.removeAttribute('data-shot'));
    const h = [...document.querySelectorAll('main h2, main h3')].find(e => e.textContent.trim() === title && e.offsetParent);
    const row = h && (h.closest('.grid') || h.parentElement);
    if (row) row.setAttribute('data-shot', 'section');
  }, title);
};

// tag the grid row around the h2 with this id (or text), for pages whose headings carry badges or links
export const markSectionId = id => async page => {
  await page.evaluate(id => {
    document.querySelectorAll('[data-shot=section]').forEach(e => e.removeAttribute('data-shot'));
    const h = document.getElementById(id) || [...document.querySelectorAll('main h2')].find(e => e.textContent.trim().startsWith(id));
    const row = h && (h.closest('.grid') || h.parentElement.parentElement);
    if (row) row.setAttribute('data-shot', 'section');
  }, id);
};

// click the element whose Alpine @click sets a value, e.g. clickAlpine("tab = 'images'")
export const clickAlpine = (expr, wait = 1200) => async page => {
  await page.evaluate(expr => [...document.querySelectorAll('[\\@click]')].find(e => e.getAttribute('@click').startsWith(expr)).click(), expr);
  await page.waitForTimeout(wait);
};

// tick the bulk-select checkboxes of these rows (shared partials/_bulk.html markup)
export const bulkSelect = (keys, wait = 400) => async page => {
  // a number ticks the first N rows that can run bulk actions
  if (typeof keys === 'number') {
    const boxes = page.locator('input.bulk-select-box:not([data-bulk-skip="*"])');
    // a DOM click, the fixed bar that appears after the first one covers the bottom rows
    for (let i = 0; i < keys; i++) await boxes.nth(i).evaluate(el => el.click());
  } else {
    for (const k of keys) await page.locator(`input.bulk-select-box[value="${k}"]`).evaluate(el => el.click());
  }
  // check() scrolls rows into view, go back up so the crop starts at the table
  await page.evaluate(() => document.querySelectorAll('*').forEach(el => { if (el.scrollTop) el.scrollTop = 0; }));
  await page.evaluate(() => window.scrollTo(0, 0));
  await page.waitForTimeout(wait);
};

// drop the table rows whose bulk checkbox isn't one of keep, so a shared test server only shows the fixtures
export const onlyRows = (table, keep) => async page => {
  await page.evaluate(([table, keep]) => {
    document.querySelectorAll(`${table} tbody input.bulk-select-box`).forEach(cb => { const tr = cb.closest('tr'); if (tr) tr.dataset.hadBox = ''; });
    document.querySelectorAll(`${table} tbody input.bulk-select-box`).forEach(cb => {
      if (keep.some(k => cb.value === k || cb.value.startsWith(k))) return;
      const row = cb.closest('tr');
      // rows like remote access hold one checkbox per host, drop just that host's line
      if (row && row.querySelectorAll('input.bulk-select-box').length > 1) cb.parentElement.remove();
      else (row || cb.parentElement).remove();
    });
    document.querySelectorAll(`${table} tbody tr`).forEach(tr => {
      if (tr.dataset.hadBox !== undefined && !tr.querySelector('input.bulk-select-box')) tr.remove();
    });
    // the "Total ...: N" badge above the table still counts the dropped rows
    const left = document.querySelectorAll(`${table} tbody input.bulk-select-box`).length;
    const header = document.querySelector('main');
    for (const el of header.querySelectorAll('span, p')) {
      if (/^Total [a-z ]+:/i.test(el.textContent.trim())) {
        const badge = el.querySelector('span');
        if (badge && /^\d+$/.test(badge.textContent.trim())) badge.textContent = String(left);
      }
    }
  }, [table, keep]);
};

// the bulk bar is fixed to the bottom of the window, for a shot put it back in the flow right under the table
export const fitBulkBar = () => async page => {
  await page.evaluate(() => {
    const bar = document.getElementById('bulk-actions-bar');
    bar.style.position = 'static';
    bar.style.left = bar.style.right = '';
    // the spacer that keeps the fixed bar off the last rows
    const spacer = bar.previousElementSibling;
    if (spacer && spacer.getAttribute('aria-hidden') === 'true') spacer.style.display = 'none';
    // tall wrappers like min-h-[90vh] would leave a gap between the table and the bar
    bar.parentElement.querySelectorAll('[class*="min-h-"]').forEach(el => { el.style.minHeight = '0'; });
  });
  await page.waitForTimeout(300);
};

// keep only the first n rows of a long table
export const firstRows = (table, n) => async page => {
  await page.evaluate(([table, n]) => {
    [...document.querySelectorAll(`${table} tbody tr`)].slice(n).forEach(tr => tr.remove());
  }, [table, n]);
};
