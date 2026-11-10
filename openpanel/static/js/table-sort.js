// Shared client-side column sorting for admin tables.
//
// Usage: render headers with the `sort_header(label, key, type)` macro from
// macros.html, and mark each row's sortable <td> with `data-sort-col="key"`
// (matching the header's key) plus an optional `data-sort-value="..."` when
// the visible text isn't the raw sortable value (badges, truncation, nested
// links). Rows are matched to columns by that key, not DOM position, so
// conditionally-rendered columns and duplicate cells (e.g. an edit-mode
// counterpart of a display cell) don't throw off the sort - a row only
// participates if it has a `[data-sort-col]` cell for the clicked key.
//
// Add `data-preserve-sort` to any link/button whose href should keep the
// current sort/direction query params (e.g. a "Refresh" link), so a manual
// reload doesn't lose the sort.

(function () {
    // Converts an elapsed-time string ("7s", "23m13s", "1h2m3s") or a
    // classic "HH:MM:SS" string into total seconds, so "duration" columns
    // sort numerically instead of alphabetically.
    function parseDurationSeconds(value) {
        if (!value) return 0;

        const durationMatch = value.match(/^(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?$/);
        if (durationMatch && (durationMatch[1] || durationMatch[2] || durationMatch[3])) {
            const hours = parseInt(durationMatch[1] || '0', 10);
            const minutes = parseInt(durationMatch[2] || '0', 10);
            const seconds = parseInt(durationMatch[3] || '0', 10);
            return hours * 3600 + minutes * 60 + seconds;
        }

        const clockParts = value.split(':').map(Number);
        if (clockParts.length >= 2 && clockParts.every(part => !Number.isNaN(part))) {
            return clockParts.reduce((total, part) => total * 60 + part, 0);
        }

        return parseFloat(value) || 0;
    }

    function compareValues(valueA, valueB, type, direction) {
        if (type === 'number') {
            const numA = parseFloat(valueA) || 0;
            const numB = parseFloat(valueB) || 0;
            return direction === 'asc' ? numA - numB : numB - numA;
        }

        if (type === 'duration') {
            const durA = parseDurationSeconds(valueA);
            const durB = parseDurationSeconds(valueB);
            return direction === 'asc' ? durA - durB : durB - durA;
        }

        const strA = String(valueA).toLowerCase();
        const strB = String(valueB).toLowerCase();
        if (strA < strB) return direction === 'asc' ? -1 : 1;
        if (strA > strB) return direction === 'asc' ? 1 : -1;
        return 0;
    }

    function sortTableByColumn(table, key, type, direction) {
        const tbody = table.tBodies[0];
        if (!tbody) return;

        if (key.startsWith('__c')) return sortByPosition(table, parseInt(key.slice(3), 10), direction);

        const selector = '[data-sort-col="' + CSS.escape(key) + '"]';
        const rows = Array.from(tbody.rows).filter(row => row.querySelector(selector));

        rows.sort((rowA, rowB) => {
            const cellA = rowA.querySelector(selector);
            const cellB = rowB.querySelector(selector);
            const valueA = cellA.dataset.sortValue ?? cellA.textContent.trim();
            const valueB = cellB.dataset.sortValue ?? cellB.textContent.trim();
            return compareValues(valueA, valueB, type, direction);
        });

        rows.forEach(row => tbody.appendChild(row));
    }

    // "5.0 KB", "1.2 GB", "75%", "300s" become numbers so they sort by size, not alphabetically
    const UNITS = { b: 1, k: 1e3, kb: 1e3, kib: 1024, m: 1e6, mb: 1e6, mib: 1048576, g: 1e9, gb: 1e9, gib: 1073741824, t: 1e12, tb: 1e12, tib: 1099511627776, '%': 1, s: 1, ms: 0.001 };
    function autoNumber(v) {
        const m = String(v).trim().replace(/,/g, '').match(/^(-?\d+(?:\.\d+)?)\s*([a-z%]*)$/i);
        if (!m) return null;
        const unit = m[2].toLowerCase();
        if (unit && !(unit in UNITS)) return null;
        return parseFloat(m[1]) * (unit ? UNITS[unit] : 1);
    }
    function autoCompare(a, b, direction) {
        const na = autoNumber(a), nb = autoNumber(b);
        let r = (na !== null && nb !== null) ? na - nb : String(a).localeCompare(String(b), undefined, { numeric: true, sensitivity: 'base' });
        return direction === 'asc' ? r : -r;
    }
    function sortByPosition(table, index, direction) {
        const tbody = table.tBodies[0];
        const width = table.tHead.rows[table.tHead.rows.length - 1].cells.length;
        // group headers and empty-state rows (colspan) stay where they are
        const rows = Array.from(tbody.rows).filter(r => r.cells.length === width && r.cells[index]);
        const val = r => { const c = r.cells[index]; return (c.dataset.sortValue ?? c.textContent).trim(); };
        rows.sort((a, b) => autoCompare(val(a), val(b), direction));
        rows.forEach(r => tbody.appendChild(r));
    }

    // tables without hand-written sort headers get a sort button on every labelled column, the file manager and Trash excepted
    function autoSortable(table) {
        if (table.dataset.autosort || table.hasAttribute('data-no-sort') || table.id === 'filemanager_table' || table.closest('#filemanager_table, [data-no-sort]')) return;
        if (!table.tHead || !table.tBodies[0] || table.querySelector('[data-sort-key]')) return;
        table.dataset.autosort = '1';
        const head = table.tHead.rows[table.tHead.rows.length - 1];
        Array.from(head.cells).forEach((th, i) => {
            const label = th.textContent.trim();
            if (!label || th.colSpan > 1 || th.querySelector('input, select, button, a') || th.hasAttribute('x-text') || th.hasAttribute('x-html') || th.closest('template')) return;
            // columns of buttons, menus or selects have nothing to sort by
            if (/^(actions?|options|manage)$/i.test(label) || isControlColumn(table, i, head.cells.length)) return;
            const btn = document.createElement('button');
            btn.type = 'button';
            btn.dataset.sortKey = '__c' + i;
            btn.dataset.sortType = 'auto';
            btn.className = 'group -mx-1.5 inline-flex cursor-pointer select-none items-center gap-1.5 whitespace-nowrap rounded px-1.5 py-1 text-xs font-semibold text-gray-900 transition hover:bg-gray-100 dark:text-gray-50 dark:hover:bg-gray-900 outline outline-0 outline-offset-2 outline-blue-500 focus-visible:outline-2';
            btn.setAttribute('onclick', 'sortTableColumn(event, this);');
            while (th.firstChild) btn.appendChild(th.firstChild);
            const icon = document.createElement('i');
            icon.setAttribute('aria-hidden', 'true');
            btn.appendChild(icon);
            th.appendChild(btn);
        });
        updateSortIndicators(table, null, null);
    }
    function isControlColumn(table, index, width) {
        const rows = Array.from(table.tBodies[0].rows).filter(r => r.cells.length === width).slice(0, 5);
        if (!rows.length) return false;
        return rows.every(r => {
            const cell = r.cells[index].cloneNode(true);
            cell.querySelectorAll('button, select, input, form, [role="menu"], script, template').forEach(n => n.remove());
            return cell.textContent.trim() === '';
        });
    }
    function autoSortAll() { document.querySelectorAll('table').forEach(autoSortable); }

    const IDLE_ICON = 'ri-arrow-up-down-line text-sm leading-none text-gray-400 group-hover:text-gray-600 dark:text-gray-600 dark:group-hover:text-gray-400';
    const ACTIVE_ICON = 'text-sm leading-none text-gray-900 dark:text-gray-50';

    // one button per column: grey up/down arrow until sorted, then a single arrow for the direction
    function updateSortIndicators(table, activeKey, activeDirection) {
        table.querySelectorAll('[data-sort-key]').forEach(function (button) {
            const isActive = button.dataset.sortKey === activeKey;
            const icon = button.querySelector('i');
            const th = button.closest('th');
            if (isActive) {
                button.dataset.direction = activeDirection;
                if (icon) icon.className = (activeDirection === 'asc' ? 'ri-arrow-up-line ' : 'ri-arrow-down-line ') + ACTIVE_ICON;
                if (th) th.setAttribute('aria-sort', activeDirection === 'asc' ? 'ascending' : 'descending');
            } else {
                delete button.dataset.direction;
                if (icon) icon.className = IDLE_ICON;
                if (th) th.removeAttribute('aria-sort');
            }
        });
    }
    window.updateSortIndicators = updateSortIndicators;

    function updateSortUrl(key, direction) {
        const url = new URL(window.location.href);
        url.searchParams.set('sort', key);
        url.searchParams.set('direction', direction);
        window.history.replaceState({}, '', url);

        document.querySelectorAll('[data-preserve-sort]').forEach(function (link) {
            const href = link.getAttribute('href');
            if (!href) return;
            const linkUrl = new URL(href, window.location.origin);
            linkUrl.searchParams.set('sort', key);
            linkUrl.searchParams.set('direction', direction);
            link.setAttribute('href', linkUrl.pathname + linkUrl.search);
        });
    }

    function applySort(link, direction) {
        const table = link.closest('table');
        if (!table) return;

        const key = link.dataset.sortKey;
        const type = link.dataset.sortType || 'string';

        sortTableByColumn(table, key, type, direction);
        updateSortIndicators(table, key, direction);
        updateSortUrl(key, direction);
    }

    window.sortTableColumn = function (event, link) {
        event.preventDefault();
        applySort(link, link.dataset.direction === 'asc' ? 'desc' : 'asc');
    };

    document.addEventListener('DOMContentLoaded', autoSortAll);
    // tables Alpine renders later
    document.addEventListener('alpine:initialized', () => setTimeout(autoSortAll, 0));
    window.autoSortTables = autoSortAll;

    document.addEventListener('DOMContentLoaded', function () {
        const params = new URLSearchParams(window.location.search);
        const sortKey = params.get('sort');
        const direction = params.get('direction') === 'asc' ? 'asc' : 'desc';
        if (!sortKey) return;

        const link = document.querySelector('[data-sort-key="' + CSS.escape(sortKey) + '"]');
        if (link) applySort(link, direction);
    });
})();
