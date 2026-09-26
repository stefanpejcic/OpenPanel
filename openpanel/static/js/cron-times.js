// run times for the cron pages, formatted in the user's language and in the cron service's time zone
function cronTimes(locale, timeZone) {
    let lang = (locale || navigator.language || 'en').replace('_', '-');
    // the panel's Serbian is in Latin script, Intl defaults sr to Cyrillic
    if (/^sr(-|$)/.test(lang) && !/Cyrl/.test(lang)) lang = 'sr-Latn';
    const tz = timeZone || 'UTC';
    let day, clock, clockSec, rel;
    try {
        day = new Intl.DateTimeFormat(lang, { weekday: 'short', day: 'numeric', month: 'short', timeZone: tz });
        clock = new Intl.DateTimeFormat(lang, { hour: '2-digit', minute: '2-digit', hourCycle: 'h23', timeZone: tz });
        clockSec = new Intl.DateTimeFormat(lang, { hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23', timeZone: tz });
        rel = new Intl.RelativeTimeFormat(lang, { numeric: 'auto' });
    } catch (e) {
        return cronTimes('en', tz);
    }
    return {
        day: d => day.format(new Date(d)),
        // seconds only matter for schedules that don't run on the minute, pass seconds=true to keep a list aligned
        time: (d, seconds) => ((seconds ?? new Date(d).getUTCSeconds()) ? clockSec : clock).format(new Date(d)),
        // "in 45 seconds", "in 16 hours", "in 3 days"
        from: (d, now) => {
            const s = Math.round((new Date(d) - (now || new Date())) / 1000);
            if (Math.abs(s) < 60) return rel.format(s, 'second');
            if (Math.abs(s) < 3600) return rel.format(Math.round(s / 60), 'minute');
            if (Math.abs(s) < 86400) return rel.format(Math.round(s / 3600), 'hour');
            return rel.format(Math.round(s / 86400), 'day');
        },
    };
}
