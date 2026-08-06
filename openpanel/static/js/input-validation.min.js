/*
Live validation hints for any <input>/<textarea> that already carries
`required` / `pattern` / `minlength` / `maxlength` / `min` / `max` / type="email|url"
attributes. No per-page wiring needed: this listens on the whole document,
so it picks up fields rendered later by Alpine/x-if/fetch too.

UX rule: negative (red) states only appear once a field has been blurred at
least once, so we don't yell at the user before they've finished typing.
Positive (green) and neutral "here's what's needed" hints can show earlier,
since those aren't accusatory.

Opt out of a single field with data-no-live-validate.

Password fields are always skipped: pages that need password UX (strength
meter, confirm-match, generate/toggle) implement it themselves (see the
Alpine `passwordStrength` component), so this library would only duplicate
or fight that.
*/

(function () {
    const I18N = window.InputValidationI18n || {};
    const t = (key, fallback) => I18N[key] || fallback;

    function isEligible(el) {
        if (!el.matches || !el.matches('input, textarea')) return false;
        if (el.hasAttribute('data-no-live-validate')) return false;
        if (el.disabled || el.readOnly) return false;
        const type = (el.type || 'text').toLowerCase();
        if (['hidden', 'submit', 'button', 'reset', 'checkbox', 'radio', 'file', 'image', 'range', 'color'].includes(type)) return false;
        if (type === 'password') return false;
        const name = (el.name || el.id || '').toLowerCase();
        if (name.includes('password')) return false;
        return el.hasAttribute('required') || el.hasAttribute('pattern') ||
            el.hasAttribute('minlength') || el.hasAttribute('maxlength') ||
            el.hasAttribute('min') || el.hasAttribute('max') ||
            type === 'email' || type === 'url';
    }

    function genericMessage(el) {
        const validity = el.validity;
        if (validity.valueMissing) return t('required', 'This field is required.');
        if (validity.typeMismatch) {
            const type = (el.type || '').toLowerCase();
            if (type === 'email') return t('typeEmail', 'Enter a valid email address.');
            if (type === 'url') return t('typeUrl', 'Enter a valid URL.');
            return t('invalidFormat', 'Invalid format.');
        }
        if (validity.tooShort) {
            return t('tooShort', 'At least {min} characters required ({count} so far).')
                .replace('{min}', el.getAttribute('minlength'))
                .replace('{count}', String(el.value.length));
        }
        if (validity.tooLong) {
            return t('tooLong', 'Maximum {max} characters.').replace('{max}', el.getAttribute('maxlength'));
        }
        if (validity.rangeUnderflow) return t('rangeUnderflow', 'Value must be at least {min}.').replace('{min}', el.getAttribute('min'));
        if (validity.rangeOverflow) return t('rangeOverflow', 'Value must be at most {max}.').replace('{max}', el.getAttribute('max'));
        if (validity.patternMismatch) return el.title || t('invalidFormat', 'Invalid format.');
        return el.title || t('invalidFormat', 'Invalid format.');
    }

    function hintHost(el) {
        // Joined/flex input groups (input + suffix buttons, etc.) need the
        // hint below the whole row, not squeezed between flex items.
        let insertAfter = el;
        const parent = el.parentElement;
        if (parent && parent.classList.contains('flex')) {
            const realSiblings = Array.from(parent.children).filter((c) => !c.classList.contains('iv-hint-host'));
            if (realSiblings.length > 1) insertAfter = parent;
        }

        let host = insertAfter.nextElementSibling;
        if (host && host.classList.contains('iv-hint-host')) return host;
        host = document.createElement('div');
        host.className = 'iv-hint-host';
        insertAfter.insertAdjacentElement('afterend', host);
        return host;
    }

    function renderMessage(host, status, message) {
        host.innerHTML = '';
        if (!message) return;
        const p = document.createElement('p');
        p.className = 'iv-hint iv-' + status;
        p.textContent = message;
        host.appendChild(p);
    }

    function evaluate(el) {
        const host = hintHost(el);
        const value = el.value;
        const touched = el.dataset.ivTouched === '1';

        el.classList.remove('iv-valid', 'iv-invalid');

        if (value.length === 0) {
            if (touched && el.hasAttribute('required')) {
                el.classList.add('iv-invalid');
                renderMessage(host, 'invalid', t('required', 'This field is required.'));
            } else {
                renderMessage(host, 'neutral', '');
            }
            return;
        }

        if (el.checkValidity()) {
            el.classList.add('iv-valid');
            renderMessage(host, 'valid', t('looksGood', 'Looks good.'));
        } else {
            const message = genericMessage(el);
            if (touched) {
                el.classList.add('iv-invalid');
                renderMessage(host, 'invalid', message);
            } else {
                renderMessage(host, 'neutral', message);
            }
        }
    }

    document.addEventListener('input', (e) => {
        if (!isEligible(e.target)) return;
        evaluate(e.target);
    }, true);

    document.addEventListener('focusout', (e) => {
        if (!isEligible(e.target)) return;
        e.target.dataset.ivTouched = '1';
        evaluate(e.target);
    }, true);
})();
