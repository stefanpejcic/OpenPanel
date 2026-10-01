// Shared Alpine.js controller for the search/column-visibility/pagination
// tables used across domains, containers, filemanager and trash pages.
// Each page passes its own storage key + column defaults; everything else
// (persisting column visibility, live row count) is identical between pages.
function createTableController({ storageKey, storage = 'local', columns = {}, trackCount = false, initialCount = 0 } = {}) {
    const store = storage === 'session' ? sessionStorage : localStorage;

    return {
        searchQuery: '',
        columns: { ...columns },
        ...(trackCount ? { count: initialCount } : {}),

        init() {
            this.loadColumns();

            if (trackCount) {
                this.$watch('searchQuery', () => {
                    this.$nextTick(() => this.updateCount());
                });
                // pages can mix in their own filters object, rows hidden by it shouldn't count
                if (this.filters) {
                    this.$watch('filters', () => {
                        this.$nextTick(() => this.updateCount());
                    });
                }
            }

            this.$watch('columns', () => {
                this.saveColumns();
            }, { deep: true });
        },

        updateCount() {
            // the bulk wrapper is its own x-data, so a tbody inside it isn't in our $refs
            const tbody = this.$refs.tbody || this.$root.querySelector('[x-ref="tbody"]');
            this.count = [...tbody.querySelectorAll('tr.user-row')]
                .filter(row => row.offsetParent !== null).length;
        },

        saveColumns() {
            store.setItem(storageKey, JSON.stringify(this.columns));
        },

        loadColumns() {
            const saved = store.getItem(storageKey);
            if (saved) {
                try {
                    // merge so columns added after the user saved their prefs still get their default
                    this.columns = { ...this.columns, ...JSON.parse(saved) };
                } catch (e) {
                    console.warn(`Failed to parse ${storageKey} from ${storage}Storage`, e);
                }
            }
        }
    };
}
