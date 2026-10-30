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

    function updateSortIndicators(table, activeKey, activeDirection) {
        table.querySelectorAll('a[data-sort-key]').forEach(function (link) {
            const isActive = link.dataset.sortKey === activeKey && link.dataset.direction === activeDirection;
            link.querySelector('svg').classList.toggle('opacity-30', !isActive);
        });
    }

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

    function applySort(link) {
        const table = link.closest('table');
        if (!table) return;

        const key = link.dataset.sortKey;
        const type = link.dataset.sortType || 'string';
        const direction = link.dataset.direction;

        sortTableByColumn(table, key, type, direction);
        updateSortIndicators(table, key, direction);
        updateSortUrl(key, direction);
    }

    window.sortTableColumn = function (event, link) {
        event.preventDefault();
        applySort(link);
    };

    document.addEventListener('DOMContentLoaded', function () {
        const params = new URLSearchParams(window.location.search);
        const sortKey = params.get('sort');
        const direction = params.get('direction') === 'asc' ? 'asc' : 'desc';
        if (!sortKey) return;

        const link = document.querySelector(
            'a[data-sort-key="' + CSS.escape(sortKey) + '"][data-direction="' + direction + '"]'
        );
        if (link) applySort(link);
    });
})();
