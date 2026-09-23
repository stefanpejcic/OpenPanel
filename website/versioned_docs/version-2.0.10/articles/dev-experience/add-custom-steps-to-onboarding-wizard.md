# Add Custom Steps to the Onboarding Wizard

OpenPanel's [Onboarding wizard](/docs/panel/dashboard/onboarding) walks new users through the features already built into the panel - webserver, PHP, database, Varnish, backups, and security. If you want to guide users through setting up something that isn't part of that list - for example, enabling Redis as a cache - you can register an extra step from custom JavaScript, without patching or forking OpenPanel.

## How it works

The wizard reads a global array, `window.OPOnboardingSteps`, when it initializes, and adds one wizard step per entry it finds there. Each entry is a plain object:

| Key | Required | Description |
| --- | --- | --- |
| `key` | Yes | A unique string identifying the step. |
| `render(host, wizard)` | Yes | Called to draw the step. `host` is an empty `<div>` to fill with your markup; `wizard` is the onboarding wizard's Alpine component instance. |
| `title` | No | Label shown in the wizard's step tabs. Falls back to `key`. |
| `enabled(opts)` | No | Return `false` to hide the step for a given user. `opts` mirrors the same feature flags the built-in steps use (`opts.php`, `opts.mysql`, etc.) - use it if you want your step gated on an existing feature. |
| `init(wizard)` | No | Called once, when the wizard opens, before the user reaches your step - use it to prefetch current state (e.g. whether a service is already running). Can be async. |
| `submit(wizard)` | No | Called when the user clicks **Continue** on your step. Return `false` (or reject) to keep the user on the step with `wizard.error` set to a message; return `true` to advance. Can be async. If omitted, the step always advances. |

Your step appears right before the final "Finish" screen, after Security.

## Where to add the code

Edit (or create) the custom JS override file:

```bash
nano /etc/openpanel/openpanel/static/js/custom.js
```

This file is always loaded by the OpenPanel interface, so anything you put there runs on every page - not just onboarding. If the file didn't already exist, restart the `openpanel` container once so it picks up the new override:

```bash
cd /root && docker compose up -d openpanel
```

Further edits to an existing `custom.js` take effect immediately on page reload - no restart needed.

## Example: add a Redis step

Redis already has a full page in OpenPanel (`/cache/redis`) with a same-origin JSON API the panel's own pages use (`GET`/`POST /cache/redis?output=json`), but it isn't one of the wizard's built-in steps. Here's a complete step that lets users enable or disable it during onboarding:

```js
// Adds a "Redis" step to the account-setup onboarding wizard.
window.OPOnboardingSteps = window.OPOnboardingSteps || [];

window.OPOnboardingSteps.push({
    key: 'redis',
    title: 'Redis',

    // Prefill from the current container state before the wizard shows it.
    async init(wizard) {
        const { ok, data } = await wizard.getJSON('/cache/redis?output=json');
        this._choice = ok && data.container_state === 'running' ? 'enabled' : 'disabled';
    },

    render(host, wizard) {
        const choice = this._choice || 'disabled';
        host.innerHTML = `
            <p class="text-sm text-gray-500 dark:text-gray-400 mb-3">Enable Redis as an in-memory cache for your applications.</p>
            <div class="space-y-2">
                <label class="flex items-center gap-3 rounded border p-3 cursor-pointer border-gray-200 dark:border-gray-700">
                    <input type="radio" name="onb_redis" value="enabled" ${choice === 'enabled' ? 'checked' : ''}>
                    <span class="text-sm font-medium text-gray-900 dark:text-gray-50">Enabled</span>
                </label>
                <label class="flex items-center gap-3 rounded border p-3 cursor-pointer border-gray-200 dark:border-gray-700">
                    <input type="radio" name="onb_redis" value="disabled" ${choice !== 'enabled' ? 'checked' : ''}>
                    <span class="text-sm font-medium text-gray-900 dark:text-gray-50">Disabled</span>
                </label>
            </div>`;
        host.querySelectorAll('input[name="onb_redis"]').forEach(input => {
            input.addEventListener('change', () => { this._choice = input.value; });
        });
    },

    async submit(wizard) {
        const action = this._choice === 'enabled' ? 'enable' : 'disable';
        const { ok, data } = await wizard.formPost('/cache/redis?output=json', { action });
        if (!ok) { wizard.error = (data && data.error) || 'Failed to update Redis.'; return false; }
        return true;
    }
});
```

A few things worth noting about the example:

- `wizard.getJSON(url)` and `wizard.formPost(url, params)` are the same helpers the built-in steps use to talk to OpenPanel's own session-authenticated page routes (not the Bearer-token `/api/*` routes) - they already attach the CSRF token, so you don't need to handle that yourself.
- Setting `wizard.error` reuses the wizard's existing error banner, shown above the step content.
- `this` inside the step definition refers to the object you pushed onto `window.OPOnboardingSteps`, so it's a convenient place to stash per-step state (like `_choice` above) between `init`, `render`, and `submit`.

## Gating a custom step

To only show your step for users who have a specific feature enabled, add an `enabled` function. For example, to only offer the Redis step to users who also have the (built-in) Varnish step:

```js
enabled(opts) {
    return !!opts.varnish;
}
```

If you need a condition that isn't one of the wizard's existing `opts` flags, call your own endpoint from `enabled` or `init` and store the result before deciding.

## Testing

The wizard only shows automatically to a user with no domains and no databases yet, on their first login. To see your changes without creating a throwaway account, delete the onboarding marker file for a test user and reload the dashboard - see [resetting onboarding for a user](/docs/admin/settings/openpanel/#onboarding).
