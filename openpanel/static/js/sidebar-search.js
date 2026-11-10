// sidebar menu search: hides links that don't match and opens the groups that still have some
(function () {
    function filter(q) {
        q = (q || '').trim().toLowerCase();
        const nav = document.getElementById('sidebar');
        if (!nav) return;
        // groups keep their own open state and just show their links while there's a query
        window.dispatchEvent(new CustomEvent('sidebar-search', { detail: q }));
        const has = (el) => !q || el.textContent.toLowerCase().includes(q);
        let any = false;
        nav.querySelectorAll('[data-nav-group]').forEach((li) => {
            const head = li.querySelector(':scope > button');
            const groupHit = head && has(head);
            let shown = 0;
            li.querySelectorAll(':scope > ul li').forEach((sub) => {
                const hit = groupHit || has(sub);
                sub.hidden = !hit;
                if (hit) shown++;
            });
            li.hidden = !groupHit && !shown;
            if (!li.hidden && q) any = true;
        });
        nav.querySelectorAll('[data-nav-item]').forEach((li) => {
            li.hidden = !has(li);
            if (!li.hidden && q) any = true;
        });
        nav.querySelectorAll('#favorites-list li').forEach((li) => {
            li.hidden = !has(li);
            if (!li.hidden && q) any = true;
        });
        // section labels only stay when something under them still shows
        nav.querySelectorAll('[data-nav-label]').forEach((label) => {
            let next = label.nextElementSibling, visible = false;
            while (next && !next.hasAttribute('data-nav-label')) { if (!next.hidden) visible = true; next = next.nextElementSibling; }
            label.hidden = !visible;
        });
        const empty = document.getElementById('sidebar-search-empty');
        if (empty) empty.hidden = !q || any;
    }

    // Enter opens the first link still showing
    function first() {
        const nav = document.getElementById('sidebar');
        const link = nav && Array.from(nav.querySelectorAll('[data-sidebar="content"] a[href]')).find((a) => a.offsetParent && a.getAttribute('href') !== '#');
        if (link) window.location.href = link.getAttribute('href');
    }

    window.sidebarSearch = { filter: filter, first: first };
})();
