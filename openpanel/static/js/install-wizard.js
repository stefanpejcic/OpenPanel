// install wizard: one Alpine component drives every app installer, each page passes its fields, steps and run checkpoints
(function () {
  const DEFAULT_T = {
    next: 'Continue', back: 'Back', cancel: 'Cancel', install: 'Install', optional: 'optional', required: 'This field is required.',
    selectDomain: 'Select a domain', pickDomain: 'Pick a domain to install on.', latest: 'Latest', loadingVersions: 'Loading versions...',
    versionsFailed: 'Could not load the version list, the latest version is installed.',
    generate: 'Generate', show: 'Show', hide: 'Hide', copy: 'Copy', copied: 'Copied to clipboard',
    weak: 'Weak password', medium: 'Medium strength password', strong: 'Strong password', tooWeak: 'Password does not meet the required strength.',
    badDB: 'Use letters, numbers and underscores only (up to 32).', badPrefix: 'Use letters, numbers and underscores only.',
    badDBPass: 'Use up to 64 characters without spaces, quotes, backslashes or backticks.',
    badEmail: 'Enter a valid email address.', badUser: 'Use letters, numbers and . _ @ - only, not starting with -.',
    badPass: 'No spaces, and it can not start with -.', badSubdir: 'Use lowercase letters, numbers, dots, dashes and underscores, separate folders with /.',
    badService: 'Start with a lowercase letter, then lowercase letters or numbers.', badPort: 'Use a port from 1000 to 65535.',
    badStartup: 'The file has to be under /var/www/html/ and end in %s.', badGit: 'Only https:// URLs are supported.',
    badProject: 'Use a Composer package (vendor/package) or an https:// archive URL ending in .zip, .tar.gz, .tgz or .tar.',
    badLen: 'Use %s to %s characters.', badNumber: 'Enter a number from %s to %s.',
    phpTooOld: '%s needs PHP %s or newer, this domain uses PHP %s. Change it under PHP first.',
    phpTooNew: '%s needs PHP %s or older, this domain uses PHP %s. Change it under PHP first.',
    phpNone: 'This domain has no PHP version set. Pick one under PHP first.',
    installing: 'Installing %s', installed: '%s is installed', failed: 'Install failed',
    running: 'Running', done: 'Done', failedBadge: 'Failed', keepOpen: 'Keep this page open until it finishes, closing it stops the install.',
    stopped: 'The install stopped without finishing. Check the log below, then try again.',
    connection: 'Lost the connection to the server during the install.', slow: 'This is taking longer than usual. It can still finish, keep the page open.',
    leave: 'The install is still running, leaving stops it.', log: 'Log', showLog: 'Show the install log', hideLog: 'Hide the install log', copyLog: 'Copy log', retry: 'Try again', edit: 'Change settings',
    siteUrl: 'Site', adminUrl: 'Admin login', shownOnce: 'Save these now, the password is only shown here.',
    manager: 'Open Site Manager', another: 'Install another app', open: 'Open', secs: '%s s', mins: '%s min %s s',
    location: 'Location', domain: 'Domain', folder: 'Folder', path: 'Path', url: 'URL', root: 'domain root', what: 'What happens next',
    review: 'Review', phpVersion: 'PHP version', notSet: 'not set', more: 'More options', fewer: 'Fewer options', stepN: 'Step %s'
  };

  const ALNUM = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
  const SYM = '!@#$%^&*()-_=+';
  function rand(n, set) {
    const a = new Uint32Array(n);
    crypto.getRandomValues(a);
    let s = '';
    for (let i = 0; i < n; i++) s += set[a[i] % set.length];
    return s;
  }
  // a leading - reads as a CLI option to the installers (#1207), so passwords start with a letter or digit
  const GEN = {
    password: () => rand(1, ALNUM) + rand(15, ALNUM + SYM),
    alnum8: () => rand(8, ALNUM), alnum12: () => rand(12, ALNUM), alnum16: () => rand(16, ALNUM),
    lower6: () => rand(6, ALNUM.slice(0, 36)).toLowerCase(), lower10: () => rand(10, ALNUM.slice(0, 26)),
    prefix: () => rand(2, ALNUM.slice(0, 52)) + '_'
  };
  function fmt(s) { let i = 1; const args = arguments; return String(s).replace(/%s/g, () => args[i++]); }
  function score(v) {
    if (window.passwordStrengthScore) return window.passwordStrengthScore(v);
    if (!v) return 0;
    let n = 0;
    if (v.length >= 8) n++;
    if (v.length >= 12) n++;
    if (/[a-z]/.test(v)) n++;
    if (/[A-Z]/.test(v)) n++;
    if (/\d/.test(v)) n++;
    if (/[^a-zA-Z0-9]/.test(v)) n++;
    return Math.round(n / 6 * 100);
  }
  function cmpVer(a, b) {
    const x = String(a).split('.').map(Number), y = String(b).split('.').map(Number);
    for (let i = 0; i < Math.max(x.length, y.length); i++) {
      if ((x[i] || 0) !== (y[i] || 0)) return (x[i] || 0) - (y[i] || 0);
    }
    return 0;
  }

  // mirrored server side in cmsapp.ValidateInstallForm, appinstall and phpapp
  const CHECK = {
    dbname: (v) => /^[A-Za-z0-9_]{1,32}$/.test(v) || 'badDB',
    dbprefix: (v) => /^[A-Za-z0-9_]{1,16}$/.test(v) || 'badPrefix',
    dbpass: (v) => (/^[\x21-\x7e]{1,64}$/.test(v) && !/['"\\`]/.test(v)) || 'badDBPass',
    email: (v) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v) || 'badEmail',
    username: (v) => (/^[A-Za-z0-9._@-]{1,60}$/.test(v) && v[0] !== '-') || 'badUser',
    adminpass: (v) => (!/\s/.test(v) && v[0] !== '-' && v.length <= 128) || 'badPass',
    service: (v) => /^[a-z][a-z0-9]{0,31}$/.test(v) || 'badService',
    port: (v) => (/^\d+$/.test(v) && +v >= 1000 && +v <= 65535) || 'badPort',
    https: (v) => /^https:\/\/[^\s'"]+$/.test(v) || 'badGit',
    project: (v) => /^[a-z0-9]([_.\-a-z0-9]*[a-z0-9])?\/[a-z0-9]([_.\-a-z0-9]*[a-z0-9])?$/.test(v) || /^https:\/\/[^\s'"]+\.(zip|tar\.gz|tgz|tar)$/i.test(v) || 'badProject'
  };

  function ndjson(endpoint, body, onEvent) {
    // handlers read r.ParseForm, which skips multipart bodies, so send it urlencoded unless there's a file
    if (body instanceof FormData && ![...body.values()].some((v) => v instanceof File)) body = new URLSearchParams(body);
    return fetch(endpoint, { method: 'POST', body: body, headers: { 'X-CSRF-TOKEN': window.csrf_token || '' } }).then((res) => {
      if (!res.body) return res.text().then((txt) => txt.split('\n').forEach(line));
      const reader = res.body.getReader();
      const dec = new TextDecoder('utf-8');
      let buf = '';
      function line(l) {
        l = l.trim();
        if (!l) return;
        try { onEvent(JSON.parse(l)); } catch (e) { /* not json, skip */ }
      }
      function pump() {
        return reader.read().then(({ done, value }) => {
          if (done) { buf += dec.decode(); line(buf); return; }
          buf += dec.decode(value, { stream: true });
          const parts = buf.split('\n');
          buf = parts.pop();
          parts.forEach(line);
          return pump();
        });
      }
      return pump();
    });
  }

  window.opInstallWizard = function (cfg) {
    return {
      cfg: cfg, t: Object.assign({}, DEFAULT_T, cfg.t || {}),
      steps: cfg.steps, fields: cfg.fields || [], run: cfg.run || [], domains: cfg.domains || [],
      step: 0, phase: 'form', domainId: '', subdir: '',
      v: {}, touched: {}, errs: {}, hints: {}, opts: {}, optsState: {}, reveal: {}, more: {},
      runIdx: -1, log: [], showLog: false, error: '', elapsed: 0, slow: false, result: {}, toast: '',

      init() {
        this.fields.forEach((f) => {
          let val = f.value !== undefined ? f.value : (f.type === 'check' ? false : '');
          if (f.gen) val = GEN[f.gen] ? GEN[f.gen]() : f.gen();
          this.v[f.name] = val;
          if (f.versions) this.loadVersions(f);
          if (f.options) this.opts[f.name] = f.options;
        });
        const want = new URLSearchParams(location.search).get('domain');
        const hit = want && this.domains.find((d) => d.DomainURL === want || String(d.DomainID) === want);
        if (hit) this.pickDomain(String(hit.DomainID));
        else if (cfg.preselect && this.domains.length === 1) this.pickDomain(String(this.domains[0].DomainID));
        window.addEventListener('beforeunload', (e) => {
          if (this.phase === 'running') { e.preventDefault(); e.returnValue = this.t.leave; return this.t.leave; }
        });
      },

      // location
      dom() { return this.domains.find((d) => String(d.DomainID) === String(this.domainId)) || null; },
      pickDomain(id) {
        const before = this.dom();
        this.domainId = id;
        this.errs.domain_id = '';
        this.fields.forEach((f) => {
          if (!f.auto || this.touched[f.name]) return;
          const was = before ? f.auto(before.DomainURL, this) : '';
          if (this.v[f.name] === '' || this.v[f.name] === was || !before) this.v[f.name] = f.auto(this.dom().DomainURL, this);
        });
      },
      cleanSubdir() {
        this.subdir = this.subdir.toLowerCase().replace(/\s+/g, '').replace(/^\/+/, '');
        // untouched fields that follow the location (startup file) catch up with the folder
        const d = this.dom();
        if (d) this.fields.forEach((f) => { if (f.auto && f.followsPath && !this.touched[f.name]) this.v[f.name] = f.auto(d.DomainURL, this); });
      },
      host() { const d = this.dom(); if (!d) return ''; return d.DomainURL + (this.subdir ? '/' + this.subdir.replace(/\/+$/, '') : ''); },
      installPath() { const d = this.dom(); if (!d) return ''; return (d.Docroot || '/var/www/html/' + d.DomainURL).replace(/\/+$/, '') + (this.subdir ? '/' + this.subdir.replace(/\/+$/, '') : ''); },
      siteUrl() { return this.host() ? 'https://' + this.host() : ''; },
      phpIssue() {
        const d = this.dom(), p = cfg.php;
        if (!d || !p) return '';
        const ver = d.PHPVersion && d.PHPVersion !== '/' ? d.PHPVersion : '';
        if (!ver) return p.needs ? this.t.phpNone : '';
        if (p.min && cmpVer(ver, p.min) < 0) return fmt(this.t.phpTooOld, cfg.name, p.min, ver);
        if (p.max && cmpVer(ver, p.max) > 0) return fmt(this.t.phpTooNew, cfg.name, p.max, ver);
        return '';
      },

      // fields
      fieldsFor(stepId) { return this.fields.filter((f) => (f.step || 'site') === stepId); },
      shown(f) { return !f.showWhen || f.showWhen(this.v, this); },
      off(f) { return !!(f.disabledWhen && f.disabledWhen(this.v, this)); },
      edit(f) {
        this.touched[f.name] = true;
        this.errs[f.name] = '';
        if (f.lower) this.v[f.name] = String(this.v[f.name]).toLowerCase();
        if (f.strip) this.v[f.name] = String(this.v[f.name]).replace(f.strip, '');
        if (f.check) {
          clearTimeout(this['_ck' + f.name]);
          this['_ck' + f.name] = setTimeout(() => {
            Promise.resolve(f.check(this.v[f.name], this)).then((h) => { this.hints[f.name] = h || null; });
          }, 350);
        }
      },
      blur(f) { if (f.onBlur) f.onBlur(this.v[f.name], this); },
      regen(f) { this.v[f.name] = GEN[f.gen] ? GEN[f.gen]() : f.gen(); this.reveal[f.name] = true; this.edit(f); },
      applyPreset(p) { Object.keys(p.values).forEach((k) => { this.v[k] = p.values[k]; this.errs[k] = ''; }); },
      presetOn(p) { return Object.keys(p.values).every((k) => String(this.v[k]) === String(p.values[k])); },
      loadVersions(f) {
        this.optsState[f.name] = 'loading';
        const head = f.latestOption ? [{ value: '', label: this.t.latest }] : [];
        this.opts[f.name] = head;
        Promise.resolve().then(() => f.versions(this)).then((list) => {
          list = (list || []).map((x) => (typeof x === 'string' ? { value: x, label: x } : x));
          if (list.length) list[0] = { value: list[0].value, label: list[0].label + ' (' + this.t.latest.toLowerCase() + ')' };
          this.opts[f.name] = head.concat(list);
          if (!f.latestOption && list.length && !this.v[f.name]) this.v[f.name] = list[0].value;
          this.optsState[f.name] = 'ok';
        }).catch(() => { this.optsState[f.name] = 'failed'; });
      },
      strength(f) {
        const val = this.v[f.name] || '';
        const s = score(val);
        if (!val) return { pct: 0, tone: '', label: '' };
        if (s <= 33) return { pct: 33, tone: 'var(--color-danger)', label: this.t.weak, s: s };
        if (s <= 67) return { pct: 66, tone: 'var(--color-warning)', label: this.t.medium, s: s };
        return { pct: 100, tone: 'var(--color-success)', label: this.t.strong, s: s };
      },

      // validation
      check(f) {
        if (!this.shown(f) || this.off(f) || f.type === 'check' || f.type === 'note' || f.type === 'presets') return '';
        const val = String(this.v[f.name] == null ? '' : this.v[f.name]).trim();
        if (!val) return f.required ? this.t.required : '';
        if (f.validate) {
          const r = CHECK[f.validate] ? CHECK[f.validate](val) : f.validate(val, this);
          if (r !== true && r) return this.t[r] || r;
        }
        if (f.ext && (!val.startsWith('/var/www/html/') || !val.toLowerCase().endsWith(f.ext) || val.includes('..'))) return fmt(this.t.badStartup, f.ext);
        if (f.minLen && (val.length < f.minLen || val.length > (f.maxLen || 999))) return fmt(this.t.badLen, f.minLen, f.maxLen || 999);
        if (f.type === 'number' && f.min != null && (isNaN(+val) || +val < f.min || +val > f.max)) return fmt(this.t.badNumber, f.min, f.max);
        if (f.strength && score(val) < (cfg.strength || 50)) return this.t.tooWeak;
        return '';
      },
      checkStep(i) {
        const s = this.steps[i];
        let first = null;
        if (s.id === 'where') {
          this.cleanSubdir();
          if (!this.dom()) { this.errs.domain_id = this.t.pickDomain; first = first || 'domain_id'; }
          const sd = this.subdir.replace(/\/+$/, '');
          if (sd && (!/^[a-z0-9._-]+(\/[a-z0-9._-]+)*$/.test(sd) || sd.split('/').some((p) => p === '..' || p === '.'))) { this.errs.subdirectory = this.t.badSubdir; first = first || 'subdirectory'; }
          if (cfg.php && cfg.php.block && this.phpIssue()) first = first || 'domain_id';
        }
        this.fieldsFor(s.id).forEach((f) => {
          const e = this.check(f);
          this.errs[f.name] = e;
          if (e && !first) first = f.name;
        });
        if (first) {
          this.$nextTick(() => { const el = document.getElementById(first); if (el) el.focus(); });
          return false;
        }
        return true;
      },
      next() { if (this.checkStep(this.step)) { this.step++; this.top(); } },
      back() { if (this.step > 0) { this.step--; this.top(); } },
      go(i) {
        if (i === this.step) return;
        if (i < this.step) { this.step = i; this.top(); return; }
        for (let k = this.step; k < i; k++) { if (!this.checkStep(k)) { this.step = k; return; } }
        this.step = i;
        this.top();
      },
      top() { this.$nextTick(() => { const el = this.$root.querySelector('.op-iw-card'); if (el && el.getBoundingClientRect().top < 0) el.scrollIntoView({ block: 'start' }); }); },

      // review
      reviewRows() {
        const rows = [];
        const d = this.dom();
        if (d) {
          rows.push({ group: this.t.location, label: this.t.url, value: this.siteUrl() });
          rows.push({ group: this.t.location, label: this.t.path, value: this.installPath(), mono: true });
        }
        this.steps.forEach((s) => {
          this.fieldsFor(s.id).forEach((f) => {
            if (f.review === false || !this.shown(f) || this.off(f) || f.type === 'note' || f.type === 'presets') return;
            let val = this.v[f.name];
            if (f.type === 'check') val = val ? (f.on || '✓') : (f.offLabel || '–');
            else if (f.type === 'select' && this.opts[f.name]) { const o = this.opts[f.name].find((x) => String(x.value) === String(val)); val = o ? o.label : val; }
            if (val === '' || val == null) val = f.empty || '–';
            if (f.prefix && cfg.mysqlPrefix && this.v[f.name]) val = cfg.mysqlPrefix + val;
            if (f.unit && this.v[f.name] !== '') val = val + ' ' + f.unit;
            rows.push({ group: s.label, name: f.name, label: f.review || f.label, value: String(val), secret: f.type === 'password' && !!this.v[f.name], mono: f.mono });
          });
        });
        return rows;
      },
      summary() {
        const rows = [];
        const d = this.dom();
        rows.push({ label: this.t.domain, value: d ? d.DomainURL : '–' });
        rows.push({ label: this.t.folder, value: this.subdir ? '/' + this.subdir.replace(/\/+$/, '') : this.t.root });
        (cfg.summary || []).forEach((name) => {
          const f = this.fields.find((x) => x.name === name);
          if (!f || !this.shown(f)) return;
          let val = this.v[name];
          if (f.type === 'select' && this.opts[name]) { const o = this.opts[name].find((x) => String(x.value) === String(val)); val = o ? o.label : val; }
          if (f.prefix && cfg.mysqlPrefix && val) val = cfg.mysqlPrefix + val;
          rows.push({ label: f.review || f.label, value: val === '' || val == null ? '–' : String(val) });
        });
        return rows;
      },

      // run
      submit() {
        for (let i = 0; i < this.steps.length; i++) {
          if (!this.checkStep(i)) { this.step = i; this.top(); return; }
        }
        this.start();
      },
      start() {
        const fd = new FormData(this.$refs.form);
        fd.set('domain_id', this.domainId);
        fd.set('subdirectory', this.subdir.replace(/\/+$/, ''));
        this.phase = 'running';
        this.runIdx = -1;
        this.log = [];
        this.error = '';
        this.result = {};
        this.slow = false;
        this.elapsed = 0;
        const t0 = Date.now();
        clearInterval(this._tick);
        this._tick = setInterval(() => { this.elapsed = Math.round((Date.now() - t0) / 1000); if (this.elapsed > 300) this.slow = true; }, 1000);
        window.scrollTo({ top: 0 });
        const send = cfg.transport || ndjson;
        send(cfg.endpoint, fd, (ev) => this.event(ev), this).then(() => {
          if (this.phase === 'running') this.fail(this.t.stopped);
        }).catch(() => { if (this.phase === 'running') this.fail(this.t.connection); });
      },
      event(ev) {
        if (this.phase !== 'running') return;
        if (ev.status) {
          this.log.push(ev.status);
          this.$nextTick(() => { const el = this.$refs.log; if (el) el.scrollTop = el.scrollHeight; });
          // first checkpoint whose trigger matches, the run only moves forward
          const hit = this.run.findIndex((r) => r.triggers.some((tr) => ev.status.includes(tr)));
          if (hit > this.runIdx) this.runIdx = hit;
          if ((cfg.success || []).some((s) => ev.status.includes(s))) this.finish(ev);
        } else if (ev.error) {
          this.log.push(ev.error);
          this.fail(ev.error);
        }
      },
      finish(ev) {
        clearInterval(this._tick);
        this.runIdx = this.run.length;
        this.result = { status: ev.status, user: ev.admin_user, pass: ev.admin_password };
        this.phase = 'done';
        if (window.showToast) window.showToast(ev.status, 'success');
      },
      fail(msg) {
        clearInterval(this._tick);
        this.error = msg;
        this.failIdx = Math.max(this.runIdx, 0);
        this.phase = 'failed';
      },
      stepState(i) {
        if (this.phase === 'done' || i < this.runIdx) return 'done';
        if (this.phase === 'failed' && i === this.failIdx) return 'error';
        if (i === this.runIdx || (this.runIdx < 0 && i === 0 && this.phase === 'running')) return 'active';
        return 'pending';
      },
      pct() {
        if (this.phase === 'done') return 100;
        return Math.round(Math.max(0, this.runIdx) / Math.max(1, this.run.length) * 100);
      },
      clock() { const s = this.elapsed; return s < 60 ? fmt(this.t.secs, s) : fmt(this.t.mins, Math.floor(s / 60), s % 60); },
      edit2() { this.phase = 'form'; this.step = this.steps.length - 1; },

      // done screen
      creds() {
        const out = [];
        (cfg.creds || []).forEach((c) => {
          let val = c.from === 'user' ? (this.result.user || this.v[c.field]) : c.from === 'pass' ? (this.result.pass || this.v[c.field]) : this.v[c.field];
          if (c.prefix && cfg.mysqlPrefix && val) val = cfg.mysqlPrefix + val;
          if (val) out.push({ label: c.label, value: String(val), secret: !!c.secret, key: c.field || c.from });
        });
        return out;
      },
      adminUrl() { return cfg.adminPath && this.siteUrl() ? this.siteUrl() + cfg.adminPath : ''; },
      managerUrl() { return typeof cfg.manager === 'function' ? cfg.manager(this) : cfg.manager; },
      copy(text) {
        const ok = () => this.notify(this.t.copied);
        if (window.copyToClipboard) window.copyToClipboard(text).then(ok);
        else if (navigator.clipboard) navigator.clipboard.writeText(text).then(ok);
      },
      notify(msg) {
        if (window.showToast) { window.showToast(msg, 'success'); return; }
        this.toast = msg;
        clearTimeout(this._toast);
        this._toast = setTimeout(() => { this.toast = ''; }, 1800);
      },
      fmt: fmt
    };
  };
})();
