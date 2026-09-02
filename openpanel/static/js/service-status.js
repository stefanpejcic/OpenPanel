// Auto-refreshes the status badge on /cache/<name> and /services/<name>
// pages while the container is in a transitional state (e.g. just enabled
// or disabled), so the user doesn't have to F5 to see the real status.

// Keep in sync with serviceStatusMap in internal/modules/services/render.go -
// "stopping" (libpod's real State.Status for a container mid-shutdown) was
// missing from both lists, so a page render landing in that few-second
// window showed "Unknown" with nothing polling to ever refresh it.
const SERVICE_STATUS_TRANSITIONAL = ['starting', 'created', 'restarting', 'removing', 'stopping'];
const SERVICE_STATUS_POLL_MS = 2000;

function computeServiceStatusKey(containerState, healthStatus) {
    if (containerState === 'running') {
        return ['healthy', 'unhealthy', 'starting'].includes(healthStatus) ? healthStatus : 'running';
    }
    return containerState;
}

function pollServiceStatusBadge(badge) {
    const statusMap = JSON.parse(badge.dataset.statusMap);
    const fetchUrl = badge.dataset.fetchUrl;
    const bars = badge.querySelectorAll('.service-status-bar');
    const label = badge.querySelector('.service-status-label');

    const tick = async () => {
        try {
            const res = await fetch(`${fetchUrl}?output=json`);
            const data = await res.json();
            const statusKey = computeServiceStatusKey(data.container_state, data.health_status);

            if (!SERVICE_STATUS_TRANSITIONAL.includes(statusKey)) {
                // Status settled: other parts of the page (action button,
                // container stats, etc.) depend on it too, so reload once.
                window.location.reload();
                return;
            }

            const [color, text] = statusMap[statusKey] || statusMap.unknown || ['orange-500', 'Unknown'];
            bars.forEach((bar) => {
                bar.className = `h-3.5 w-1 rounded-sm bg-${color} dark:bg-${color}`;
            });
            if (label) label.textContent = text;

            setTimeout(tick, SERVICE_STATUS_POLL_MS);
        } catch (err) {
            console.error('Service status poll error:', err);
        }
    };

    setTimeout(tick, SERVICE_STATUS_POLL_MS);
}

document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('[data-status-poll]').forEach((badge) => {
        if (SERVICE_STATUS_TRANSITIONAL.includes(badge.dataset.statusKey)) {
            pollServiceStatusBadge(badge);
        }
    });
});
