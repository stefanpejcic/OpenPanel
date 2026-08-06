// Copy text to the clipboard. navigator.clipboard is only available in
// secure contexts (HTTPS/localhost), so plain HTTP installs need the
// execCommand fallback below. Always returns a Promise.
window.copyToClipboard = function (text) {
    if (navigator.clipboard && window.isSecureContext) {
        return navigator.clipboard.writeText(text);
    }

    return new Promise((resolve, reject) => {
        const textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.top = '0';
        textarea.style.left = '0';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.focus();
        textarea.select();
        try {
            document.execCommand('copy') ? resolve() : reject(new Error('execCommand copy failed'));
        } catch (err) {
            reject(err);
        } finally {
            document.body.removeChild(textarea);
        }
    });
};
