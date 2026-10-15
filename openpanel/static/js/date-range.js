// shared date range picker (calendar + presets), markup lives in templates/partials/_date_range.html
function dateRange(from, to, locale, L, opts) {
    opts = opts || {};
    const pad = n => String(n).padStart(2, '0');
    const iso = d => d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate());
    const parse = s => { const [y, m, d] = s.split('-').map(Number); return new Date(y, m - 1, d); };
    const lang = (locale || 'en').replace('_', '-');
    let fmt, monthFmt, dayFmt, longFmt;
    try {
        fmt = new Intl.DateTimeFormat(lang, { day: 'numeric', month: 'short', year: 'numeric' });
        monthFmt = new Intl.DateTimeFormat(lang, { month: 'long', year: 'numeric' });
        dayFmt = new Intl.DateTimeFormat(lang, { weekday: 'short' });
        longFmt = new Intl.DateTimeFormat(lang, { dateStyle: 'full' });
    } catch (e) {
        fmt = new Intl.DateTimeFormat('en', { day: 'numeric', month: 'short', year: 'numeric' });
        monthFmt = new Intl.DateTimeFormat('en', { month: 'long', year: 'numeric' });
        dayFmt = new Intl.DateTimeFormat('en', { weekday: 'short' });
        longFmt = new Intl.DateTimeFormat('en', { dateStyle: 'full' });
    }

    const now = new Date();
    const today = iso(now);
    const daysAgo = n => { const d = new Date(now); d.setDate(d.getDate() - n); return iso(d); };
    const monthsAgo = n => { const d = new Date(now); d.setMonth(d.getMonth() - n); d.setDate(d.getDate() + 1); return iso(d); };
    const presets = [
        { key: 'today', short: L.todayShort, label: L.today, from: today, to: today },
        { key: '7d', short: '7D', label: L.last7, from: daysAgo(6), to: today },
        { key: '30d', short: '30D', label: L.last30, from: daysAgo(29), to: today },
        { key: '3m', short: '3M', label: L.last3m, from: monthsAgo(3), to: today },
        { key: '6m', short: '6M', label: L.last6m, from: monthsAgo(6), to: today },
        { key: 'ytd', short: 'YTD', label: L.ytd, from: now.getFullYear() + '-01-01', to: today },
    ];
    // week starts on monday, 2024-01-01 was a monday
    const weekdays = Array.from({ length: 7 }, (_, i) => dayFmt.format(new Date(2024, 0, 1 + i)));

    return {
        open: false,
        popShift: 0,
        from: from || '',
        to: to || from || '',
        draftFrom: '',
        draftTo: '',
        hover: '',
        view: null,
        today,
        weekdays,
        presets,
        quickPresets: presets.slice(0, 5),

        init() { this.resetDraft(); },
        resetDraft() {
            this.draftFrom = this.from;
            this.draftTo = this.to;
            this.hover = '';
            // right calendar shows the range end (or this month), left one the month before
            const end = this.to ? parse(this.to) : new Date(now);
            this.view = new Date(end.getFullYear(), end.getMonth() - 1, 1);
        },
        toggle() {
            if (this.open) return this.close();
            this.resetDraft();
            this.popShift = 0;
            this.open = true;
            // slide it left just enough to fit, phones get a fixed bottom sheet instead
            this.$nextTick(() => {
                if (window.innerWidth < 768) return;
                const r = this.$refs.pop.getBoundingClientRect();
                this.popShift = Math.max(0, Math.min(r.right - (window.innerWidth - 16), r.left - 16));
            });
        },
        close() { this.open = false; },

        rangeLabel(a, b) {
            if (!a) return '';
            if (!b || a === b) return fmt.format(parse(a));
            return fmt.format(parse(a)) + ' – ' + fmt.format(parse(b));
        },
        longLabel(d) { return longFmt.format(parse(d)); },
        isCurrent(p) { return this.from === p.from && this.to === p.to; },

        months() {
            return [0, 1].map(offset => {
                const first = new Date(this.view.getFullYear(), this.view.getMonth() + offset, 1);
                const lead = (first.getDay() + 6) % 7;
                const count = new Date(first.getFullYear(), first.getMonth() + 1, 0).getDate();
                const cells = [];
                for (let i = 0; i < lead; i++) cells.push({ key: 'b' + i, date: '', day: '' });
                for (let d = 1; d <= count; d++) {
                    const date = iso(new Date(first.getFullYear(), first.getMonth(), d));
                    cells.push({ key: date, date, day: d });
                }
                return { key: iso(first), title: monthFmt.format(first), cells };
            });
        },
        canGoNext() {
            const nextRight = new Date(this.view.getFullYear(), this.view.getMonth() + 2, 1);
            return iso(nextRight) <= today;
        },
        shift(n) {
            if (n > 0 && !this.canGoNext()) return;
            this.view = new Date(this.view.getFullYear(), this.view.getMonth() + n, 1);
        },

        pickDay(d) {
            if (!this.draftFrom || this.draftTo) {
                this.draftFrom = d;
                this.draftTo = '';
            } else if (d < this.draftFrom) {
                this.draftTo = this.draftFrom;
                this.draftFrom = d;
            } else {
                this.draftTo = d;
            }
        },
        pickPreset(p) {
            this.draftFrom = p.from;
            this.draftTo = p.to;
            const end = parse(p.to);
            this.view = new Date(end.getFullYear(), end.getMonth() - 1, 1);
        },
        cellClass(c) {
            if (!c.date) return '';
            // while picking the end, preview the range up to the hovered day
            let a = this.draftFrom, b = this.draftTo || (this.draftFrom && this.hover) || this.draftFrom;
            if (a && b && b < a) [a, b] = [b, a];
            return {
                'is-start': c.date === a,
                'is-end': c.date === b,
                'is-in-range': a && b && c.date > a && c.date < b,
                'is-today': c.date === today,
                'is-single': a && c.date === a && a === b,
            };
        },

        apply() {
            if (!this.draftFrom) return this.go('', '');
            this.go(this.draftFrom, this.draftTo || this.draftFrom);
        },
        go(a, b) {
            const params = new URLSearchParams(window.location.search);
            params.delete('page');
            a ? params.set('from', a) : params.delete('from');
            b ? params.set('to', b) : params.delete('to');
            // pages that compare against exact timestamps need the browser's day boundaries
            if (opts.tz) a ? params.set('tz', String(-new Date().getTimezoneOffset())) : params.delete('tz');
            const qs = params.toString();
            window.location.href = window.location.pathname + (qs ? '?' + qs : '');
        },
    };
}
