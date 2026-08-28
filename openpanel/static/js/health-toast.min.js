/*
HEALTH TOAST ENGINE

Pages call this with a list of currently-detected problems; it dedupes
against what was already shown to this browser (so the user isn't nagged
on every reload) and shows a toast for anything new or whose cooldown
has expired.

  reportHealthIssues('emails', [
    { id: 'quota:user@example.com', severity: 'warning', message: 'Email user@example.com is reaching its quota (85%).' },
  ], 24); // ttlHours: how long before this same issue can be shown again if still unresolved

issue.id is optional - if omitted, a hash of the message is used.
*/

const HEALTH_TOAST_STORAGE_KEY = 'healthToastSeen';

function healthToastHash(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        hash = (hash << 5) - hash + str.charCodeAt(i);
        hash |= 0;
    }
    return String(hash);
}

function reportHealthIssues(pageKey, issues, ttlHours = 24) {
    if (!issues || !issues.length) return;

    let store;
    try {
        store = JSON.parse(localStorage.getItem(HEALTH_TOAST_STORAGE_KEY)) || {};
    } catch (e) {
        store = {};
    }

    const ttlMs = ttlHours * 60 * 60 * 1000;
    const now = Date.now();
    const pageSeen = store[pageKey] || {};
    const currentIds = new Set();

    issues.forEach((issue) => {
        const id = issue.id || healthToastHash(issue.message);
        currentIds.add(id);

        const lastShown = pageSeen[id];
        if (!lastShown || (now - lastShown) >= ttlMs) {
            const type = issue.severity === 'error' ? 'error' : (issue.severity === 'info' ? 'info' : 'warning');
            showToast(issue.message, type, false, issue.link || false);
            pageSeen[id] = now;
        }
    });

    // drop entries for issues that are no longer present (resolved)
    Object.keys(pageSeen).forEach((id) => {
        if (!currentIds.has(id)) delete pageSeen[id];
    });

    store[pageKey] = pageSeen;
    localStorage.setItem(HEALTH_TOAST_STORAGE_KEY, JSON.stringify(store));
}
