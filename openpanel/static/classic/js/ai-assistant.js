// AI Assistant drawer, the chat lives in sessionStorage so it follows the user across pages
(function () {
  const btn = document.getElementById('ai-assistant-btn');
  const drawer = document.getElementById('ai-assistant-drawer');
  if (!btn || !drawer) return;
  // the sticky header is its own stacking context, so the drawer goes straight under body
  document.body.appendChild(drawer);

  const list = document.getElementById('ai-assistant-messages');
  const form = document.getElementById('ai-assistant-form');
  const input = document.getElementById('ai-assistant-input');
  const send = document.getElementById('ai-assistant-send');
  const about = document.getElementById('ai-assistant-about');
  const S = JSON.parse(drawer.dataset.strings || '{}');
  const STORE = 'ai_assistant_chat';
  const OPEN = 'ai_assistant_open';

  let messages = load();
  let busy = false;
  let info = null;

  input.placeholder = S.placeholder || '';
  about.querySelector('[data-ai-usage-label]').textContent = S.usage || '';

  function load() {
    try { return JSON.parse(sessionStorage.getItem(STORE)) || []; } catch (e) { return []; }
  }
  function save() {
    try { sessionStorage.setItem(STORE, JSON.stringify(messages.slice(-60))); } catch (e) {}
  }

  function escapeHTML(s) {
    return String(s).replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
  }

  // small markdown subset, everything is escaped first so the model can't inject html, only panel links become clickable
  function inline(s) {
    return s
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
      .replace(/(^|[^*])\*([^*\s][^*]*)\*/g, '$1<em>$2</em>')
      .replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+|\/[^\s)]*)\)/g, (m, text, url) => {
        if (/^\/(?!\/)/.test(url)) return '<a href="' + url + '">' + text + '</a>';
        return text + ' (' + url + ')';
      });
  }

  function markdown(src) {
    const parts = escapeHTML(src).split(/```[^\n]*\n?/);
    let html = '';
    parts.forEach((part, i) => {
      if (i % 2 === 1) { html += '<pre><code>' + part.replace(/\n$/, '') + '</code></pre>'; return; }
      let listType = null;
      let para = [];
      const flushPara = () => { if (para.length) { html += '<p>' + inline(para.join('<br>')) + '</p>'; para = []; } };
      const closeList = () => { if (listType) { html += '</' + listType + '>'; listType = null; } };
      part.split('\n').forEach(line => {
        const ul = line.match(/^\s*[-*]\s+(.*)/);
        const ol = line.match(/^\s*\d+[.)]\s+(.*)/);
        const h = line.match(/^#{1,6}\s+(.*)/);
        if (ul || ol) {
          flushPara();
          const type = ul ? 'ul' : 'ol';
          if (listType !== type) { closeList(); html += '<' + type + '>'; listType = type; }
          html += '<li>' + inline((ul || ol)[1]) + '</li>';
        } else if (h) {
          flushPara(); closeList();
          html += '<h4>' + inline(h[1]) + '</h4>';
        } else if (line.trim() === '') {
          flushPara(); closeList();
        } else {
          closeList();
          para.push(line);
        }
      });
      flushPara(); closeList();
    });
    return html;
  }

  function bubble(role, text) {
    const el = document.createElement('div');
    el.className = 'ai-msg ai-msg-' + role;
    if (role === 'assistant') el.innerHTML = markdown(text);
    else el.textContent = text;
    list.appendChild(el);
    return el;
  }

  function actionLine(m) {
    const el = document.createElement('div');
    const state = m.declined ? 'declined' : (m.ok === false ? 'failed' : 'ok');
    el.className = 'ai-action ai-action-' + state;
    el.textContent = (state === 'ok' ? '✓ ' : '✕ ') + m.content;
    list.appendChild(el);
  }

  // the card is built only from the confirm event the server sends, never from model text
  function card(m) {
    const c = m.card;
    const el = document.createElement('div');
    el.className = 'ai-card ai-card-' + c.kind;
    let html = '<div class="ai-card-head"><span class="ai-card-kind">' + escapeHTML(c.kind === 'destructive' ? S.destructive : S.change) + '</span>' +
      '<strong>' + escapeHTML(c.title) + '</strong></div>' +
      '<code class="ai-card-req">' + escapeHTML(c.request) + '</code>';
    if (c.fields && c.fields.length) {
      html += '<dl class="ai-card-fields">' + c.fields.map(f => '<dt>' + escapeHTML(f[0]) + '</dt><dd>' + escapeHTML(f[1]) + '</dd>').join('') + '</dl>';
    }
    if (c.kind === 'destructive') html += '<p class="ai-card-note">' + escapeHTML(S.destructiveNote) + '</p>';
    if (c.tainted) html += '<p class="ai-card-note">' + escapeHTML(S.taintedNote) + '</p>';
    if (m.state === 'pending') {
      html += '<div class="ai-card-actions"><button type="button" data-d="allow" class="ai-btn-primary">' + escapeHTML(S.approve) + '</button>' +
        (c.remember ? '<button type="button" data-d="always">' + escapeHTML(S.always) + '</button>' : '') +
        '<button type="button" data-d="deny">' + escapeHTML(S.deny) + '</button></div>';
    } else {
      html += '<p class="ai-card-state">' + escapeHTML(S['state_' + m.state] || '') + '</p>';
    }
    el.innerHTML = html;
    el.querySelectorAll('[data-d]').forEach(b => b.addEventListener('click', () => decide(m, b.dataset.d)));
    list.appendChild(el);
  }

  function render() {
    list.innerHTML = '';
    if (info && !info.configured) {
      bubble('error', S.notConfigured);
      return;
    }
    if (!messages.length) {
      bubble('assistant', S.greeting);
      const wrap = document.createElement('div');
      wrap.className = 'ai-suggestions';
      const keys = info && info.actions === 'all' ? ['s1', 's5', 's6', 's4'] : ['s1', 's2', 's3', 's4'];
      keys.forEach(k => {
        const b = document.createElement('button');
        b.type = 'button';
        b.textContent = S[k];
        b.addEventListener('click', () => ask(S[k]));
        wrap.appendChild(b);
      });
      list.appendChild(wrap);
    }
    messages.forEach(m => {
      if (m.role === 'action') actionLine(m);
      else if (m.role === 'confirm') card(m);
      else bubble(m.role, m.content);
    });
    list.scrollTop = list.scrollHeight;
  }

  function showInfo() {
    if (!info) return;
    about.querySelector('[data-ai-usage]').textContent = info.limit > 0 ? info.used + ' / ' + info.limit : info.used + ' (' + S.unlimited + ')';
    about.querySelector('[data-ai-mode]').textContent = S['mode_' + info.actions] || '';
    const box = about.querySelector('[data-ai-allowed]');
    box.innerHTML = '';
    if (info.actions !== 'all') return;
    const h = document.createElement('p');
    h.className = 'ai-allowed-title';
    h.textContent = S.allowedTitle;
    box.appendChild(h);
    if (!info.allowed || !info.allowed.length) {
      const p = document.createElement('p');
      p.textContent = S.allowedNone;
      box.appendChild(p);
      return;
    }
    info.allowed.forEach(item => {
      const row = document.createElement('div');
      row.className = 'ai-allowed-row';
      const t = document.createElement('span');
      t.textContent = item.title;
      const rm = document.createElement('button');
      rm.type = 'button';
      rm.textContent = S.forget;
      rm.addEventListener('click', () => {
        fetch('/ai/permissions/forget', { method: 'POST', headers: { 'Content-Type': 'application/json', 'X-CSRF-TOKEN': csrf_token }, body: JSON.stringify({ action: item.action }) })
          .then(r => { if (r.ok) { info.allowed = info.allowed.filter(a => a.action !== item.action); showInfo(); } });
      });
      row.append(t, rm);
      box.appendChild(row);
    });
  }

  function loadInfo() {
    return fetch('/ai/info', { headers: { Accept: 'application/json' } })
      .then(r => r.ok ? r.json() : null)
      .then(d => {
        if (!d) return;
        info = d;
        if (d.name) document.getElementById('ai-assistant-title').textContent = d.name;
        showInfo();
        if (!busy) render();
      })
      .catch(() => {});
  }

  function open() {
    drawer.hidden = false;
    requestAnimationFrame(() => drawer.classList.add('ai-open'));
    btn.setAttribute('aria-expanded', 'true');
    try { sessionStorage.setItem(OPEN, '1'); } catch (e) {}
    loadInfo();
    render();
    setTimeout(() => input.focus(), 50);
  }

  function close() {
    drawer.classList.remove('ai-open');
    btn.setAttribute('aria-expanded', 'false');
    try { sessionStorage.removeItem(OPEN); } catch (e) {}
    setTimeout(() => { if (!drawer.classList.contains('ai-open')) drawer.hidden = true; }, 200);
  }

  function setBusy(b) {
    busy = b;
    send.disabled = b;
  }

  // stream reads the ndjson answer: text, action lines, a confirmation card or an error
  async function stream(url, payload) {
    let el = null;
    let answer = '';
    let failed = '';
    const thinking = bubble('assistant', '');
    thinking.innerHTML = '<span class="ai-thinking">' + escapeHTML(S.thinking) + '</span>';
    list.scrollTop = list.scrollHeight;

    const flushAnswer = () => {
      if (answer.trim()) { messages.push({ role: 'assistant', content: answer }); save(); }
      answer = '';
      el = null;
    };

    try {
      const res = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-CSRF-TOKEN': csrf_token },
        body: JSON.stringify(payload)
      });
      if (!res.ok || !res.body) {
        const d = await res.json().catch(() => ({}));
        failed = d.error || S.error;
      } else {
        const reader = res.body.getReader();
        const dec = new TextDecoder();
        let buf = '';
        for (;;) {
          const { value, done } = await reader.read();
          if (done) break;
          buf += dec.decode(value, { stream: true });
          let nl;
          while ((nl = buf.indexOf('\n')) >= 0) {
            const line = buf.slice(0, nl).trim();
            buf = buf.slice(nl + 1);
            if (!line) continue;
            let ev;
            try { ev = JSON.parse(line); } catch (e) { continue; }
            if (ev.text) {
              thinking.remove();
              if (!el) el = bubble('assistant', '');
              answer += ev.text;
              el.innerHTML = markdown(answer);
            }
            if (ev.step && answer) answer += '\n\n';
            if (ev.action) {
              flushAnswer();
              const m = { role: 'action', content: ev.action.note, ok: ev.action.ok, declined: !!ev.action.declined };
              messages.push(m); save();
              render(); list.appendChild(thinking);
            }
            if (ev.confirm) {
              flushAnswer();
              messages.push({ role: 'confirm', card: ev.confirm, state: 'pending' }); save();
              thinking.remove(); render();
            }
            if (ev.error) failed = ev.error;
            if (ev.done && info) { info.used = ev.used; info.limit = ev.limit; showInfo(); }
            list.scrollTop = list.scrollHeight;
          }
        }
      }
    } catch (e) {
      failed = S.error;
    }
    thinking.remove();
    flushAnswer();
    return failed;
  }

  async function ask(question) {
    question = (question || '').trim();
    if (!question || busy) return;
    setBusy(true);
    input.value = '';
    // a card left unanswered is dropped on the server once a new question comes in
    messages.forEach(m => { if (m.role === 'confirm' && m.state === 'pending') m.state = 'skipped'; });
    messages.push({ role: 'user', content: question });
    const sent = messages.length;
    save();
    render();

    const history = messages.filter(m => m.role === 'user' || m.role === 'assistant' || m.role === 'action').map(m => ({ role: m.role, content: m.content }));
    const failed = await stream('/ai/chat', { messages: history, page: { path: location.pathname, title: document.title.split(' - ')[0] } });
    if (failed) {
      // drop the unanswered question so it doesn't get sent twice
      if (messages.length === sent) { messages.pop(); input.value = question; }
      save();
      render();
      bubble('error', failed);
    } else {
      render();
    }
    setBusy(false);
    list.scrollTop = list.scrollHeight;
    input.focus();
  }

  async function decide(m, decision) {
    if (busy || m.state !== 'pending') return;
    setBusy(true);
    m.state = decision === 'deny' ? 'declined' : 'approved';
    save();
    render();
    const failed = await stream('/ai/confirm', { id: m.card.id, decision: decision });
    if (failed) {
      if (/expired/i.test(failed)) { m.state = 'expired'; save(); }
      render();
      bubble('error', failed);
    } else {
      render();
    }
    if (decision === 'always') loadInfo();
    setBusy(false);
    list.scrollTop = list.scrollHeight;
  }

  btn.addEventListener('click', () => drawer.classList.contains('ai-open') ? close() : open());
  document.getElementById('ai-assistant-close').addEventListener('click', close);
  document.getElementById('ai-assistant-info').addEventListener('click', () => { about.hidden = !about.hidden; });
  document.getElementById('ai-assistant-clear').addEventListener('click', () => { if (busy) return; messages = []; save(); render(); input.focus(); });
  form.addEventListener('submit', e => { e.preventDefault(); ask(input.value); });
  input.addEventListener('keydown', e => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); ask(input.value); }
  });
  input.addEventListener('input', () => { input.style.height = 'auto'; input.style.height = Math.min(input.scrollHeight, 160) + 'px'; });
  document.addEventListener('keydown', e => { if (e.key === 'Escape' && drawer.classList.contains('ai-open')) close(); });

  // keep it open when the user navigates to another page while chatting
  try { if (sessionStorage.getItem(OPEN)) open(); } catch (e) {}
})();
