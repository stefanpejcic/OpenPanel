// shared checkbox selection + bulk-actions bar, markup lives in templates/partials/_bulk.html
function bulkActions(cfg) {
    return {
        bulkSelected: {},
        bulkRunning: false,
        bulkConfirming: null,
        bulkPending: [],
        bulkSkipped: 0,
        bulkValue: '',
        // an object, a string :style would overwrite x-show's display:none
        bulkBarStyle: {},

        // fixed to the bottom of the screen but only as wide as main, so it doesn't cover the sidebar
        bulkPlaceBar() {
            const main = this.$root.closest('main') || document.querySelector('main');
            if (!main) return;
            const place = () => {
                const rect = main.getBoundingClientRect();
                this.bulkBarStyle = { left: rect.left + 'px', right: Math.max(0, window.innerWidth - rect.right) + 'px' };
            };
            place();
            // follows window resizes and the sidebar collapsing
            if (!this._bulkObserver && window.ResizeObserver) {
                this._bulkObserver = new ResizeObserver(place);
                this._bulkObserver.observe(main);
            }
        },

        // only rows the search hasn't hidden
        bulkBoxes() {
            return Array.from(document.querySelectorAll(cfg.table + ' tbody input.bulk-select-box')).filter(cb => cb.offsetParent !== null);
        },
        get bulkCount() { return Object.keys(this.bulkSelected).length; },
        get bulkAllChecked() {
            const boxes = this.bulkBoxes();
            return boxes.length > 0 && boxes.every(cb => this.bulkSelected[cb.value]);
        },
        get bulkSomeChecked() {
            return this.bulkBoxes().some(cb => this.bulkSelected[cb.value]) && !this.bulkAllChecked;
        },
        bulkToggle(cb) {
            if (cb.checked) {
                this.bulkSelected[cb.value] = (cb.dataset.bulkSkip || '').split(',').filter(Boolean);
            } else {
                delete this.bulkSelected[cb.value];
            }
        },
        bulkToggleAll(checked) {
            this.bulkBoxes().forEach(cb => {
                cb.checked = checked;
                this.bulkToggle(cb);
            });
        },
        bulkClear() {
            this.bulkSelected = {};
            this.bulkConfirming = null;
            document.querySelectorAll(cfg.table + ' tbody input.bulk-select-box').forEach(cb => { cb.checked = false; });
        },
        bulkStart(action, initial) {
            if (this.bulkRunning) return;
            const keys = Object.keys(this.bulkSelected);
            this.bulkPending = keys.filter(k => !this.bulkSelected[k].includes(action) && !this.bulkSelected[k].includes('*'));
            this.bulkSkipped = keys.length - this.bulkPending.length;
            if (this.bulkPending.length === 0) {
                showToast(cfg.messages.none, 'warning');
                return;
            }
            this.bulkValue = initial || '';
            this.bulkConfirming = action;
            this.$nextTick(() => {
                const input = this.$root.querySelector('[data-bulk-input="' + action + '"]');
                if (input) input.focus();
            });
        },
        bulkCancel() {
            if (!this.bulkRunning) this.bulkConfirming = null;
        },
        bulkConfirm() {
            const action = this.bulkConfirming;
            if (!action || this.bulkRunning || this.bulkPending.length === 0) return;
            const input = this.$root.querySelector('[data-bulk-input="' + action + '"]');
            if (input && !input.reportValidity()) return;

            this.bulkRunning = true;
            fetch(cfg.endpoint, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrf_token },
                body: JSON.stringify({ action: action, value: String(this.bulkValue), items: this.bulkPending })
            })
                .then(response => {
                    if (!response.ok) throw new Error('HTTP ' + response.status);
                    return response.json();
                })
                // the server flashes a summary banner, so just reload to show it
                .then(() => window.location.reload())
                .catch(error => {
                    console.error('Bulk action error:', error);
                    showToast(cfg.messages.error, 'error');
                    this.bulkRunning = false;
                    this.bulkConfirming = null;
                });
        }
    };
}
