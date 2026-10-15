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
    try { sessionStorage.setItem(STORE, JSON.stringify(messages.slice(-40))); } catch (e) {}
  }

  function escapeHTML(s) {
    return s.replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
  }

  // small markdown subset, everything is escaped first so the model can't inject html
  function inline(s) {
    return s
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
      .replace(/(^|[^*])\*([^*\s][^*]*)\*/g, '$1<em>$2</em>')
      .replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+|\/[^\s)]*)\)/g, (m, text, url) => {
        const ext = /^https?:/.test(url);
        return '<a href="' + url + '"' + (ext ? ' target="_blank" rel="noopener noreferrer"' : '') + '>' + text + '</a>';
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
      ['s1', 's2', 's3', 's4'].forEach(k => {
        const b = document.createElement('button');
        b.type = 'button';
        b.textContent = S[k];
        b.addEventListener('click', () => ask(S[k]));
        wrap.appendChild(b);
      });
      list.appendChild(wrap);
    }
    messages.forEach(m => bubble(m.role, m.content));
    list.scrollTop = list.scrollHeight;
  }

  function showUsage() {
    if (!info) return;
    about.querySelector('[data-ai-usage]').textContent = info.limit > 0 ? info.used + ' / ' + info.limit : info.used + ' (' + S.unlimited + ')';
  }

  function loadInfo() {
    return fetch('/ai/info', { headers: { Accept: 'application/json' } })
      .then(r => r.ok ? r.json() : null)
      .then(d => {
        if (!d) return;
        info = d;
        if (d.name) document.getElementById('ai-assistant-title').textContent = d.name;
        showUsage();
        render();
      })
      .catch(() => {});
  }

  function open() {
    drawer.hidden = false;
    requestAnimationFrame(() => drawer.classList.add('ai-open'));
    btn.setAttribute('aria-expanded', 'true');
    try { sessionStorage.setItem(OPEN, '1'); } catch (e) {}
    if (!info) loadInfo();
    render();
    setTimeout(() => input.focus(), 50);
  }

  function close() {
    drawer.classList.remove('ai-open');
    btn.setAttribute('aria-expanded', 'false');
    try { sessionStorage.removeItem(OPEN); } catch (e) {}
    setTimeout(() => { if (!drawer.classList.contains('ai-open')) drawer.hidden = true; }, 200);
  }

  async function ask(question) {
    question = (question || '').trim();
    if (!question || busy) return;
    busy = true;
    send.disabled = true;
    input.value = '';
    messages.push({ role: 'user', content: question });
    save();
    render();

    const el = bubble('assistant', '');
    el.innerHTML = '<span class="ai-thinking">' + escapeHTML(S.thinking) + '</span>';
    list.scrollTop = list.scrollHeight;

    let answer = '';
    let failed = '';
    try {
      const res = await fetch('/ai/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-CSRF-TOKEN': csrf_token },
        body: JSON.stringify({ messages: messages, page: { path: location.pathname, title: document.title.split(' - ')[0] } })
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
              answer += ev.text;
              el.innerHTML = markdown(answer);
              list.scrollTop = list.scrollHeight;
            }
            if (ev.error) failed = ev.error;
            if (ev.done && info) { info.used = ev.used; info.limit = ev.limit; showUsage(); }
          }
        }
      }
    } catch (e) {
      failed = S.error;
    }

    if (answer) {
      messages.push({ role: 'assistant', content: answer });
      save();
    }
    if (failed) {
      if (!answer) el.remove();
      bubble('error', failed);
      // drop the unanswered question so it doesn't get sent twice
      if (!answer) { messages.pop(); save(); input.value = question; }
    }
    busy = false;
    send.disabled = false;
    list.scrollTop = list.scrollHeight;
    input.focus();
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
