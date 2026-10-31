// shared bits for the web server conf and vhost editors: CodeMirror mode, find, go to line, pre-save check
(function () {
    if (!window.CodeMirror) return;

    // directives, <Sections>, server { } blocks, comments, quoted strings, $variables and numbers
    CodeMirror.defineMode('opconf', function () {
        return {
            startState: function () { return { first: true, inTag: false }; },
            token: function (stream, state) {
                if (stream.sol()) { state.first = true; state.inTag = false; }
                if (stream.eatSpace()) return null;
                var ch = stream.peek();
                var prev = stream.pos > 0 ? stream.string.charAt(stream.pos - 1) : ' ';
                if (ch === '#' && /\s/.test(prev)) { stream.skipToEnd(); return 'comment'; }
                if (ch === '"' || ch === "'") {
                    stream.next();
                    var esc = false, c;
                    while ((c = stream.next()) != null) {
                        if (c === ch && !esc) break;
                        esc = !esc && c === '\\';
                    }
                    state.first = false;
                    return 'string';
                }
                if (ch === '<' && stream.match(/^<\/?[A-Za-z][\w:-]*/)) { state.inTag = true; state.first = false; return 'tag'; }
                if (ch === '>' && state.inTag) { stream.next(); state.inTag = false; return 'tag'; }
                if (ch === '{' || ch === '}' || ch === ';') { stream.next(); state.first = true; return 'bracket'; }
                if (ch === '$' && stream.match(/^\$\{?\w+\}?/)) { state.first = false; return 'variable-2'; }
                if (state.first && !state.inTag && stream.match(/^[A-Za-z_][\w.-]*/)) { state.first = false; return 'keyword'; }
                if (stream.match(/^\d+(\.\d+)*[kKmMgGsShHdD]?(?![\w.])/)) { state.first = false; return 'number'; }
                stream.next();
                stream.eatWhile(/[^\s"'{};<>$]/);
                state.first = false;
                return null;
            },
            lineComment: '#',
        };
    });
})();

// pre-save check: Apache tags must close (no self-closing tags, <Directory /> opens a block), other servers need balanced braces and quotes
function opConfIssues(content, kind, msg) {
    var issues = [];
    var add = function (line, text) { issues.push({ line: line, text: text }); };
    var fmt = function (m, vals) { return m.replace(/\{(\w+)\}/g, function (all, k) { return k in vals ? vals[k] : all; }); };
    var lines = content.split('\n');
    if (kind === 'apache') {
        var stack = [], re = /<(\/?)\s*([A-Za-z][A-Za-z0-9]*)\b[^>]*>/g;
        lines.forEach(function (line, i) {
            if (line.trim().charAt(0) === '#') return;
            var m;
            re.lastIndex = 0;
            while ((m = re.exec(line)) !== null) {
                var tag = m[2].toLowerCase();
                if (m[1] !== '/') { stack.push({ tag: tag, line: i + 1 }); continue; }
                if (!stack.length) add(i + 1, fmt(msg.unexpectedClose, { line: i + 1, tag: m[2] }));
                else if (stack[stack.length - 1].tag !== tag) { var top = stack.pop(); add(i + 1, fmt(msg.mismatch, { line: i + 1, tag: m[2], open: top.tag, openLine: top.line })); }
                else stack.pop();
            }
        });
        stack.forEach(function (t) { add(t.line, fmt(msg.neverClosed, { line: t.line, tag: t.tag })); });
        return issues;
    }
    var open = [];
    lines.forEach(function (line, i) {
        var quote = '', esc = false;
        for (var j = 0; j < line.length; j++) {
            var ch = line[j];
            if (esc) { esc = false; continue; }
            if (quote) {
                if (ch === '\\') esc = true;
                else if (ch === quote) quote = '';
                continue;
            }
            if (ch === '#') break;
            if (ch === '"' || ch === "'") quote = ch;
            else if (ch === '{') open.push(i + 1);
            else if (ch === '}') {
                if (open.length) open.pop();
                else add(i + 1, fmt(msg.unexpectedBrace, { line: i + 1 }));
            }
        }
        if (quote) add(i + 1, fmt(msg.unterminatedQuote, { line: i + 1, quote: quote }));
    });
    open.forEach(function (ln) { add(ln, fmt(msg.unclosedBrace, { line: ln })); });
    return issues;
}

// opConfEditor is the x-data of the editor <form>: cfg = { kind, msg: { i18n strings }, busy: { save, restore, steps, note }, unsaved, errorLine }
function opConfEditor(cfg) {
    var cm = null, gen = 0, overlay = null, current = null, flashTimer = 0, submitting = false;
    var msg = cfg.msg || window.opConfMsg || {};
    return {
        mode: '', query: '', lineNo: '', dirty: false, line: 1, col: 1, lines: 0,
        hits: 0, hit: 0, issues: [], confirmRestore: false, force: false,

        init() {
            var self = this;
            if (!this.$refs.ta) return;
            var keys = {
                'Ctrl-F': function () { self.open('find'); }, 'Cmd-F': function () { self.open('find'); },
                'Ctrl-G': function () { self.open('line'); }, 'Cmd-G': function () { self.open('line'); },
                'Ctrl-S': function () { self.save(); }, 'Cmd-S': function () { self.save(); },
                'Esc': function () { self.close(); },
            };
            cm = CodeMirror.fromTextArea(this.$refs.ta, { mode: 'opconf', theme: 'op', lineNumbers: true, indentUnit: 4, tabSize: 4, extraKeys: keys, screenReaderLabel: this.$refs.ta.getAttribute('aria-label') || '' });
            // a rejected save comes back with the user's text in the editor, it isn't on disk so it stays dirty
            gen = cfg.unsaved ? -1 : cm.changeGeneration();
            this.dirty = !!cfg.unsaved;
            if (cfg.errorLine > 0 && cfg.errorLine <= cm.lineCount()) cm.addLineClass(cfg.errorLine - 1, 'background', 'cm-op-error');
            cm.on('cursorActivity', function () { self.update(); });
            cm.on('change', function () {
                self.dirty = !cm.isClean(gen);
                if (self.query) self.search(false);
                self.update();
            });
            this.update();
            window.addEventListener('beforeunload', function (e) { if (self.dirty && !submitting) { e.preventDefault(); e.returnValue = ''; } });
        },
        update() {
            var c = cm.getCursor();
            this.line = c.line + 1;
            this.col = c.ch + 1;
            this.lines = cm.lineCount();
        },
        open(m) {
            this.mode = this.mode === m ? '' : m;
            if (!this.mode) { this.clearSearch(); cm.focus(); return; }
            if (m !== 'find') this.clearSearch();
            var sel = cm.getSelection();
            if (m === 'find' && sel && sel.indexOf('\n') === -1) { this.query = sel; }
            this.$nextTick(() => { var el = this.$refs[m]; el.focus(); el.select(); if (m === 'find') this.search(false); });
        },
        close() {
            if (!this.mode) return;
            this.mode = '';
            this.clearSearch();
            cm.focus();
        },

        // find: every match is highlighted through an overlay, the current one is marked and selected
        clearSearch() {
            if (overlay) { cm.removeOverlay(overlay); overlay = null; }
            if (current) { current.clear(); current = null; }
            this.hits = 0;
            this.hit = 0;
        },
        positions() {
            var needle = this.query.toLowerCase(), text = cm.getValue().toLowerCase(), out = [], i = 0;
            if (!needle) return out;
            while ((i = text.indexOf(needle, i)) !== -1 && out.length < 10000) { out.push(i); i += needle.length; }
            return out;
        },
        search(jump) {
            var q = this.query;
            if (overlay) { cm.removeOverlay(overlay); overlay = null; }
            if (current) { current.clear(); current = null; }
            var found = this.positions();
            this.hits = found.length;
            this.hit = 0;
            if (!q) return;
            var needle = q.toLowerCase();
            overlay = {
                token: function (stream) {
                    var rest = stream.string.slice(stream.pos).toLowerCase(), at = rest.indexOf(needle);
                    if (at === 0) { stream.pos += needle.length; return 'op-match'; }
                    if (at > 0) stream.pos += at; else stream.skipToEnd();
                    return null;
                },
            };
            cm.addOverlay(overlay, { opaque: false });
            if (jump && found.length) this.go(found, cm.indexFromPos(cm.getCursor('from')), 1);
        },
        // go selects the first match at or after `from` (dir 1) or before it (dir -1), wrapping around the file
        go(found, from, dir) {
            var idx = -1;
            if (dir > 0) { idx = found.findIndex(function (p) { return p >= from; }); if (idx === -1) idx = 0; }
            else { for (var k = found.length - 1; k >= 0; k--) { if (found[k] < from) { idx = k; break; } } if (idx === -1) idx = found.length - 1; }
            var a = cm.posFromIndex(found[idx]), b = cm.posFromIndex(found[idx] + this.query.length);
            if (current) current.clear();
            current = cm.markText(a, b, { className: 'cm-op-match-current' });
            cm.setSelection(a, b, { scroll: false });
            cm.scrollIntoView({ from: a, to: b }, Math.round(cm.getScrollInfo().clientHeight / 3));
            this.hit = idx + 1;
            this.hits = found.length;
        },
        findNext(dir) {
            dir = dir || 1;
            if (!this.query) return;
            var found = this.positions();
            if (!found.length) { this.hits = 0; this.hit = 0; return; }
            var from = dir > 0 ? cm.indexFromPos(cm.getCursor('to')) : cm.indexFromPos(cm.getCursor('from'));
            // a fresh search starts at the cursor, not after it
            if (dir > 0 && !this.hit) from = cm.indexFromPos(cm.getCursor('from'));
            this.go(found, from, dir);
        },
        onQuery() { this.search(true); },
        statusLabel() {
            return (msg.status || 'Line {line}, column {col} · {lines} lines').replace('{line}', this.line).replace('{col}', this.col).replace('{lines}', this.lines);
        },
        countLabel() {
            if (!this.query) return '';
            if (!this.hits) return msg.noMatches || 'No matches';
            return (msg.countOf || '{n} of {total}').replace('{n}', this.hit || '-').replace('{total}', this.hits);
        },

        goToLine() {
            var n = parseInt(this.lineNo, 10);
            if (!n) return;
            n = Math.min(Math.max(n, 1), cm.lineCount());
            this.jump(n);
            this.mode = '';
            this.lineNo = '';
        },
        // jump puts the cursor at the start of line n, scrolls it to the middle of the editor and flashes it
        jump(n) {
            var ln = n - 1;
            cm.setCursor({ line: ln, ch: 0 }, null, { scroll: false });
            cm.scrollIntoView({ line: ln, ch: 0 }, Math.round(cm.getScrollInfo().clientHeight / 2) - 10);
            cm.addLineClass(ln, 'background', 'cm-op-flash');
            clearTimeout(flashTimer);
            flashTimer = setTimeout(function () { cm.removeLineClass(ln, 'background', 'cm-op-flash'); }, 1400);
            this.$root.querySelector('.CodeMirror').scrollIntoView({ block: 'nearest' });
            cm.focus();
            this.update();
        },
        firstIssue() { if (this.issues.length) this.jump(this.issues[0].line); },

        save() {
            var btn = this.$refs.save;
            if (!btn || btn.disabled) return;
            this.$root.requestSubmit(btn);
        },
        saveAnyway() {
            this.force = true;
            this.$root.requestSubmit(this.$refs.save);
        },
        // onSubmit runs the block check before a save and sets the busy toast for whichever button was pressed
        onSubmit(e) {
            cm.save();
            var restore = e.submitter && e.submitter.value === 'restore_default';
            if (!restore && !this.force && cfg.kind) {
                this.issues = opConfIssues(cm.getValue(), cfg.kind, msg);
                if (this.issues.length) { e.preventDefault(); return; }
            }
            this.force = false;
            this.issues = [];
            var b = cfg.busy || {};
            if (b.save) {
                this.$root.dataset.busy = restore ? b.restore : b.save;
                this.$root.dataset.busySteps = b.steps || '';
                this.$root.dataset.busyNote = b.note || '';
                this.$root.dataset.busyLeave = 'no';
            }
            submitting = true;
        },
    };
}
