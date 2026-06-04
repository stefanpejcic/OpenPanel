# Usage

##  Prepare OpenPanel server:

1. Install/update OpenPanel on the server: `bash <(curl -sSL https://openpanel.org)` | `opencli update --beta`
2. On OpenPanel server run:
   
   ```bash
   bash <(curl -sSL https://raw.githubusercontent.com/stefanpejcic/openpanel-tests/refs/heads/main/openpanel/prepare.sh)
   ```

## Prepare Playwright tests:

1. Download all tests: `cd ~/playwright-test && git pull`
2. Add logins to `/root/playwright-test/openpanel/.env`:
   ```bash
   BASE_URL=
   PANEL_USERNAME=
   PANEL_PASSWORD=
   ```
4. Run tests:   
   ```
   cd /root/playwright-test && npx playwright test -c openpanel/playwright.config.ts --project=tests --project=tests --ui
   ```
