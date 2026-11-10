// fields filled before Alpine starts (a fast typist, autofill, a test) would be reset by x-model, so a component's init() hands them back
window.keepFilled = function (c) {
    const root = c.$el;
    root.querySelectorAll('input, select, textarea').forEach(el => {
        if (el.type === 'checkbox' || el.type === 'radio' || !el.value) return;
        // nested components keep their own fields
        if (el.closest('[x-data]') !== root) return;
        const attr = Array.from(el.attributes).find(a => a.name === 'x-model' || a.name.startsWith('x-model.'));
        if (!attr) return;
        const key = attr.value.trim();
        if (key in c && !c[key]) c[key] = el.value;
    });
};
