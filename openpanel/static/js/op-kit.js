// small Alpine helpers for pages on the op- kit
// row menu that opens fixed next to its button, so the table's scroll box can't clip it
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

// sidebar menu search: open every group while there's a query, hide links that don't match, and section labels left with nothing under them
window.opNavFilter = function (q) {
    q = (q || '').trim().toLowerCase();
    const nav = document.querySelector('.op-sidebar-nav');
    if (!nav) return;
    window.dispatchEvent(new CustomEvent('sidebar-toggle-all', { detail: { open: !!q } }));
    const has = (el) => !q || el.textContent.toLowerCase().includes(q);
    let any = false;
    nav.querySelectorAll('.op-nav > li:not(.op-nav-label)').forEach((li) => {
        const group = li.querySelector(':scope > .op-nav-group');
        if (group) {
            const groupHit = has(group);
            let shown = 0;
            li.querySelectorAll('.op-nav-sub').forEach((a) => {
                const hit = groupHit || has(a);
                a.parentElement.hidden = !hit;
                if (hit) shown++;
            });
            li.hidden = !groupHit && !shown;
        } else {
            const a = li.querySelector('.op-nav-item');
            if (a) li.hidden = !has(a);
        }
        if (!li.hidden) any = true;
    });
    nav.querySelectorAll('#favorites-list a').forEach((a) => {
        a.hidden = !has(a);
        if (!a.hidden && q) any = true;
    });
    nav.querySelectorAll('.op-nav > li.op-nav-label').forEach((label) => {
        let next = label.nextElementSibling, visible = false;
        while (next && !next.classList.contains('op-nav-label')) { if (!next.hidden) visible = true; next = next.nextElementSibling; }
        label.hidden = !visible;
    });
    const empty = document.getElementById('op-nav-empty');
    if (empty) empty.hidden = !q || any;
};

// Enter in the menu search opens the first link still showing
window.opNavFirst = function () {
    const link = Array.from(document.querySelectorAll('.op-sidebar-nav a.op-nav-item, .op-sidebar-nav a.op-nav-sub')).find((a) => a.offsetParent && !a.classList.contains('is-locked'));
    if (link) window.location.href = link.getAttribute('href');
};

// copy buttons: data-op-copy-from="<id>" copies that element's text, data-op-copy="<text>" copies the text itself
document.addEventListener('click', function (e) {
    const b = e.target.closest && e.target.closest('[data-op-copy], [data-op-copy-from]');
    if (!b || !window.copyToClipboard) return;
    const src = b.dataset.opCopyFrom ? document.getElementById(b.dataset.opCopyFrom) : null;
    const text = src ? (src.firstChild && src.firstChild.nodeType === 3 ? src.firstChild.textContent : src.textContent) : b.dataset.opCopy;
    if (!text || !text.trim()) return;
    e.preventDefault();
    window.copyToClipboard(text.trim()).then(function () { if (window.showToast) showToast(b.dataset.opCopied || 'Copied to clipboard', 'success'); });
});

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

// sortable headers: one button per column, the first click sorts ascending and the next flips it; aria-sort on the th draws the arrow, fn is the page's sorter
window.opSortToggle = function (event, btn, fn) {
    const th = btn.closest('th'), table = btn.closest('table');
    const next = th && th.getAttribute('aria-sort') === 'ascending' ? 'desc' : 'asc';
    btn.dataset.direction = next;
    if (table) table.querySelectorAll('.op-th-sort[data-sort-key]').forEach((b) => { const h = b.closest('th'); if (h) h.setAttribute('aria-sort', b === btn ? (next === 'asc' ? 'ascending' : 'descending') : 'none'); });
    (fn || window.sortTableColumn)(event, btn);
};
