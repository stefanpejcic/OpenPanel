// Fits a tab bar on one row and moves the tabs that don't fit into a "More" dropdown that opens on hover.
// Usage: add data-tab-overflow (and data-more-label for the translated "More") to the tablist <ul>.
(function () {
    const STYLE = `
.tab-overflow{display:flex!important;flex-wrap:nowrap;grid-template-columns:none!important}
.tab-overflow>li{min-width:0}
.tab-overflow-more{position:relative}
.tab-overflow-btn{display:inline-flex;align-items:center;justify-content:center;gap:.25rem;width:100%;max-width:14rem;padding:.5rem .75rem;border-radius:.125rem;font-size:.875rem;line-height:1.25rem;font-weight:500;white-space:nowrap;color:#6b7280;background:transparent;cursor:pointer}
.tab-overflow-btn:hover{color:#374151}
.tab-overflow-btn.is-active{background:#fff;color:#3b82f6;box-shadow:0 1px 2px 0 rgb(0 0 0/.05)}
.tab-overflow-label{overflow:hidden;text-overflow:ellipsis}
.tab-overflow-btn svg{flex-shrink:0}
.dark .tab-overflow-btn{color:#9ca3af}
.dark .tab-overflow-btn:hover{color:#e5e7eb}
.dark .tab-overflow-btn.is-active{background:#111827;color:#3b82f6}
.tab-overflow-menu{display:none;position:absolute;right:0;top:100%;z-index:30;min-width:12rem;padding:.25rem;margin-top:.25rem;background:#fff;border:1px solid #e5e7eb;border-radius:.375rem;box-shadow:0 10px 15px -3px rgb(0 0 0/.1),0 4px 6px -4px rgb(0 0 0/.1)}
.tab-overflow-menu:before{content:"";position:absolute;left:0;right:0;top:-.5rem;height:.5rem}
.tab-overflow-more.is-open .tab-overflow-menu{display:block}
.dark .tab-overflow-menu{background:#030712;border-color:#1f2937}
.tab-overflow-item{display:flex;align-items:center;gap:.5rem;width:100%;padding:.5rem .75rem;border-radius:.25rem;font-size:.875rem;line-height:1.25rem;color:#374151;text-align:left;white-space:nowrap;cursor:pointer;background:transparent}
.tab-overflow-item:hover{background:#f3f4f6}
.tab-overflow-item.is-active{color:#3b82f6;font-weight:500}
.dark .tab-overflow-item{color:#d1d5db}
.dark .tab-overflow-item:hover{background:#111827}
.dark .tab-overflow-item.is-active{color:#3b82f6}
.tab-overflow-item svg{flex-shrink:0;margin:0}
`;
    const CHEVRON = '<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true"><path fill-rule="evenodd" d="M1.646 4.646a.5.5 0 0 1 .708 0L8 10.293l5.646-5.647a.5.5 0 0 1 .708.708l-6 6a.5.5 0 0 1-.708 0l-6-6a.5.5 0 0 1 0-.708"/></svg>';

    function ensureStyle() {
        if (document.getElementById('tab-overflow-style')) return;
        const style = document.createElement('style');
        style.id = 'tab-overflow-style';
        style.textContent = STYLE;
        document.head.appendChild(style);
    }

    function tabLink(li) {
        return li.querySelector('a,button');
    }

    function tabText(li) {
        const link = tabLink(li);
        if (!link) return '';
        const titled = link.querySelector('[title]');
        return (titled && titled.getAttribute('title')) || link.textContent.trim();
    }

    // the tabs' own Alpine :class puts bg-white on the active one
    function isActive(li) {
        const link = tabLink(li);
        return !!link && link.classList.contains('bg-white');
    }

    function init(ul) {
        if (ul.dataset.tabOverflowReady) return;
        ul.dataset.tabOverflowReady = '1';
        ensureStyle();
        ul.classList.add('tab-overflow');

        const items = Array.from(ul.children).filter(el => el.tagName === 'LI');
        const moreLabel = ul.dataset.moreLabel || 'More';

        const more = document.createElement('li');
        more.className = 'tab-overflow-more';
        more.setAttribute('role', 'presentation');
        more.innerHTML = '<button type="button" class="tab-overflow-btn" aria-haspopup="true" aria-expanded="false"><span class="tab-overflow-label"></span>' + CHEVRON + '</button><div class="tab-overflow-menu" role="menu"></div>';
        ul.appendChild(more);
        const btn = more.querySelector('.tab-overflow-btn');
        const label = more.querySelector('.tab-overflow-label');
        const menu = more.querySelector('.tab-overflow-menu');
        let hidden = [];
        let closeTimer = null;

        function open() {
            clearTimeout(closeTimer);
            more.classList.add('is-open');
            btn.setAttribute('aria-expanded', 'true');
        }
        function close() {
            more.classList.remove('is-open');
            btn.setAttribute('aria-expanded', 'false');
        }

        more.addEventListener('mouseenter', open);
        more.addEventListener('mouseleave', () => { closeTimer = setTimeout(close, 150); });
        btn.addEventListener('click', () => (more.classList.contains('is-open') ? close() : open()));
        document.addEventListener('click', e => { if (!more.contains(e.target)) close(); });
        document.addEventListener('keydown', e => { if (e.key === 'Escape') close(); });

        function refresh() {
            const activeHidden = hidden.find(isActive);
            label.textContent = activeHidden ? tabText(activeHidden) : moreLabel;
            btn.classList.toggle('is-active', !!activeHidden);
            menu.querySelectorAll('.tab-overflow-item').forEach((item, i) => item.classList.toggle('is-active', isActive(hidden[i])));
        }

        function renderMenu() {
            menu.innerHTML = '';
            hidden.forEach(li => {
                const link = tabLink(li);
                const item = document.createElement('button');
                item.type = 'button';
                item.className = 'tab-overflow-item';
                item.setAttribute('role', 'menuitem');
                const icon = link && link.querySelector('svg');
                if (icon) item.appendChild(icon.cloneNode(true));
                item.appendChild(document.createTextNode(tabText(li)));
                item.addEventListener('click', e => {
                    e.preventDefault();
                    if (link) link.click();
                    close();
                });
                menu.appendChild(item);
            });
        }

        function layout() {
            items.forEach(li => { li.style.display = ''; li.style.flex = '0 0 auto'; });
            more.style.display = '';
            more.style.flex = '0 0 auto';
            label.textContent = moreLabel;

            const cs = getComputedStyle(ul);
            const avail = ul.clientWidth - parseFloat(cs.paddingLeft) - parseFloat(cs.paddingRight) - 1;
            const widths = items.map(li => li.getBoundingClientRect().width);
            const total = widths.reduce((a, b) => a + b, 0);

            hidden = [];
            if (total <= avail) {
                more.style.display = 'none';
            } else {
                // keep tabs in order until the next one plus "More" no longer fits
                let used = more.getBoundingClientRect().width;
                let fit = 0;
                while (fit < items.length && used + widths[fit] <= avail) {
                    used += widths[fit];
                    fit++;
                }
                hidden = items.slice(fit);
                hidden.forEach(li => { li.style.display = 'none'; });
            }
            items.forEach(li => { if (li.style.display !== 'none') li.style.flex = '1 1 auto'; });
            renderMenu();
            refresh();
        }

        let frame = 0;
        const schedule = fn => { cancelAnimationFrame(frame); frame = requestAnimationFrame(fn); };
        new ResizeObserver(() => schedule(layout)).observe(ul);
        // Alpine flips the active classes when a tab is picked, keep the More button in sync
        new MutationObserver(records => {
            if (records.some(r => !more.contains(r.target))) schedule(refresh);
        }).observe(ul, { subtree: true, attributes: true, attributeFilter: ['class', 'aria-selected'] });
        ul.addEventListener('tab-overflow:relayout', () => schedule(layout));
        layout();
    }

    function initAll() {
        document.querySelectorAll('[data-tab-overflow]').forEach(init);
    }

    window.initTabOverflow = init;
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initAll);
    } else {
        initAll();
    }
    // Alpine can swap tab classes after load, so measure again once fonts and icons settle
    window.addEventListener('load', () => document.querySelectorAll('[data-tab-overflow]').forEach(ul => ul.dispatchEvent(new Event('tab-overflow:relayout'))));
})();
