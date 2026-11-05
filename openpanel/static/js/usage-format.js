// usage pages get sizes like "670.9M" and ISO timestamps from the server, show them as "670.9 MB" and a local date
(function () {
    const withUnit = text => text
        .replace(/(\d)\s*([KMGTP])(?:i?B)?\b/g, '$1 $2B')
        .replace(/(\d)\s*bit\b/g, '$1 bit');

    document.querySelectorAll('[data-size]').forEach(el => {
        el.textContent = withUnit(el.textContent);
    });

    document.querySelectorAll('time[datetime]').forEach(el => {
        const date = new Date(el.getAttribute('datetime'));
        if (isNaN(date)) return;
        el.title = el.getAttribute('datetime');
        el.textContent = date.toLocaleString(undefined, { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
    });
})();
