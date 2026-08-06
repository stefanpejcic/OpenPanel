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
            }

            this.$watch('columns', () => {
                this.saveColumns();
            }, { deep: true });
        },

        updateCount() {
            this.count = [...this.$refs.tbody.querySelectorAll('tr.user-row')]
                .filter(row => row.offsetParent !== null).length;
        },

        saveColumns() {
            store.setItem(storageKey, JSON.stringify(this.columns));
        },

        loadColumns() {
            const saved = store.getItem(storageKey);
            if (saved) {
                try {
                    this.columns = JSON.parse(saved);
                } catch (e) {
                    console.warn(`Failed to parse ${storageKey} from ${storage}Storage`, e);
                }
            }
        }
    };
}
