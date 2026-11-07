// branded stand-in for window.confirm, resolves true or false
window.opConfirm = function (message, opts) {
    opts = opts || {};
    const text = String(message || '').trim();
    const verb = (text.match(/^(delete|remove|kill|disable|clear|leave|detach|stop|reset|drop|terminate|restart|cancel)\b/i) || [])[1];
    const danger = opts.danger !== undefined ? opts.danger : !!verb;
    const okLabel = opts.ok || (verb ? verb.charAt(0).toUpperCase() + verb.slice(1).toLowerCase() : 'Continue');
    const btn = 'relative inline-flex items-center justify-center gap-1.5 whitespace-nowrap rounded-md border px-3 py-2 text-center text-sm font-medium shadow-xs transition-all duration-100 ease-in-out outline outline-0 outline-offset-2 outline-blue-500 focus-visible:outline-2';
    const okClass = danger
        ? 'border-transparent bg-red-600 text-white hover:bg-red-700 dark:bg-red-700 dark:hover:bg-red-600'
        : 'border-transparent bg-blue-500 text-white hover:bg-blue-600';
    const icon = danger
        ? '<span class="flex size-10 shrink-0 items-center justify-center rounded-full bg-red-50 text-red-600 dark:bg-red-500/10 dark:text-red-500"><svg xmlns="http://www.w3.org/2000/svg" class="remixicon size-5" aria-hidden="true" viewBox="0 0 24 24" fill="currentColor"><path d="M12.8659 3.00017L22.3922 19.5002C22.6684 19.9785 22.5045 20.5901 22.0262 20.8662C21.8742 20.954 21.7017 21.0002 21.5262 21.0002H2.47363C1.92135 21.0002 1.47363 20.5525 1.47363 20.0002C1.47363 19.8246 1.51984 19.6522 1.60761 19.5002L11.1339 3.00017C11.41 2.52187 12.0216 2.358 12.4999 2.63414C12.6519 2.72191 12.7782 2.84815 12.8659 3.00017ZM4.20568 19.0002H19.7941L11.9999 5.50017L4.20568 19.0002ZM10.9999 16.0002H12.9999V18.0002H10.9999V16.0002ZM10.9999 9.00017H12.9999V14.0002H10.9999V9.00017Z"/></svg></span>'
        : '';
    return new Promise(function (resolve) {
        const prev = document.activeElement;
        const wrap = document.createElement('div');
        wrap.className = 'fixed inset-0 z-[60] flex items-center justify-center bg-black/70 p-4';
        wrap.setAttribute('role', danger ? 'alertdialog' : 'dialog');
        wrap.setAttribute('aria-modal', 'true');
        wrap.setAttribute('aria-labelledby', 'op-confirm-title');
        wrap.setAttribute('aria-describedby', 'op-confirm-text');
        wrap.innerHTML =
            '<div class="w-full max-w-lg rounded-md border border-gray-200 bg-white shadow-lg dark:border-gray-900 dark:bg-raised">' +
            '<div class="flex gap-4 px-6 pt-6">' + icon + '<div class="min-w-0">' +
            '<h2 id="op-confirm-title" class="text-base font-semibold text-gray-900 dark:text-gray-50"></h2>' +
            '<p id="op-confirm-text" class="mt-1 text-sm break-words text-gray-500 dark:text-gray-500"></p>' +
            (opts.confirmText ? '<label for="op-confirm-input" class="mt-4 block text-sm text-gray-900 dark:text-gray-50">Type <b class="font-semibold" data-name></b> to confirm</label>' +
                '<input id="op-confirm-input" type="text" autocomplete="off" spellcheck="false" class="mt-2 block rounded-md border border-gray-300 bg-white text-gray-900 shadow-xs outline-hidden transition placeholder:text-gray-400 focus:border-blue-500 focus:ring-2 focus:ring-blue-200 disabled:border-gray-300 disabled:bg-gray-100 disabled:text-gray-400 aria-[invalid=true]:border-red-500 aria-[invalid=true]:ring-2 aria-[invalid=true]:ring-red-200 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-50 dark:placeholder:text-gray-500 dark:focus:border-blue-700 dark:focus:ring-blue-700/30 dark:disabled:border-gray-700 dark:disabled:bg-gray-800 dark:disabled:text-gray-500 dark:aria-[invalid=true]:border-red-700 dark:aria-[invalid=true]:ring-red-700/30 w-full appearance-none px-2.5 py-2 text-sm">' : '') +
            '</div></div>' +
            '<div class="mt-6 flex flex-col-reverse gap-2 border-t border-gray-200 px-6 py-4 sm:flex-row sm:justify-end dark:border-gray-900">' +
            '<button type="button" data-act="cancel" class="' + btn + ' border-gray-300 bg-white text-gray-900 hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-50 dark:hover:bg-gray-900/60">Cancel</button>' +
            '<button type="button" data-act="ok" class="' + btn + ' ' + okClass + '"></button></div></div>';
        wrap.querySelector('#op-confirm-title').textContent = opts.title || (danger ? 'Are you sure?' : 'Please confirm');
        wrap.querySelector('#op-confirm-text').textContent = text;
        const ok = wrap.querySelector('[data-act=ok]');
        const cancel = wrap.querySelector('[data-act=cancel]');
        ok.textContent = okLabel;
        // type-to-confirm keeps the button disabled until the name matches
        const typed = wrap.querySelector('#op-confirm-input');
        if (typed) {
            wrap.querySelector('[data-name]').textContent = opts.confirmText;
            ok.disabled = true;
            ok.classList.add('disabled:opacity-50', 'disabled:pointer-events-none');
            typed.addEventListener('input', function () { ok.disabled = typed.value.trim() !== String(opts.confirmText); });
        }
        function close(result) {
            document.removeEventListener('keydown', onKey, true);
            wrap.remove();
            if (prev && prev.focus) prev.focus();
            resolve(result);
        }
        function onKey(e) {
            if (e.key === 'Escape') { e.preventDefault(); close(false); }
            if (e.key === 'Tab') {
                e.preventDefault();
                const stops = [typed, cancel, ok].filter(function (el) { return el && !el.disabled; });
                const at = stops.indexOf(document.activeElement);
                stops[(at + (e.shiftKey ? stops.length - 1 : 1)) % stops.length].focus();
            }
        }
        ok.addEventListener('click', function () { if (!ok.disabled) close(true); });
        cancel.addEventListener('click', function () { close(false); });
        wrap.addEventListener('mousedown', function (e) { if (e.target === wrap) close(false); });
        document.addEventListener('keydown', onKey, true);
        document.body.appendChild(wrap);
        (typed || (danger ? cancel : ok)).focus();
    });
};

// data-confirm on a form or its submit button asks before the form goes out
document.addEventListener('submit', function (e) {
    const form = e.target;
    const sub = e.submitter;
    const msg = (sub && sub.getAttribute('data-confirm')) || form.getAttribute('data-confirm');
    if (!msg) return;
    if (form.__opConfirmed) { form.__opConfirmed = false; return; }
    e.preventDefault();
    e.stopImmediatePropagation();
    // data-confirm-text="name" adds a type-the-name step for deletes that can't be undone
    const name = (sub && sub.getAttribute('data-confirm-text')) || form.getAttribute('data-confirm-text');
    window.opConfirm(msg, name ? { confirmText: name } : undefined).then(function (yes) {
        if (!yes) return;
        form.__opConfirmed = true;
        if (form.requestSubmit) form.requestSubmit(sub || undefined); else form.submit();
    });
}, true);
