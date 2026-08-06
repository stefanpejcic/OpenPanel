// Shared human-readable date helpers.
// Translated strings come from window.RelativeDateI18n, populated server-side
// via Babel's _() in templates/base.html, so wording stays translation-ready.

function relativeDatePlural(value, singularKey, pluralKey, unit) {
    const i18n = window.RelativeDateI18n || {};
    if (value === 1 && i18n[singularKey]) {
        return i18n[singularKey];
    }
    const template = i18n[pluralKey] || ('%%(n)s ' + unit + 's ago');
    return template.replace('%%(n)s', value);
}

// Calendar-bucketed "today" / "yesterday" / "X weeks/months/years ago".
function formatRelativeDay(input) {
    const i18n = window.RelativeDateI18n || {};
    const date = input instanceof Date ? input : new Date(input);
    const diffMs = new Date().getTime() - date.getTime();

    if (isNaN(diffMs) || diffMs < 0) {
        return i18n.today || 'today';
    }

    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return i18n.today || 'today';
    if (diffDays === 1) return i18n.yesterday || 'yesterday';
    if (diffDays < 7) return relativeDatePlural(diffDays, 'dayAgo', 'daysAgo', 'day');
    if (diffDays < 30) {
        const weeks = Math.floor(diffDays / 7);
        return relativeDatePlural(weeks, 'weekAgo', 'weeksAgo', 'week');
    }
    if (diffDays < 365) {
        const months = Math.floor(diffDays / 30);
        return months === 1 && i18n.lastMonth
            ? i18n.lastMonth
            : relativeDatePlural(months, 'monthAgo', 'monthsAgo', 'month');
    }

    const years = Math.floor(diffDays / 365);
    return years === 1 && i18n.lastYear
        ? i18n.lastYear
        : relativeDatePlural(years, 'yearAgo', 'yearsAgo', 'year');
}

// Auto-applies formatRelativeDay() to any element marked with
// data-relative-date, reading the raw date from the attribute (or the
// element's own text on first run) and writing the formatted result back.
// The exact date is kept available on hover via the title attribute.
function applyRelativeDates(root) {
    (root || document).querySelectorAll('[data-relative-date]').forEach((el) => {
        const raw = el.getAttribute('data-relative-date') || el.textContent.trim();
        if (!raw) return;

        const date = new Date(raw);
        if (isNaN(date.getTime())) return;

        if (!el.getAttribute('data-relative-date')) {
            el.setAttribute('data-relative-date', raw);
        }
        if (!el.getAttribute('title')) {
            el.setAttribute('title', raw);
        }

        el.textContent = formatRelativeDay(date);
    });
}

document.addEventListener('DOMContentLoaded', function () {
    applyRelativeDates();
});
