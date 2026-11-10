// two-click delete for table rows: the first click arms the button for 5 seconds, a second click submits its form
// btn-danger / btn-dark are only hooks, the upstream e2e specs click them
(function () {
    const ARMED = ['btn-dark', 'bg-red-50', 'text-red-600', 'dark:bg-red-400/10', 'dark:text-red-400'];

    window.confirmDelete = function (button) {
        if (button.dataset.armed) {
            button.closest('form').submit();
            return;
        }
        const idleHTML = button.innerHTML;
        const idleTitle = button.title;
        let left = 5;
        const draw = () => {
            button.innerHTML = '<i class="ri-delete-bin-fill text-base leading-none" aria-hidden="true"></i><span class="text-xs font-medium tabular-nums">' + left + '</span>';
        };

        button.dataset.armed = '1';
        button.classList.remove('btn-danger');
        button.classList.add(...ARMED);
        button.title = button.dataset.confirmLabel || 'Confirm';
        draw();

        const timer = setInterval(() => {
            left--;
            if (left > 0) return draw();
            clearInterval(timer);
            delete button.dataset.armed;
            button.classList.remove(...ARMED);
            button.classList.add('btn-danger');
            button.title = idleTitle;
            button.innerHTML = idleHTML;
        }, 1000);
    };
})();
