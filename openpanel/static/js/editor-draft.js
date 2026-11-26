// keeps the last unsaved edit of a raw file editor in localStorage and offers it back via the editor_draft partial
(function () {
    // plain LCS line diff, these files are small enough that O(n*m) is fine
    function lineDiff(a, b) {
        const A = a.split('\n'), B = b.split('\n');
        const n = A.length, m = B.length;
        const dp = Array.from({ length: n + 1 }, () => new Uint32Array(m + 1));
        for (let i = n - 1; i >= 0; i--)
            for (let j = m - 1; j >= 0; j--)
                dp[i][j] = A[i] === B[j] ? dp[i + 1][j + 1] + 1 : Math.max(dp[i + 1][j], dp[i][j + 1]);
        const out = [];
        let i = 0, j = 0;
        while (i < n && j < m) {
            if (A[i] === B[j]) { out.push([' ', A[i]]); i++; j++; }
            else if (dp[i + 1][j] >= dp[i][j + 1]) out.push(['-', A[i++]]);
            else out.push(['+', B[j++]]);
        }
        while (i < n) out.push(['-', A[i++]]);
        while (j < m) out.push(['+', B[j++]]);
        return out;
    }

    const lineClass = {
        '+': 'px-3 whitespace-pre-wrap break-all bg-emerald-50 text-emerald-800 dark:bg-emerald-950/70 dark:text-emerald-300',
        '-': 'px-3 whitespace-pre-wrap break-all bg-red-50 text-red-800 dark:bg-red-950/70 dark:text-red-300',
        ' ': 'px-3 whitespace-pre-wrap break-all text-gray-500',
    };

    // opts: key, get(), set(value), onChange(cb), initial (what the server sent, use the textarea's defaultValue since Firefox refills textareas on reload), root, form (warns before leaving unless it's submitting)
    window.opEditorDraft = function (opts) {
        const root = opts.root || document.querySelector('[data-op-draft]');
        if (!root) return null;
        const $ = (s) => root.querySelector(s);
        const key = 'opdraft:' + opts.key;
        let onDisk = typeof opts.initial === 'string' ? opts.initial : opts.get();
        let draft = null;

        const read = () => { try { return JSON.parse(localStorage.getItem(key) || 'null'); } catch (e) { return null; } };
        const drop = () => { try { localStorage.removeItem(key); } catch (e) {} };
        const diffBox = $('[data-op-draft-diff]');
        const viewBtn = $('[data-op-draft-view]');
        const hide = () => { root.hidden = true; diffBox.hidden = true; viewBtn.setAttribute('aria-expanded', 'false'); };

        draft = read();
        // same as the file means it got saved, nothing to offer
        if (draft && (typeof draft.content !== 'string' || draft.content === onDisk)) { drop(); draft = null; }
        if (draft) {
            $('[data-op-draft-time]').textContent = new Date(draft.ts).toLocaleString(root.dataset.locale || undefined);
            $('[data-op-draft-stale]').hidden = draft.base === onDisk;
            root.hidden = false;
        }

        viewBtn.addEventListener('click', () => {
            const open = diffBox.hidden;
            if (open) {
                const body = $('[data-op-draft-body]');
                body.textContent = '';
                let added = 0, removed = 0;
                lineDiff(onDisk, draft.content).forEach(([op, text]) => {
                    if (op === '+') added++;
                    if (op === '-') removed++;
                    const line = document.createElement('div');
                    line.className = lineClass[op];
                    line.textContent = op + ' ' + text;
                    body.appendChild(line);
                });
                $('[data-op-draft-stats]').textContent = '+' + added + ' −' + removed;
            }
            diffBox.hidden = !open;
            viewBtn.setAttribute('aria-expanded', String(open));
        });
        $('[data-op-draft-restore]').addEventListener('click', () => { opts.set(draft.content); hide(); });
        $('[data-op-draft-discard]').addEventListener('click', () => { drop(); hide(); });

        let timer = null;
        opts.onChange(() => {
            clearTimeout(timer);
            timer = setTimeout(() => {
                const content = opts.get();
                if (content === onDisk) { drop(); return; }
                try { localStorage.setItem(key, JSON.stringify({ content, base: onDisk, ts: Date.now() })); } catch (e) {}
                hide();
            }, 500);
        });

        let submitting = false;
        if (opts.form) opts.form.addEventListener('submit', () => { submitting = true; });
        window.addEventListener('beforeunload', (e) => {
            if (submitting || opts.get() === onDisk) return;
            e.preventDefault();
            e.returnValue = '';
        });

        return {
            // for pages that save over fetch, the saved text becomes the new baseline
            saved(content) {
                clearTimeout(timer);
                onDisk = content === undefined ? opts.get() : content;
                drop();
                hide();
            },
        };
    };
})();
