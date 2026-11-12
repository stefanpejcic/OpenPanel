// sidebar menu filter for the theme pages, driven by the search box's x-effect
function opFilterNav(root, q) {
  q = (q || '').trim().toLowerCase();
  var any = false;
  root.querySelectorAll('.op-nav').forEach(function (nav) {
    var shown = 0;
    nav.querySelectorAll('.op-nav-item').forEach(function (a) {
      var hit = !q || a.textContent.toLowerCase().indexOf(q) !== -1;
      a.style.display = hit ? '' : 'none';
      if (hit) shown++;
    });
    nav.style.display = shown ? '' : 'none';
    if (shown) any = true;
  });
  var empty = root.querySelector('.op-nav-empty');
  if (empty) empty.hidden = any;
}

// Enter in the search box opens the first item still showing
function opFirstNavHit(root) {
  var hit = Array.prototype.find.call(root.querySelectorAll('.op-nav-item'), function (a) { return a.style.display !== 'none' && a.closest('.op-nav').style.display !== 'none'; });
  if (hit) location.href = hit.getAttribute('href');
}

// menus open fixed next to their button so a table's scroll box can't clip them, same as the panel's op-kit.js
window.opMenu = function () {
  return {
    open: false,
    toggle() {
        this.open = !this.open;
        if (!this.open) return;
        // $nextTick waits a task, so the menu would paint once at its default spot before moving; place it before the next paint and keep it hidden until then
        const menu = this.$refs.menu;
        // fixed at the corner while hidden: laid out inside the row it would overflow the table and flash scrollbars
        Object.assign(menu.style, { position: 'fixed', top: '0px', left: '0px', right: 'auto', visibility: 'hidden' });
        const show = () => { this.place(); menu.style.visibility = ''; };
        queueMicrotask(() => (getComputedStyle(menu).display !== 'none' ? show() : this.$nextTick(show)));
    },
    init() {
      // a fixed menu doesn't move with the page, so it follows its button and only closes once the button leaves the screen
      const follow = () => {
        if (!this.open) return;
        const r = this.$refs.btn.getBoundingClientRect();
        if (r.bottom < 0 || r.top > window.innerHeight) { this.open = false; return; }
        this.place();
      };
      window.addEventListener('scroll', follow, true);
      window.addEventListener('resize', follow);
    },
    place() {
      const menu = this.$refs.menu, btn = this.$refs.btn.getBoundingClientRect();
      menu.style.position = 'fixed';
      // data-align="start" lines the menu up with the button's left edge instead of its right
      if (menu.dataset.align === 'start') {
        menu.style.right = 'auto';
        menu.style.left = Math.max(8, Math.min(btn.left, window.innerWidth - menu.offsetWidth - 8)) + 'px';
      } else {
        menu.style.left = 'auto';
        menu.style.right = Math.max(8, window.innerWidth - btn.right) + 'px';
      }
      const h = menu.offsetHeight;
      menu.style.top = (btn.bottom + 4 + h <= window.innerHeight ? btn.bottom + 4 : Math.max(8, btn.top - h - 4)) + 'px';
    },
  };
};

// list rows: data-href opens the row's page, data-row-check ticks its checkbox, data-row-menu opens its actions; controls inside the row keep their own click
if (!window.opRowClicks) {
  window.opRowClicks = true;
  document.addEventListener('click', function (e) {
    const row = e.target.closest('tr[data-href], tr[data-row-check], tr[data-row-menu]');
    if (!row || e.target.closest('a, button, input, select, textarea, label, .op-menu, .op-cell-menu')) return;
    if (String(window.getSelection ? window.getSelection() : '')) return;
    if (row.hasAttribute('data-row-check')) {
      const box = row.querySelector('input[type="checkbox"]');
      if (box && !box.disabled) box.click();
    } else if (row.hasAttribute('data-row-menu')) {
      const btn = row.querySelector('.op-cell-menu button');
      // after the menu's own outside-click handler has run, or it would close again straight away
      if (btn) setTimeout(function () { btn.click(); }, 0);
    } else if (e.ctrlKey || e.metaKey) {
      window.open(row.dataset.href, '_blank');
    } else {
      window.location.href = row.dataset.href;
    }
  });
}

// top bar tools search, separate from the sidebar's menu search; Ctrl K opens it
var OP_TOOLS = [
  ['File Manager', 'files.html', 'Files'],
  ['Upload files', 'files.html', 'Files'],
  ['Backups', 'backups.html', 'Files'],
  ['Backup Wizard', 'backup-wizard.html', 'Files'],
  ['Disk Usage', 'disk.html', 'Files'],
  ['Inodes', 'disk.html', 'Files'],
  ['Domains', 'domains.html', 'Domains'],
  ['Add a Domain', 'domain-new.html', 'Domains'],
  ['DNS Zone Editor', 'dns.html', 'Domains'],
  ['SSL Certificates', 'ssl.html', 'Domains'],
  ['Dynamic DNS', 'domains-dynamic-dns.html', 'Domains'],
  ['Redirects', 'domains-redirects.html', 'Domains'],
  ['Websites', 'websites.html', 'Websites'],
  ['Site Manager', 'websites.html', 'Websites'],
  ['WordPress', 'wordpress.html', 'Websites'],
  ['Auto Installer', 'install.html', 'Websites'],
  ['MySQL Databases', 'mysql.html', 'Databases'],
  ['MySQL Users', 'mysql.html', 'Databases'],
  ['phpMyAdmin', 'mysql.html', 'Databases'],
  ['PostgreSQL', 'postgresql.html', 'Databases'],
  ['Email Accounts', 'email.html', 'Email'],
  ['Webmail', 'email.html', 'Email'],
  ['Deliverability', 'deliverability.html', 'Email'],
  ['Email Migration', 'migrate.html', 'Email'],
  ['PHP Versions', 'php.html', 'Configure'],
  ['PHP Options', 'php-options.html', 'Configure'],
  ['Cron Jobs', 'cron.html', 'Configure'],
  ['Web Server Config', 'webserver.html', 'Configure'],
  ['Redis and Cache', 'cache.html', 'Configure'],
  ['Statistics', 'statistics.html', 'Monitor'],
  ['Usage History', 'statistics-history.html', 'Monitor'],
  ['Daily Summary', 'statistics-summary.html', 'Monitor'],
  ['Visitors', 'visitors.html', 'Monitor'],
  ['Access Logs', 'domains-access-logs.html', 'Monitor'],
  ['Raw access logs', 'domains-access-logs.html', 'Monitor'],
  ['Error Logs', 'error-logs.html', 'Monitor'],
  ['PHP error log', 'error-logs.html', 'Monitor'],
  ['Services', 'services.html', 'Monitor'],
  ['Security', 'security.html', 'Monitor'],
  ['Containers', 'containers.html', 'Advanced'],
  ['Terminal', 'terminal.html', 'Advanced'],
  ['Container logs', 'containers-logs.html', 'Advanced'],
  ['Software Versions', 'containers-versions.html', 'Advanced'],
  ['Docker image tags', 'containers-versions.html', 'Advanced'],
  ['Account', 'account.html', 'Account'],
  ['Two-factor authentication', 'account.html', 'Account'],
  ['API keys', 'account.html', 'Account']
];
window.opToolSearch = function () {
  return {
    open: false, q: '', active: 0,
    get results() {
      var q = this.q.trim().toLowerCase();
      if (!q) return [];
      return OP_TOOLS.filter(function (t) { return (t[0] + ' ' + t[2]).toLowerCase().indexOf(q) !== -1; }).slice(0, 8);
    },
    show() { this.open = true; this.active = 0; this.$nextTick(() => this.$refs.q.focus()); },
    hide() { this.open = false; this.q = ''; },
    move(d) { var n = this.results.length; if (n) this.active = (this.active + d + n) % n; },
    go() { var t = this.results[this.active]; if (t) location.href = t[1]; },
  };
};

// AI assistant drawer: same strings, states and approval flow as the panel's partials/_ai_assistant.html + ai-assistant.js, no model behind it
(function () {
  var S = {
    greeting: 'Hi! Ask me anything about your hosting account and this panel.',
    placeholder: 'Ask a question...',
    notConfigured: 'The AI Assistant is not configured yet, please contact your hosting provider.',
    usage: 'Requests this month', unlimited: 'unlimited',
    s1: 'How do I add a new domain?', s2: 'How do I create an email account?', s3: 'How do I install WordPress?',
    s4: 'What can I do on this page?', s5: 'Which domains do I have?', s6: 'Create an email account for me',
    change: 'Change', destructive: 'Destructive', approve: 'Approve', always: 'Always allow', deny: 'Decline',
    destructiveNote: 'This can delete or overwrite data. Check the details before you approve.',
    state_approved: 'Approved', state_declined: 'Declined', state_skipped: 'Not answered', state_expired: 'Expired',
    mode_off: "It answers questions using the panel documentation. It can't see your data or make changes for you.",
    mode_read: "It answers questions using the panel documentation and can look up information in your account, but it can't make changes.",
    mode_all: 'It answers questions and can look up information or make changes in your account. Changes only run after you approve them, and everything it does is recorded in your Activity log.',
    allowedTitle: 'Always allowed actions',
    allowedNone: 'None yet. Click Always allow on an action to skip the approval next time. Actions that delete or overwrite data always ask.',
    forget: 'Remove'
  };
  var SHORT = { off: 'Answers from the documentation', read: 'Can look up your account, no changes', all: 'Changes need your approval' };
  // the test install: provider set but no model, read-only actions, 100 requests a month
  var REAL = { name: 'AI Assistant', configured: false, actions: 'read', limit: 100, used: 0, allowed: [] };
  var DEMO = { name: 'AI Assistant', configured: true, actions: 'all', limit: 100, used: 3, allowed: [{ title: 'Create cron job' }] };
  var DEMO_CHAT = [
    { role: 'user', content: 'Which domains do I have?' },
    { role: 'action', content: 'List all domains', ok: true },
    { role: 'assistant', content: 'You have **23 domains**, all under `tests.openpanel.org`. A few of them:\n- wordpress.tests.openpanel.org\n- php.tests.openpanel.org\n- shared.tests.openpanel.org\n\nSee all of them on [Domains](domains.html).' },
    { role: 'user', content: 'Create a database called shop' },
    { role: 'confirm', state: 'pending', card: { kind: 'change', title: 'Create a new database', request: 'POST /api/mysql/databases', fields: [['name', 'testinguser_shop']], remember: true } }
  ];
  var CANNED = {};
  CANNED[S.s1] = 'Open **Domains** and click **Add a Domain**. Enter the name, pick a document root (it defaults to `/var/www/html/`) and save. [Add a Domain](domain-new.html)';
  CANNED[S.s5] = DEMO_CHAT[2].content;
  CANNED[S.s6] = 'Sure. Which address should I create, and for which domain? For example `info@wordpress.tests.openpanel.org`.';
  var I = function (p) { return '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/>' + p + '</svg>'; };
  var IC = {
    info: I('<path d="M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0"/><path d="M12 9h.01"/><path d="M11 12h1v4h1"/>'),
    plus: I('<path d="M8 9h8"/><path d="M8 13h6"/><path d="M12.01 18.594l-4.01 2.406v-3h-2a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12a3 3 0 0 1 3 3v5.5"/><path d="M16 19h6"/><path d="M19 16v6"/>'),
    x: I('<path d="M18 6l-12 12"/><path d="M6 6l12 12"/>'),
    send: I('<path d="M12 5l0 14"/><path d="M18 11l-6 -6"/><path d="M6 11l6 -6"/>'),
    spark: I('<path d="M16 18a2 2 0 0 1 2 2a2 2 0 0 1 2 -2a2 2 0 0 1 -2 -2a2 2 0 0 1 -2 2zm0 -12a2 2 0 0 1 2 2a2 2 0 0 1 2 -2a2 2 0 0 1 -2 -2a2 2 0 0 1 -2 2zm-7 12a6 6 0 0 1 6 -6a6 6 0 0 1 -6 -6a6 6 0 0 1 -6 6a6 6 0 0 1 6 6z"/>'),
    check: I('<path d="M5 12l5 5l10 -10"/>'),
    alert: I('<path d="M12 9v4"/><path d="M10.363 3.591l-8.106 13.534a1.914 1.914 0 0 0 1.636 2.871h16.214a1.914 1.914 0 0 0 1.636 -2.87l-8.106 -13.536a1.914 1.914 0 0 0 -3.274 0z"/><path d="M12 16h.01"/>')
  };
  var esc = function (s) { return String(s).replace(/[&<>"']/g, function (c) { return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]; }); };
  function md(src) {
    var inline = function (t) { return esc(t).replace(/\*\*(.+?)\*\*/g, '<b>$1</b>').replace(/`([^`]+)`/g, '<code>$1</code>').replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, '<a href="$2">$1</a>'); };
    var out = '', inList = false;
    src.split('\n').forEach(function (l) {
      if (/^- /.test(l)) { if (!inList) { out += '<ul>'; inList = true; } out += '<li>' + inline(l.slice(2)) + '</li>'; return; }
      if (inList) { out += '</ul>'; inList = false; }
      if (l.trim()) out += '<p>' + inline(l) + '</p>';
    });
    return out + (inList ? '</ul>' : '');
  }

  var demo = /(^|[?&])ai=demo(&|$)/.test(location.search), info, messages, drawer, scrim, list, about, input, send, btns = [], busy = false;
  function reset() {
    info = JSON.parse(JSON.stringify(demo ? DEMO : REAL));
    messages = demo ? JSON.parse(JSON.stringify(DEMO_CHAT)) : [];
  }

  function build() {
    var link = document.createElement('link');
    link.rel = 'stylesheet'; link.href = 'ai.css';
    document.head.appendChild(link);
    scrim = document.createElement('div');
    scrim.className = 'ai-scrim'; scrim.hidden = true;
    drawer = document.createElement('aside');
    drawer.className = 'ai-drawer'; drawer.id = 'ai-assistant-drawer'; drawer.hidden = true;
    drawer.setAttribute('role', 'dialog'); drawer.setAttribute('aria-labelledby', 'ai-title');
    drawer.innerHTML =
      '<div class="ai-head"><span class="ai-mark">AI</span><div class="ai-head-text"><b id="ai-title"></b><span data-ai-short></span></div>' +
      '<button class="op-icon-btn" type="button" data-ai="about" aria-pressed="false" aria-label="About" title="About">' + IC.info + '</button>' +
      '<button class="op-icon-btn" type="button" data-ai="clear" aria-label="New chat" title="New chat">' + IC.plus + '</button>' +
      '<button class="op-icon-btn" type="button" data-ai="close" aria-label="Close" title="Close">' + IC.x + '</button></div>' +
      '<div class="ai-about" hidden></div>' +
      '<div class="ai-list" aria-live="polite"></div>' +
      '<form class="ai-form"><textarea rows="1" maxlength="4000" aria-label="' + S.placeholder + '" placeholder="' + S.placeholder + '"></textarea><button class="ai-send" type="submit" aria-label="Send" title="Send">' + IC.send + '</button></form>' +
      '<p class="ai-foot">AI can make mistakes. <span data-ai-proto></span></p>';
    document.body.append(scrim, drawer);
    list = drawer.querySelector('.ai-list'); about = drawer.querySelector('.ai-about');
    input = drawer.querySelector('textarea'); send = drawer.querySelector('.ai-send');
    drawer.querySelector('[data-ai="close"]').addEventListener('click', close);
    drawer.querySelector('[data-ai="clear"]').addEventListener('click', function () { messages = []; render(); if (info.configured) input.focus(); });
    drawer.querySelector('[data-ai="about"]').addEventListener('click', function () { about.hidden = !about.hidden; this.setAttribute('aria-pressed', String(!about.hidden)); });
    drawer.querySelector('form').addEventListener('submit', function (e) { e.preventDefault(); ask(input.value); });
    input.addEventListener('keydown', function (e) { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); ask(input.value); } });
    input.addEventListener('input', function () { input.style.height = 'auto'; input.style.height = Math.min(input.scrollHeight, 140) + 'px'; });
    scrim.addEventListener('click', close);
    drawer.addEventListener('click', function (e) { if (e.target.closest('[data-ai-toggle]')) { demo = !demo; reset(); paint(); } });
    reset(); paint();
  }

  function paint() {
    drawer.querySelector('#ai-title').textContent = info.name;
    drawer.querySelector('[data-ai-short]').textContent = info.configured ? SHORT[info.actions] : 'Not configured';
    drawer.querySelector('[data-ai-proto]').innerHTML = 'Preview: <button type="button" data-ai-toggle>' + (demo ? 'show the test install' : 'show a sample chat') + '</button>';
    input.disabled = !info.configured; send.disabled = !info.configured;
    showInfo(); render();
  }

  function showInfo() {
    var pct = info.limit > 0 ? Math.min(100, Math.round(info.used / info.limit * 100)) : 0;
    var html = '<p>' + esc(S['mode_' + info.actions]) + '</p><p>AI can make mistakes, double-check important steps before you apply them.</p>' +
      '<div class="ai-meter"><span>' + S.usage + '</span><b>' + (info.limit > 0 ? info.used + ' / ' + info.limit : info.used + ' (' + S.unlimited + ')') + '</b></div>' +
      (info.limit > 0 ? '<div class="ai-bar"><i style="width:' + pct + '%"></i></div>' : '');
    if (info.actions === 'all') {
      html += '<div class="ai-allowed"><b>' + S.allowedTitle + '</b>' + (info.allowed.length ? info.allowed.map(function (a, i) {
        return '<div class="ai-allowed-row"><span>' + esc(a.title) + '</span><button class="op-btn op-btn-ghost op-btn-sm" type="button" data-forget="' + i + '" style="color:var(--color-danger)">' + S.forget + '</button></div>';
      }).join('') : '<p>' + S.allowedNone + '</p>') + '</div>';
    }
    about.innerHTML = html;
    about.querySelectorAll('[data-forget]').forEach(function (b) { b.addEventListener('click', function () { info.allowed.splice(+b.dataset.forget, 1); showInfo(); }); });
  }

  function bubble(role, text) {
    var el = document.createElement('div');
    el.className = 'ai-msg ai-msg-' + role;
    if (role === 'assistant') el.innerHTML = md(text);
    else if (role === 'error') el.innerHTML = IC.alert + '<span>' + esc(text) + '</span>';
    else el.textContent = text;
    list.appendChild(el);
  }

  function card(m) {
    var c = m.card, el = document.createElement('div');
    el.className = 'ai-card' + (c.kind === 'destructive' ? ' is-destructive' : '');
    var html = '<div class="ai-card-head"><span class="ai-kind">' + (c.kind === 'destructive' ? S.destructive : S.change) + '</span><b>' + esc(c.title) + '</b></div><code class="ai-req">' + esc(c.request) + '</code>';
    if (c.fields.length) html += '<dl class="ai-fields">' + c.fields.map(function (f) { return '<dt>' + esc(f[0]) + '</dt><dd>' + esc(f[1]) + '</dd>'; }).join('') + '</dl>';
    if (c.kind === 'destructive') html += '<p class="ai-note">' + S.destructiveNote + '</p>';
    if (m.state === 'pending') {
      html += '<div class="ai-card-actions"><button class="op-btn op-btn-primary op-btn-sm" type="button" data-d="allow">' + S.approve + '</button>' +
        (c.remember ? '<button class="op-btn op-btn-secondary op-btn-sm" type="button" data-d="always">' + S.always + '</button>' : '') +
        '<button class="op-btn op-btn-ghost op-btn-sm" type="button" data-d="deny">' + S.deny + '</button></div>';
    } else html += '<p class="ai-state">' + S['state_' + m.state] + '</p>';
    el.innerHTML = html;
    el.querySelectorAll('[data-d]').forEach(function (b) { b.addEventListener('click', function () { decide(m, b.dataset.d); }); });
    list.appendChild(el);
  }

  function render() {
    list.innerHTML = '';
    if (!info.configured) { bubble('error', S.notConfigured); return; }
    if (!messages.length) {
      bubble('assistant', S.greeting);
      var wrap = document.createElement('div');
      wrap.className = 'ai-suggest';
      (info.actions === 'all' ? ['s1', 's5', 's6', 's4'] : ['s1', 's2', 's3', 's4']).forEach(function (k) {
        var b = document.createElement('button');
        b.type = 'button'; b.innerHTML = IC.spark + '<span>' + esc(S[k]) + '</span>';
        b.addEventListener('click', function () { ask(S[k]); });
        wrap.appendChild(b);
      });
      list.appendChild(wrap);
    }
    messages.forEach(function (m) {
      if (m.role === 'action') {
        var el = document.createElement('div');
        el.className = 'ai-action' + (m.ok ? '' : ' is-declined');
        el.innerHTML = (m.ok ? IC.check : IC.x) + '<span>' + esc(m.content) + '</span>';
        list.appendChild(el);
      } else if (m.role === 'confirm') card(m);
      else bubble(m.role, m.content);
    });
    if (busy) { var t = document.createElement('div'); t.className = 'ai-thinking'; t.setAttribute('aria-label', 'Thinking...'); t.innerHTML = '<i></i><i></i><i></i>'; list.appendChild(t); }
    list.scrollTop = list.scrollHeight;
  }

  function reply(push) {
    busy = true; send.disabled = true; render();
    setTimeout(function () { busy = false; send.disabled = false; push(); info.used++; showInfo(); render(); }, 700);
  }

  function decide(m, d) {
    m.state = d === 'deny' ? 'declined' : 'approved';
    if (d === 'always' && !info.allowed.some(function (a) { return a.title === m.card.title; })) info.allowed.push({ title: m.card.title });
    if (d === 'deny') {
      messages.push({ role: 'action', content: 'Declined: ' + m.card.title, ok: false });
      reply(function () { messages.push({ role: 'assistant', content: "OK, I won't create it." }); });
    } else {
      messages.push({ role: 'action', content: m.card.title, ok: true });
      reply(function () { messages.push({ role: 'assistant', content: 'Done. **testinguser_shop** is created. Assign a user to it on [MySQL](mysql-users.html) so an app can connect.' }); });
    }
    showInfo(); render();
  }

  function ask(q) {
    q = (q || '').trim();
    if (!q || busy || !info.configured) return;
    input.value = ''; input.style.height = 'auto';
    messages.forEach(function (m) { if (m.role === 'confirm' && m.state === 'pending') m.state = 'skipped'; });
    messages.push({ role: 'user', content: q });
    reply(function () {
      if (q === S.s5) messages.push({ role: 'action', content: 'List all domains', ok: true });
      var a = CANNED[q] || (q === S.s4 ? 'You are on **' + document.title.split(' - ')[0] + '**. Ask me how to do something here, or what a setting means.' : 'This is a preview, so no model answers here. In the panel the assistant replies from the OpenPanel documentation and, when your provider allows it, your account data.');
      messages.push({ role: 'assistant', content: a });
    });
    render();
  }

  function open() {
    if (!drawer) build();
    scrim.hidden = false; drawer.hidden = false;
    requestAnimationFrame(function () { scrim.classList.add('is-open'); drawer.classList.add('is-open'); });
    btns.forEach(function (b) { b.setAttribute('aria-expanded', 'true'); });
    setTimeout(function () { (info.configured ? input : drawer.querySelector('[data-ai="close"]')).focus(); }, 60);
  }
  function close() {
    if (!drawer || drawer.hidden) return;
    scrim.classList.remove('is-open'); drawer.classList.remove('is-open');
    btns.forEach(function (b) { b.setAttribute('aria-expanded', 'false'); });
    setTimeout(function () { if (!drawer.classList.contains('is-open')) { drawer.hidden = true; scrim.hidden = true; } }, 220);
    if (btns[0]) btns[0].focus();
  }

  document.addEventListener('DOMContentLoaded', function () {
    btns = [].slice.call(document.querySelectorAll('.op-ai'));
    btns.forEach(function (b) {
      b.setAttribute('aria-controls', 'ai-assistant-drawer'); b.setAttribute('aria-expanded', 'false');
      b.addEventListener('click', function () { if (drawer && !drawer.hidden) close(); else open(); });
    });
    // #ai opens the drawer on load, handy for links and screenshots
    if (location.hash === '#ai') open();
  });
  document.addEventListener('keydown', function (e) { if (e.key === 'Escape' && drawer && !drawer.hidden) close(); });
})();

// sidebar Favorites group and the top bar star, kept in localStorage; starts with the account's real favorites on the test install
(function () {
  var KEY = 'op-favorites', FOLD = 'op-favorites-folded';
  var DEFAULTS = [
    { title: 'Restore & Download', href: 'backup-restore.html' },
    { title: 'Web Firewall', href: 'security.html' },
    { title: 'DNS Zone Editor', href: 'dns.html' },
    { title: 'MySQL Databases', href: 'mysql.html' }
  ];
  var STAR = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 17.75l-6.172 3.245l1.179 -6.873l-5 -4.867l6.9 -1l3.086 -6.253l3.086 6.253l6.9 1l-5 4.867l1.179 6.873z"/></svg>';
  var CHEV = '<svg class="op-nav-chev" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M6 9l6 6l6 -6"/></svg>';
  function read(k, d) { try { var v = localStorage.getItem(k); return v === null ? d : JSON.parse(v); } catch (e) { return d; } }
  function write(k, v) { try { localStorage.setItem(k, JSON.stringify(v)); } catch (e) {} }
  function here() { return (location.pathname.split('/').pop() || 'index.html'); }
  function pageTitle() {
    var b = document.querySelector('.op-crumb b');
    return (b && b.textContent.trim()) || document.title.split(' - ')[0];
  }
  function esc(s) { return String(s).replace(/[&<>"]/g, function (c) { return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c]; }); }

  var nav, star;
  function render() {
    var favs = read(KEY, DEFAULTS), cur = here();
    var list = favs.length
      ? favs.map(function (f) { return '<a class="op-nav-item" href="' + esc(f.href) + '"' + (f.href === cur ? ' aria-current="page"' : '') + '>' + esc(f.title) + '</a>'; }).join('')
      : '<p class="op-nav-favs-empty">Use the star in the top bar to pin a page here.</p>';
    nav.querySelector('.op-nav-favs-list').innerHTML = list;
    if (star) {
      var on = favs.some(function (f) { return f.href === cur; });
      star.setAttribute('aria-pressed', on ? 'true' : 'false');
      star.title = on ? 'Remove from favorites' : 'Add to favorites';
      star.setAttribute('aria-label', star.title);
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    var first = document.querySelector('.op-sidebar .op-nav');
    if (!first) return;
    nav = document.createElement('nav');
    nav.className = 'op-nav op-nav-favs';
    nav.setAttribute('aria-label', 'Favorites');
    if (read(FOLD, false)) nav.setAttribute('data-collapsed', '');
    nav.innerHTML = '<button class="op-nav-label op-nav-toggle" type="button" aria-expanded="' + !read(FOLD, false) + '">' + STAR + ' Favorites ' + CHEV + '</button><div class="op-nav-favs-list"></div>';
    first.parentNode.insertBefore(nav, first);
    nav.querySelector('.op-nav-toggle').addEventListener('click', function () {
      var folded = !nav.hasAttribute('data-collapsed');
      nav.toggleAttribute('data-collapsed', folded);
      this.setAttribute('aria-expanded', String(!folded));
      write(FOLD, folded);
    });

    var ai = document.querySelector('.op-topbar .op-ai');
    if (ai) {
      star = document.createElement('button');
      star.type = 'button';
      star.className = 'op-icon-btn op-fav-btn';
      star.innerHTML = STAR;
      ai.parentNode.insertBefore(star, ai);
      star.addEventListener('click', function () {
        var favs = read(KEY, DEFAULTS), cur = here();
        var i = favs.findIndex(function (f) { return f.href === cur; });
        if (i >= 0) favs.splice(i, 1); else favs.push({ title: pageTitle(), href: cur });
        write(KEY, favs);
        render();
      });
    }
    render();
  });
})();

// sortable headers: click a th's .op-th-sort button to sort that column, again to flip; numbers and sizes sort as numbers, a cell's data-sort wins over its text
window.opSortCol = function (btn) {
  const th = btn.closest('th'), table = th.closest('table'), body = table.tBodies[0], idx = th.cellIndex;
  const dir = th.getAttribute('aria-sort') === 'ascending' ? 'descending' : 'ascending';
  table.querySelectorAll('th[aria-sort]').forEach((h) => h.setAttribute('aria-sort', 'none'));
  th.setAttribute('aria-sort', dir);
  const key = (row) => { const c = row.cells[idx]; return c ? (c.dataset.sort ?? c.textContent.trim()) : ''; };
  const num = (t) => { const m = t.replace(/,/g, '').match(/^-?\d+(\.\d+)?/); if (!m) return NaN; const u = { kb: 1e3, mb: 1e6, gb: 1e9, tb: 1e12 }[(t.match(/[kmgt]b/i) || [''])[0].toLowerCase()] || 1; return parseFloat(m[0]) * u; };
  const rows = [...body.rows];
  rows.sort((a, b) => { const x = key(a), y = key(b), nx = num(x), ny = num(y); const c = !isNaN(nx) && !isNaN(ny) ? nx - ny : x.localeCompare(y, undefined, { numeric: true, sensitivity: 'base' }); return dir === 'ascending' ? c : -c; });
  rows.forEach((r) => body.appendChild(r));
};

// busy toast for slow server actions: says what's going on and whether leaving the page is safe; forms opt in with data-busy="<title>"
window.opBusy = window.opBusy || (function () {
  const svg = (d) => '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">' + d + '</svg>';
  const icons = { done: svg('<path d="M5 12l5 5l10 -10"/>'), failed: svg('<path d="M12 9v4"/><path d="M12 16v.01"/><path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0 -18 0"/>'), safe: svg('<path d="M5 12l5 5l10 -10"/>'), wait: svg('<path d="M12 7v5l3 3"/><path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0 -18 0"/>'), x: svg('<path d="M18 6l-12 12"/><path d="M6 6l12 12"/>') };
  let el, timer, started, hideAfter;
  const q = (s) => el.querySelector(s);
  function build() {
    if (el) return;
    el = document.createElement('div');
    el.className = 'op-busy op-scope';
    el.hidden = true;
    el.setAttribute('role', 'status');
    el.setAttribute('aria-live', 'polite');
    el.innerHTML = '<div class="op-busy-head"><span class="op-busy-ico"></span><div><b class="op-busy-title"></b><span class="op-busy-text"></span></div><button type="button" class="op-busy-x" aria-label="Dismiss">' + icons.x + '</button></div><button type="button" class="op-busy-more"></button><ul class="op-busy-steps" hidden></ul><div class="op-busy-foot"><span class="op-busy-note"></span><span class="op-busy-time" aria-hidden="true"></span></div>';
    document.body.appendChild(el);
    q('.op-busy-x').addEventListener('click', () => api.hide());
    q('.op-busy-more').addEventListener('click', () => { const ul = q('.op-busy-steps'); ul.hidden = !ul.hidden; q('.op-busy-more').textContent = ul.hidden ? el.dataset.more : el.dataset.less; });
  }
  const tick = () => { const s = Math.round((Date.now() - started) / 1000); q('.op-busy-time').textContent = s < 60 ? s + ' s' : Math.floor(s / 60) + ' min ' + (s % 60) + ' s'; };
  const api = {
    show(o) {
      build();
      clearTimeout(hideAfter);
      el.classList.remove('is-done', 'is-failed');
      el.setAttribute('role', 'status');
      el.dataset.more = o.more || 'What happens';
      el.dataset.less = o.less || 'Hide steps';
      q('.op-busy-ico').innerHTML = '<span class="op-busy-spin"></span>';
      q('.op-busy-title').textContent = o.title || '';
      q('.op-busy-text').textContent = o.text || '';
      q('.op-busy-text').hidden = !o.text;
      const ul = q('.op-busy-steps');
      ul.replaceChildren(...(o.steps || []).map((t) => Object.assign(document.createElement('li'), { textContent: t })));
      ul.hidden = true;
      const more = q('.op-busy-more');
      more.hidden = !ul.children.length;
      more.textContent = el.dataset.more;
      q('.op-busy-foot').className = 'op-busy-foot ' + (o.leave ? 'is-safe' : 'is-wait');
      q('.op-busy-foot').hidden = false;
      const note = q('.op-busy-note');
      note.innerHTML = o.note ? icons[o.leave ? 'safe' : 'wait'] + '<span></span>' : '';
      if (o.note) note.lastChild.textContent = o.note;
      // the x only hides the toast, so it's offered while running only when leaving is safe
      q('.op-busy-x').hidden = !o.leave;
      started = Date.now(); tick(); clearInterval(timer); timer = setInterval(tick, 1000);
      el.hidden = false;
    },
    done(message, ok = true) {
      if (!el) return;
      clearInterval(timer);
      el.classList.add(ok ? 'is-done' : 'is-failed');
      el.setAttribute('role', ok ? 'status' : 'alert');
      q('.op-busy-ico').innerHTML = icons[ok ? 'done' : 'failed'];
      if (message) { q('.op-busy-text').textContent = message; q('.op-busy-text').hidden = false; }
      q('.op-busy-steps').hidden = true;
      q('.op-busy-more').hidden = true;
      q('.op-busy-foot').hidden = true;
      q('.op-busy-x').hidden = false;
      el.hidden = false;
      if (ok) hideAfter = setTimeout(() => api.hide(), 6000);
    },
    hide() { if (el) { el.hidden = true; clearInterval(timer); clearTimeout(hideAfter); } },
  };
  document.addEventListener('submit', (e) => {
    const f = e.target;
    if (e.defaultPrevented || !(f instanceof HTMLFormElement) || !f.dataset.busy) return;
    api.show({ title: f.dataset.busy, text: f.dataset.busyText, steps: (f.dataset.busySteps || '').split('|').filter(Boolean), note: f.dataset.busyNote, leave: f.dataset.busyLeave !== 'no', more: f.dataset.busyMore, less: f.dataset.busyLess });
  });
  window.addEventListener('pageshow', (e) => { if (e.persisted) api.hide(); });
  document.addEventListener('keydown', (e) => { if (e.key === 'Escape' && el && !el.hidden && !q('.op-busy-x').hidden) api.hide(); });
  return api;
})();
