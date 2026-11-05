// brand palette and Chart.js defaults for every chart in the panel, load right after chart.umd.min.js
(function () {
    // series order matches the Planner chart colors
    var SERIES = {
        blue: '59,130,246', emerald: '16,185,129', violet: '139,92,246', amber: '245,158,11',
        cyan: '6,182,212', pink: '236,72,153', orange: '249,115,22', gray: '156,163,175',
        yellow: '234,179,8', red: '239,68,68'
    };
    var ORDER = ['blue', 'emerald', 'violet', 'amber', 'cyan', 'pink', 'orange', 'gray'];

    function isDark() {
        return document.documentElement.classList.contains('dark');
    }

    function theme() {
        var d = isDark();
        return {
            grid: d ? '#1f2937' : '#e5e7eb',
            tick: '#6b7280',
            tipBg: d ? '#030712' : '#ffffff',
            tipBorder: d ? '#1f2937' : '#e5e7eb',
            tipTitle: d ? '#f9fafb' : '#111827',
            tipBody: d ? '#9ca3af' : '#4b5563'
        };
    }

    function rgb(name, alpha) {
        var v = SERIES[name] || name;
        return alpha === undefined ? 'rgb(' + v + ')' : 'rgba(' + v + ',' + alpha + ')';
    }

    function nth(i, alpha) {
        return rgb(ORDER[i % ORDER.length], alpha);
    }

    function tooltip() {
        var t = theme();
        return {
            backgroundColor: t.tipBg, borderColor: t.tipBorder, borderWidth: 1,
            titleColor: t.tipTitle, bodyColor: t.tipBody, padding: 10, cornerRadius: 6,
            boxWidth: 8, boxHeight: 8, usePointStyle: true
        };
    }

    function applyDefaults() {
        if (!window.Chart) return;
        var t = theme();
        var C = window.Chart.defaults;
        C.font.family = 'Geist, ui-sans-serif, system-ui, sans-serif';
        C.font.size = 12;
        C.color = t.tick;
        C.borderColor = t.grid;
        C.elements.line.borderWidth = 2;
        C.elements.line.tension = 0.3;
        C.elements.point.radius = 0;
        C.elements.point.hoverRadius = 4;
        C.plugins.legend.labels.boxWidth = 8;
        C.plugins.legend.labels.boxHeight = 8;
        C.plugins.legend.labels.usePointStyle = true;
        Object.assign(C.plugins.tooltip, tooltip());
    }

    // charts drawn before a theme switch get the new grid, tick and tooltip colors
    function recolor() {
        if (!window.Chart) return;
        applyDefaults();
        var t = theme();
        Object.values(window.Chart.instances || {}).forEach(function (chart) {
            var o = chart.options;
            Object.values(o.scales || {}).forEach(function (s) {
                if (s.grid && s.grid.display !== false) s.grid.color = t.grid;
                if (s.ticks) s.ticks.color = t.tick;
            });
            if (o.plugins && o.plugins.tooltip) Object.assign(o.plugins.tooltip, tooltip());
            if (o.plugins && o.plugins.legend && o.plugins.legend.labels) o.plugins.legend.labels.color = t.tick;
            chart.update('none');
        });
    }

    window.opChart = { series: SERIES, order: ORDER, rgb: rgb, nth: nth, theme: theme, tooltip: tooltip, isDark: isDark };
    applyDefaults();
    new MutationObserver(recolor).observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });
})();
