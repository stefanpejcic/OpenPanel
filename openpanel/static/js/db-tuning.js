// dbTuning drives the optimization banner, the per setting buttons and the confirm dialog on the MySQL and PostgreSQL configuration pages
function dbTuning(url) {
    return {
        searchQuery: '',
        loading: true,
        error: '',
        recs: [],
        facts: {},
        notes: [],
        open: false,
        selected: [],
        applied: {},
        byKey(key) {
            return this.recs.find(r => r.key === key);
        },
        async load() {
            try {
                const res = await fetch(url, { headers: { 'Accept': 'application/json' } });
                const data = await res.json();
                if (!res.ok) throw new Error(data.error || res.statusText);
                this.recs = data.recommendations || [];
                this.facts = data.facts || {};
                this.notes = data.notes || [];
            } catch (e) {
                this.error = e.message;
            } finally {
                this.loading = false;
            }
        },
        review(key) {
            this.selected = key ? this.recs.filter(r => r.key === key) : this.recs.slice();
            this.open = true;
        },
        applySelected() {
            this.selected.forEach(r => {
                const input = document.getElementById(r.key);
                if (input) input.value = r.recommended;
                this.applied[r.key] = true;
            });
            this.open = false;
        },
        save() {
            this.applySelected();
            this.$refs.form.submit();
        }
    };
}
