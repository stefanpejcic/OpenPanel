// passkeys.js - WebAuthn client helpers (registration + usernameless login)

function base64urlToBuffer(base64url) {
    const padding = '='.repeat((4 - (base64url.length % 4)) % 4);
    const base64 = (base64url + padding).replace(/-/g, '+').replace(/_/g, '/');
    const raw = atob(base64);
    const buffer = new Uint8Array(raw.length);
    for (let i = 0; i < raw.length; i++) buffer[i] = raw.charCodeAt(i);
    return buffer;
}

function bufferToBase64url(buffer) {
    const bytes = new Uint8Array(buffer);
    let str = '';
    for (let i = 0; i < bytes.byteLength; i++) str += String.fromCharCode(bytes[i]);
    return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

async function registerPasskey(name) {
    const beginResp = await fetch('/account/passkeys/register/begin', {
        method: 'POST',
        headers: { 'X-CSRF-Token': csrf_token }
    });
    if (!beginResp.ok) throw new Error('Failed to start passkey registration.');
    const options = await beginResp.json();

    options.challenge = base64urlToBuffer(options.challenge);
    options.user.id = base64urlToBuffer(options.user.id);
    if (options.excludeCredentials) {
        options.excludeCredentials = options.excludeCredentials.map(c => ({
            ...c,
            id: base64urlToBuffer(c.id)
        }));
    }

    const credential = await navigator.credentials.create({ publicKey: options });

    const payload = {
        id: credential.id,
        rawId: bufferToBase64url(credential.rawId),
        type: credential.type,
        response: {
            clientDataJSON: bufferToBase64url(credential.response.clientDataJSON),
            attestationObject: bufferToBase64url(credential.response.attestationObject),
            transports: credential.response.getTransports ? credential.response.getTransports() : []
        },
        clientExtensionResults: credential.getClientExtensionResults ? credential.getClientExtensionResults() : {}
    };

    const completeResp = await fetch(`/account/passkeys/register/complete?name=${encodeURIComponent(name || 'Passkey')}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrf_token },
        body: JSON.stringify(payload)
    });
    const result = await completeResp.json();
    if (!completeResp.ok || result.error) throw new Error(result.error || 'Failed to save passkey.');
    return result;
}

async function loginWithPasskey() {
    const beginResp = await fetch('/login/passkey/begin', {
        method: 'POST',
        headers: { 'X-CSRF-Token': csrf_token }
    });
    if (!beginResp.ok) throw new Error('Failed to start passkey login.');
    const options = await beginResp.json();

    options.challenge = base64urlToBuffer(options.challenge);
    if (options.allowCredentials) {
        options.allowCredentials = options.allowCredentials.map(c => ({
            ...c,
            id: base64urlToBuffer(c.id)
        }));
    }

    const credential = await navigator.credentials.get({ publicKey: options });

    const payload = {
        id: credential.id,
        rawId: bufferToBase64url(credential.rawId),
        type: credential.type,
        response: {
            clientDataJSON: bufferToBase64url(credential.response.clientDataJSON),
            authenticatorData: bufferToBase64url(credential.response.authenticatorData),
            signature: bufferToBase64url(credential.response.signature),
            userHandle: credential.response.userHandle ? bufferToBase64url(credential.response.userHandle) : null
        },
        clientExtensionResults: credential.getClientExtensionResults ? credential.getClientExtensionResults() : {}
    };

    const completeResp = await fetch('/login/passkey/complete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrf_token },
        body: JSON.stringify(payload)
    });
    const result = await completeResp.json();
    if (!completeResp.ok || result.error) throw new Error(result.error || 'Passkey login failed.');
    return result;
}
