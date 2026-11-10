// row action menus open fixed next to their button, so table scroll boxes can't clip them
document.addEventListener('alpine:init', () => {
    Alpine.data('rowMenu', () => ({
        open: false,
        pos: '',
        toggle() { this.open ? this.close() : this.show(); },
        show() {
            this.pos = 'visibility:hidden';
            this.open = true;
            this.$nextTick(() => {
                const b = this.$refs.btn.getBoundingClientRect();
                const m = this.$refs.menu;
                const h = m.offsetHeight, w = m.offsetWidth;
                // flip up when there's no room below
                const up = b.bottom + 4 + h > window.innerHeight && b.top - 4 - h > 0;
                const top = up ? b.top - 4 - h : b.bottom + 4;
                const left = Math.max(8, Math.min(b.right - w, window.innerWidth - w - 8));
                this.pos = `position:fixed;top:${top}px;left:${left}px;right:auto;margin:0`;
            });
        },
        close() { this.open = false; },
        init() {
            const c = () => { if (this.open) this.close(); };
            window.addEventListener('scroll', c, true);
            window.addEventListener('resize', c);
        },
    }));
});
