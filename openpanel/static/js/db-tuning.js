// dbTuning drives the optimization banner, the per setting buttons and the confirm dialog on the MySQL, PostgreSQL and PHP options pages
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
        // fills a suggested value into whatever control the page uses for that key
        setField(key, value) {
            const combined = document.querySelector(`.combined-input[data-key="${CSS.escape(key)}"]`);
            if (combined) {
                const m = String(value).match(/^(\d+(?:\.\d+)?)\s*([KMG])?$/i);
                if (m) {
                    combined.querySelector('.numeric-part').value = m[1];
                    combined.querySelector('.unit-part').value = (m[2] || 'M').toUpperCase();
                    combined.querySelector('.numeric-part').dispatchEvent(new Event('input', { bubbles: true }));
                }
                return;
            }
            const field = document.getElementById(key) || document.querySelector(`[name="${CSS.escape(key)}"]`);
            if (!field) return;
            if (field.type === 'checkbox') {
                field.checked = ['on', '1', 'yes', 'true'].includes(String(value).toLowerCase());
            } else {
                field.value = value;
            }
            field.dispatchEvent(new Event('change', { bubbles: true }));
        },
        applySelected() {
            this.selected.forEach(r => {
                this.setField(r.key, r.recommended);
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
