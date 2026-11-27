// File Manager editor: starts Monaco, Ace, CodeMirror or the plain textarea and adds the toolbar around it
// (undo/redo, text size, syntax mode with auto-detection, encoding + path with Reopen, browser drafts, save over fetch).
// The page provides the markup (data-fe-* controls) and calls opFileEditor(cfg), see templates/files/edit_file.html.
(function () {
    'use strict';

    const store = {
        get(k) { try { return localStorage.getItem('opFileEditor.' + k); } catch (e) { return null; } },
        set(k, v) { try { localStorage.setItem('opFileEditor.' + k, v); } catch (e) {} },
    };

    // ---------- encodings (names match internal/modules/filemanager/encoding.go) ----------
    const ENCODINGS = [
        ['utf-8', 'UTF-8'],
        ['iso-8859-1', 'ISO-8859-1 (Western)'], ['windows-1252', 'Windows-1252 (Western)'], ['iso-8859-15', 'ISO-8859-15 (Western, €)'],
        ['iso-8859-2', 'ISO-8859-2 (Central European)'], ['windows-1250', 'Windows-1250 (Central European)'],
        ['windows-1251', 'Windows-1251 (Cyrillic)'], ['koi8-r', 'KOI8-R (Russian)'], ['koi8-u', 'KOI8-U (Ukrainian)'], ['iso-8859-5', 'ISO-8859-5 (Cyrillic)'],
        ['iso-8859-7', 'ISO-8859-7 (Greek)'], ['windows-1253', 'Windows-1253 (Greek)'],
        ['iso-8859-9', 'ISO-8859-9 (Turkish)'], ['windows-1254', 'Windows-1254 (Turkish)'],
        ['windows-1256', 'Windows-1256 (Arabic)'],
        ['shift_jis', 'Shift_JIS (Japanese)'], ['euc-jp', 'EUC-JP (Japanese)'], ['euc-kr', 'EUC-KR (Korean)'],
        ['gbk', 'GBK (Simplified Chinese)'], ['gb18030', 'GB18030 (Chinese)'], ['big5', 'Big5 (Traditional Chinese)'],
        ['utf-16le', 'UTF-16 LE'], ['utf-16be', 'UTF-16 BE'],
    ];

    const FONT_SIZES = [11, 12, 13, 14, 15, 16, 18, 20, 22, 24];

    // ---------- syntax modes ----------
    // content sniffing for files whose name says nothing (no extension, error_log, a script with a shebang), result uses Ace's mode names
    function sniffMode(text) {
        const head = text.slice(0, 4000);
        const first = head.split('\n', 1)[0];
        const bang = /^#!\s*(?:\S*\/)?(?:env\s+)?(\S+)/.exec(first);
        if (bang) {
            const b = bang[1];
            if (/^(ba|z|k|da)?sh$/.test(b)) return 'sh';
            if (/^python/.test(b)) return 'python';
            if (/^(node|nodejs|deno|bun)$/.test(b)) return 'javascript';
            if (/^perl/.test(b)) return 'perl';
            if (/^php/.test(b)) return 'php';
            if (/^ruby/.test(b)) return 'ruby';
            if (/^lua/.test(b)) return 'lua';
        }
        const t = head.replace(/^﻿/, '').trimStart();
        if (/^<\?php/i.test(t)) return 'php';
        if (/^<\?xml/i.test(t)) return /<svg[\s>]/i.test(head) ? 'svg' : 'xml';
        if (/^<!doctype html|^<html[\s>]/i.test(t)) return 'html';
        if (/^<svg[\s>]/i.test(t)) return 'svg';
        if (/^[[{]/.test(t) && text.length < 2e6) { try { JSON.parse(text); return 'json'; } catch (e) {} }
        if (/^---\s*$/m.test(first) || /^[\w.-]+:\s+\S.*\n[\w.-]+:\s/m.test(t) && !/[;{}]\s*$/m.test(t)) return 'yaml';
        if (/<VirtualHost|^\s*Rewrite(Engine|Rule|Cond)\b|^\s*<IfModule|^\s*(Order|Allow|Deny|Require)\s+\w/m.test(head)) return 'apache_conf';
        if (/^\s*server\s*\{|^\s*location\s+[~=^/@]|^\s*(listen|server_name|proxy_pass)\s+\S+;/m.test(head)) return 'nginx';
        if (/^\s*\[[^\]\n]+\]\s*$/m.test(head) && /^\s*[\w.-]+\s*=/m.test(head)) return 'ini';
        if (/^\s*(CREATE|INSERT\s+INTO|SELECT|DROP|ALTER\s+TABLE|UPDATE\s+\w+\s+SET)\b/im.test(t)) return 'sql';
        if (/^FROM\s+\S+/m.test(t) && /^(RUN|COPY|CMD|ENTRYPOINT|WORKDIR)\b/m.test(head)) return 'dockerfile';
        if (/^#{1,6}\s+\S/m.test(head) && /(\n|^)(\s*[-*]\s|\[.+\]\(.+\)|```)/.test(head)) return 'markdown';
        return '';
    }

    // Ace mode name -> Monaco language id where the name differs
    const ACE_TO_MONACO = {
        sh: 'shell', c_cpp: 'cpp', text: 'plaintext', golang: 'go', csharp: 'csharp', objectivec: 'objective-c',
        batchfile: 'bat', svg: 'xml', nginx: 'plaintext', apache_conf: 'plaintext', jsx: 'javascript', tsx: 'typescript',
        rst: 'restructuredtext', sass: 'scss', pgsql: 'pgsql', mysql: 'mysql', makefile: 'plaintext',
    };
    // CodeMirror only has the modes the page loads
    const CM_MODES = [
        ['null', 'Plain text'], ['text/html', 'HTML'], ['application/xml', 'XML'], ['text/css', 'CSS'],
        ['text/javascript', 'JavaScript'], ['application/json', 'JSON'], ['application/x-httpd-php', 'PHP'],
        ['text/x-python', 'Python'], ['text/x-sh', 'Shell'], ['text/x-sql', 'SQL'], ['text/x-yaml', 'YAML'],
        ['text/x-markdown', 'Markdown'], ['text/x-csrc', 'C'], ['text/x-c++src', 'C++'], ['text/x-java', 'Java'], ['text/x-csharp', 'C#'],
    ];
    const ACE_TO_CM = {
        text: 'null', html: 'text/html', xml: 'application/xml', svg: 'application/xml', css: 'text/css', javascript: 'text/javascript',
        json: 'application/json', php: 'application/x-httpd-php', python: 'text/x-python', sh: 'text/x-sh', sql: 'text/x-sql',
        mysql: 'text/x-sql', pgsql: 'text/x-sql', yaml: 'text/x-yaml', markdown: 'text/x-markdown', c_cpp: 'text/x-c++src',
        java: 'text/x-java', csharp: 'text/x-csharp',
    };
    const CM_EXT = {
        html: 'text/html', htm: 'text/html', html5: 'text/html', xml: 'application/xml', svg: 'application/xml', css: 'text/css',
        js: 'text/javascript', mjs: 'text/javascript', cjs: 'text/javascript', json: 'application/json',
        php: 'application/x-httpd-php', php5: 'application/x-httpd-php', php7: 'application/x-httpd-php', php8: 'application/x-httpd-php', phtml: 'application/x-httpd-php',
        py: 'text/x-python', sh: 'text/x-sh', bash: 'text/x-sh', zsh: 'text/x-sh', sql: 'text/x-sql', yaml: 'text/x-yaml', yml: 'text/x-yaml',
        md: 'text/x-markdown', markdown: 'text/x-markdown', c: 'text/x-csrc', h: 'text/x-csrc', cpp: 'text/x-c++src', cc: 'text/x-c++src', hpp: 'text/x-c++src',
        java: 'text/x-java', cs: 'text/x-csharp',
    };

    function baseName(p) { return p.split('/').pop(); }
    function ext(p) { const b = baseName(p).toLowerCase(); const i = b.lastIndexOf('.'); return i > 0 ? b.slice(i + 1) : (i === 0 ? b.slice(1) : ''); }

    const confirmBox = (msg, opts) => (window.opConfirm ? window.opConfirm(msg, opts) : Promise.resolve(window.confirm(msg)));

    // ---------- the editor ----------
    window.opFileEditor = function (cfg) {
        const root = cfg.root;
        const $ = (s) => root.querySelector(s);
        const i18n = cfg.i18n || {};
        const tr = (k, fallback) => i18n[k] || fallback;
        const isDark = document.documentElement.classList.contains('dark');
        const backend = cfg.backend;
        const params = new URLSearchParams(location.search);
        const encoding = (params.get('encoding') || 'utf-8').toLowerCase();

        let ed = null; // the backend adapter, set once the editor exists
        let saved = cfg.content;
        const changeCbs = [];
        const onChange = (cb) => changeCbs.push(cb);
        const fireChange = () => changeCbs.forEach((cb) => cb());

        // ----- hash line (#42) -----
        const hashLine = () => { const n = parseInt((location.hash || '').slice(1), 10); return isNaN(n) ? null : n; };
        const setHashLine = (n) => { const h = '#' + n; if (location.hash !== h) history.replaceState(null, '', h); };

        // ----- backends: each gives the same small interface -----
        function initMonaco(done) {
            require.config({ paths: { vs: cfg.urls.monacoVs } });
            require(['vs/editor/editor.main'], function () {
                const langs = monaco.languages.getLanguages();
                const byId = new Set(langs.map((l) => l.id));
                const e = monaco.editor.create(document.getElementById('editor'), {
                    value: cfg.content, language: 'plaintext',
                    theme: isDark ? 'vs-dark' : 'vs', automaticLayout: true, lineNumbers: 'on', fontSize: 14,
                });
                const model = e.getModel();
                const a = {
                    kind: 'monaco', raw: e,
                    getValue: () => e.getValue(),
                    setValue: (v) => { model.pushEditOperations([], [{ range: model.getFullModelRange(), text: v }], () => null); },
                    undo: () => { e.focus(); e.trigger('toolbar', 'undo', null); },
                    redo: () => { e.focus(); e.trigger('toolbar', 'redo', null); },
                    focus: () => e.focus(),
                    setFontSize: (px) => e.updateOptions({ fontSize: px }),
                    modes: () => langs.map((l) => [l.id, (l.aliases && l.aliases[0]) || l.id]).sort((x, y) => x[1].localeCompare(y[1])),
                    modeFromName(path) {
                        const b = baseName(path).toLowerCase();
                        let best = '', bestLen = 0;
                        for (const l of langs) {
                            if ((l.filenames || []).some((f) => f.toLowerCase() === b)) return l.id;
                            for (const x of l.extensions || []) {
                                const xl = x.toLowerCase();
                                if (b.endsWith(xl) && xl.length > bestLen) { best = l.id; bestLen = xl.length; }
                            }
                        }
                        return best || 'plaintext';
                    },
                    fromAce: (m) => { const id = ACE_TO_MONACO[m] || m; return byId.has(id) ? id : 'plaintext'; },
                    plain: 'plaintext',
                    getMode: () => model.getLanguageId ? model.getLanguageId() : model.getModeId(),
                    setMode: (id) => monaco.editor.setModelLanguage(model, id),
                    onChange: (cb) => e.onDidChangeModelContent(cb),
                };
                const line = hashLine();
                if (line) { e.revealLineInCenter(line); e.setPosition({ lineNumber: line, column: 1 }); e.focus(); }
                e.onDidChangeCursorPosition(() => { const p = e.getPosition(); if (p) setHashLine(p.lineNumber); });
                window._editorLayout = () => e.layout();
                done(a);
            });
        }

        function initAce(done) {
            const e = ace.edit(document.getElementById('editor-ace'));
            ace.config.set('basePath', cfg.urls.ace);
            const modelist = ace.require('ace/ext/modelist');
            e.setTheme(isDark ? 'ace/theme/monokai' : 'ace/theme/chrome');
            e.setValue(cfg.content, -1);
            e.session.getUndoManager().reset();
            e.setOptions({ fontSize: '14px', showPrintMargin: false, enableLiveAutocompletion: true });
            let mode = 'text';
            const a = {
                kind: 'ace', raw: e,
                getValue: () => e.getValue(),
                setValue: (v) => { e.session.doc.setValue(v); },
                undo: () => { e.undo(); e.focus(); },
                redo: () => { e.redo(); e.focus(); },
                focus: () => e.focus(),
                setFontSize: (px) => e.setFontSize(px + 'px'),
                modes: () => (modelist ? modelist.modes.map((m) => [m.name, m.caption]) : [['text', 'Text']]).sort((x, y) => x[1].localeCompare(y[1])),
                modeFromName: (path) => (modelist ? modelist.getModeForPath(baseName(path)).name : 'text'),
                fromAce: (m) => m,
                plain: 'text',
                getMode: () => mode,
                setMode: (m) => { mode = m; e.session.setMode('ace/mode/' + m); },
                onChange: (cb) => e.session.on('change', cb),
            };
            const line = hashLine();
            if (line) e.gotoLine(line, 0, true);
            e.selection.on('changeCursor', () => setHashLine(e.getCursorPosition().row + 1));
            window._editorLayout = () => e.resize();
            done(a);
        }

        function initCodeMirror(done) {
            const cm = CodeMirror(document.getElementById('editor-codemirror-mount'), {
                value: cfg.content, mode: null, theme: isDark ? 'dracula' : 'default',
                lineNumbers: true, matchBrackets: true, autoCloseBrackets: true,
                indentUnit: 4, tabSize: 4, indentWithTabs: false, lineWrapping: false,
                extraKeys: { Tab: (c) => c.replaceSelection('    ') },
            });
            let mode = 'null';
            const a = {
                kind: 'codemirror', raw: cm,
                getValue: () => cm.getValue(),
                setValue: (v) => cm.setValue(v),
                undo: () => { cm.undo(); cm.focus(); },
                redo: () => { cm.redo(); cm.focus(); },
                focus: () => cm.focus(),
                setFontSize: (px) => { cm.getWrapperElement().style.fontSize = px + 'px'; cm.refresh(); },
                modes: () => CM_MODES.map(([id, label]) => [id, id === 'null' ? tr('plainText', label) : label]),
                modeFromName: (path) => CM_EXT[ext(path)] || 'null',
                fromAce: (m) => ACE_TO_CM[m] || 'null',
                plain: 'null',
                getMode: () => mode,
                setMode: (m) => { mode = m; cm.setOption('mode', m === 'null' ? null : m); },
                onChange: (cb) => cm.on('change', cb),
            };
            requestAnimationFrame(() => requestAnimationFrame(() => cm.refresh()));
            const line = hashLine();
            if (line) requestAnimationFrame(() => requestAnimationFrame(() => { cm.scrollIntoView({ line: line - 1, ch: 0 }, 200); cm.setCursor({ line: line - 1, ch: 0 }); }));
            cm.on('cursorActivity', () => setHashLine(cm.getCursor().line + 1));
            window._editorLayout = () => cm.refresh();
            done(a);
        }

        function initText(done) {
            const ta = document.getElementById('editor-text');
            ta.value = cfg.content;
            // insertText keeps the browser's own undo history, assigning .value would wipe it
            ta.addEventListener('keydown', (e) => {
                if (e.key === 'Tab' && !e.ctrlKey && !e.altKey && !e.metaKey) {
                    e.preventDefault();
                    if (!document.execCommand('insertText', false, '    ')) {
                        const s = ta.selectionStart;
                        ta.setRangeText('    ', s, ta.selectionEnd, 'end');
                        fireChange();
                    }
                }
            });
            const a = {
                kind: 'text', raw: ta,
                getValue: () => ta.value,
                setValue: (v) => { ta.focus(); ta.select(); if (!document.execCommand('insertText', false, v)) ta.value = v; fireChange(); },
                undo: () => { ta.focus(); document.execCommand('undo'); },
                redo: () => { ta.focus(); document.execCommand('redo'); },
                focus: () => ta.focus(),
                setFontSize: (px) => { ta.style.fontSize = px + 'px'; },
                modes: null,
                onChange: (cb) => ta.addEventListener('input', cb),
            };
            done(a);
        }

        // ----- toolbar -----
        function fillSelect(sel, items, value) {
            sel.textContent = '';
            for (const [v, label] of items) {
                const o = document.createElement('option');
                o.value = v;
                o.textContent = label;
                sel.appendChild(o);
            }
            sel.value = value;
        }

        function setupToolbar() {
            const undoBtn = $('[data-fe-undo]'), redoBtn = $('[data-fe-redo]');
            if (undoBtn) undoBtn.addEventListener('click', () => ed.undo());
            if (redoBtn) redoBtn.addEventListener('click', () => ed.redo());

            const sizeSel = $('[data-fe-size]');
            const size = parseInt(store.get('fontSize'), 10) || 14;
            if (sizeSel) {
                fillSelect(sizeSel, FONT_SIZES.map((s) => [String(s), s + ' px']), String(FONT_SIZES.includes(size) ? size : 14));
                sizeSel.addEventListener('change', () => { const px = parseInt(sizeSel.value, 10); ed.setFontSize(px); store.set('fontSize', px); });
            }
            if (size !== 14) ed.setFontSize(size);

            const modeSel = $('[data-fe-mode]');
            if (ed.modes) {
                // the name decides first, the content only when the name says nothing
                let mode = ed.modeFromName(cfg.filePath);
                if (mode === ed.plain) { const s = sniffMode(cfg.content); if (s) mode = ed.fromAce(s); }
                ed.setMode(mode);
                if (modeSel) {
                    fillSelect(modeSel, ed.modes(), mode);
                    modeSel.addEventListener('change', () => ed.setMode(modeSel.value));
                }
            } else if (modeSel) {
                const wrap = modeSel.closest('[data-fe-mode-wrap]') || modeSel;
                wrap.hidden = true;
            }

            // encoding + path + Reopen
            const encSel = $('[data-fe-encoding]');
            if (encSel) {
                const list = ENCODINGS.slice();
                if (!list.some(([v]) => v === encoding)) list.push([encoding, encoding.toUpperCase()]);
                fillSelect(encSel, list, encoding);
            }
            const pathIn = $('[data-fe-path]');
            const reopenBtn = $('[data-fe-reopen]');
            if (pathIn && !pathIn.value) pathIn.value = cfg.homePrefix + cfg.filePath;
            // a long path shows its end (the file name) with the cut start faded, typing still shows it all
            if (pathIn) {
                const fade = 'linear-gradient(to right, transparent 0, #000 2rem)';
                const showEnd = () => {
                    if (document.activeElement === pathIn) return;
                    pathIn.scrollLeft = pathIn.scrollWidth;
                    const cut = pathIn.scrollWidth > pathIn.clientWidth + 1;
                    pathIn.style.webkitMaskImage = pathIn.style.maskImage = cut ? fade : '';
                };
                pathIn.addEventListener('focus', () => { pathIn.style.webkitMaskImage = pathIn.style.maskImage = ''; });
                pathIn.addEventListener('blur', showEnd);
                // the field settles its width after the selects and fonts, so follow its size instead of measuring once
                if (window.ResizeObserver) new ResizeObserver(showEnd).observe(pathIn);
                else window.addEventListener('resize', showEnd);
                requestAnimationFrame(showEnd);
            }
            const reopen = async () => {
                let p = (pathIn ? pathIn.value : cfg.homePrefix + cfg.filePath).trim();
                if (p.startsWith(cfg.homePrefix)) p = p.slice(cfg.homePrefix.length);
                else if (p.startsWith('/')) { showToast(tr('outsideHome', 'Only files in ' + cfg.homePrefix + ' can be opened.'), 'error'); return; }
                p = p.replace(/^\/+/, '');
                if (!p) return;
                const q = new URLSearchParams();
                q.set('editor', backend);
                if (params.get('fullscreen')) q.set('fullscreen', params.get('fullscreen'));
                const enc = encSel ? encSel.value : encoding;
                if (enc !== 'utf-8' || p === cfg.filePath) q.set('encoding', enc);
                const url = '/file-manager/edit-file/' + encodeURIComponent(p) + '?' + q.toString();
                if (isDirty() && !(await confirmBox(tr('reopenUnsaved', 'You have unsaved changes. Reopen anyway?'), { ok: tr('reopen', 'Reopen'), danger: true }))) return;
                // ask first so a typo in the path shows a message here instead of leaving the editor
                try {
                    const res = await fetch(url + '&output=json', { headers: { Accept: 'application/json' } });
                    if (!res.ok) {
                        const data = await res.json().catch(() => ({}));
                        const msg = data.error || tr('reopenFailed', 'The file could not be opened.');
                        // a missing file can be started empty, it's created on the first save (its folder has to exist)
                        const create = res.status === 404
                            ? { label: tr('createFile', 'Create file'), onClick: () => { leaving = true; location.assign(url + '&new=true'); } }
                            : false;
                        // with a button to press it stays until closed, a 5 second toast is gone before you reach it
                        showToast(msg, 'error', !!create, create);
                        return;
                    }
                } catch (err) {
                    showToast(tr('reopenFailed', 'The file could not be opened.'), 'error');
                    return;
                }
                leaving = true;
                location.assign(url);
            };
            if (reopenBtn) reopenBtn.addEventListener('click', reopen);
            if (pathIn) pathIn.addEventListener('keydown', (e) => { if (e.key === 'Enter') { e.preventDefault(); reopen(); } });

            // the editor picker links keep the encoding
            root.querySelectorAll('a[href^="?editor="]').forEach((a) => {
                const u = new URLSearchParams(a.getAttribute('href').slice(1));
                if (params.get('encoding')) u.set('encoding', params.get('encoding'));
                a.setAttribute('href', '?' + u.toString());
            });

            // opened byte for byte because the file isn't UTF-8
            const notice = $('[data-fe-notice]');
            if (params.get('encoding_guess')) {
                if (notice) notice.hidden = false;
                params.delete('encoding_guess');
                history.replaceState(null, '', location.pathname + '?' + params.toString() + location.hash);
            }
        }

        // ----- dirty state, drafts, save -----
        let leaving = false;
        const isDirty = () => ed && ed.getValue() !== saved;
        let draft = null;

        function setupSave() {
            // opened from the file manager: offer Close, and tell the file manager after a save so its list is current
            if (cfg.opener === undefined) {
                try { cfg.opener = window.opener && window.opener !== window && window.opener.location.origin === location.origin ? window.opener : null; } catch (e) { cfg.opener = null; }
            }
            const closeBtn = document.getElementById('closeWindowButton');
            if (cfg.opener && closeBtn) {
                closeBtn.hidden = false;
                closeBtn.addEventListener('click', () => window.close());
            }
            const form = root.closest('form') || document.querySelector('form');
            const saveBtn = cfg.saveButton || document.getElementById('editorcontentvalue');
            const hidden = document.getElementById('editor_content');
            const icon = saveBtn && saveBtn.querySelector('i');
            const iconClass = icon ? icon.className : '';
            let saving = false;
            document.addEventListener('keydown', (e) => {
                if ((e.ctrlKey || e.metaKey) && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 's') { e.preventDefault(); saveBtn.click(); }
            });
            form.addEventListener('submit', async (e) => {
                // a theme on an older panel can ask for the plain form post, those panels answer a save with a redirect
                if (cfg.legacySave && cfg.legacySave()) { hidden.value = ed.getValue(); leaving = true; return; }
                e.preventDefault();
                if (saving) return;
                saving = true;
                saveBtn.disabled = true;
                if (cfg.busyClass) saveBtn.classList.add(cfg.busyClass);
                if (icon) icon.className = (cfg.spinnerClass || 'ri-loader-4-line animate-spin');
                const sent = ed.getValue();
                hidden.value = sent;
                const url = new URL(location.href);
                url.hash = '';
                url.searchParams.set('output', 'json');
                try {
                    const res = await fetch(url, { method: 'POST', body: new FormData(form), headers: { Accept: 'application/json' } });
                    const data = await res.json().catch(() => ({}));
                    if (!res.ok) throw new Error(data.error || tr('saveFailed', 'Error saving file.'));
                    saved = sent;
                    if (draft) draft.saved(sent);
                    showToast(data.message || tr('saved', 'File saved successfully'), 'success');
                    if (cfg.opener) { try { cfg.opener.postMessage({ type: 'op-file-saved', path: cfg.filePath }, location.origin); } catch (err) {} }
                } catch (err) {
                    showToast(err.message || tr('saveFailed', 'Error saving file.'), 'error');
                } finally {
                    saving = false;
                    saveBtn.disabled = false;
                    if (cfg.busyClass) saveBtn.classList.remove(cfg.busyClass);
                    if (icon) icon.className = iconClass;
                }
            });
            window.addEventListener('beforeunload', (e) => {
                if (leaving || !isDirty()) return;
                e.preventDefault();
                e.returnValue = '';
            });
            document.addEventListener('click', (e) => {
                if (!isDirty()) return;
                const a = e.target.closest('a[href]');
                if (!a) return;
                const href = a.getAttribute('href');
                if (!href || href.startsWith('#') || a.closest('[x-data]')) return;
                e.preventDefault();
                confirmBox(tr('unsavedChanges', 'You have unsaved changes. Leave anyway?'), { ok: tr('leave', 'Leave'), danger: true }).then((ok) => {
                    if (!ok) return;
                    leaving = true;
                    if (a.target === '_blank') window.open(a.href, '_blank'); else location.href = a.href;
                });
            });
        }

        function setupDraft() {
            if (!window.opEditorDraft) return;
            const draftRoot = document.querySelector('[data-op-draft]');
            if (!draftRoot) return;
            draft = window.opEditorDraft({
                key: 'file:' + cfg.filePath, root: draftRoot, initial: cfg.content,
                get: () => ed.getValue(), set: (v) => ed.setValue(v), onChange: (cb) => onChange(cb),
            });
            // the draft notice sits above the editor, give the editor its space back when it closes
            if (window._editorLayout) new MutationObserver(() => window._editorLayout()).observe(draftRoot, { attributes: true, attributeFilter: ['hidden'], subtree: true });
        }

        // a theme that ships this script can also run on a panel that ignores ?encoding=, those panels don't send this header
        const hideEncoding = () => { const el = $('[data-fe-encoding]'); if (el) el.hidden = true; };
        const probe = cfg.probe
            ? fetch(location.pathname + location.search, { method: 'HEAD', cache: 'no-store' }).then((r) => r.headers.get('X-Editor-Encodings') === '1').catch(() => false)
            : Promise.resolve(true);

        const start = { monaco: initMonaco, ace: initAce, codemirror: initCodeMirror }[backend] || initText;
        start((a) => {
            ed = a;
            ed.onChange(fireChange);
            setupToolbar();
            setupSave();
            setupDraft();
            probe.then((ok) => { if (!ok) hideEncoding(); });
        });
    };
})();
