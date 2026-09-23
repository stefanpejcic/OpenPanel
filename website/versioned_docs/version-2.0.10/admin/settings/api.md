---
sidebar_position: 5
---

# API Access

:::info
The API Access page is only available with an **Enterprise** license. It is also disabled while HTTP Basic Authentication is enabled for OpenAdmin — disable Basic Authentication first to use API Access.
:::

Use the API Access page to browse every API endpoint, test API calls, and see request and response examples.

![API Access page showing API access as disabled with the Enable API access button](/img/openadmin-screenshots/settings/api-page.png#gh-light-mode-only)
![API Access page showing API access as disabled with the Enable API access button](/img/openadmin-screenshots/settings/api-page_dark.png#gh-dark-mode-only)

To begin:

1. **Enable API Access**

   Click **Enable API access** to turn on the API and the interactive reference.

2. **Browse the API reference**

   Once enabled, the page shows the interactive OpenAdmin API reference, with every endpoint grouped by section (Authentication, Users, Domains, Plans, Containers, DNS, Email, and more). Use **Download spec** to download the OpenAPI specification file, or **Disable API access** to turn the API off again.

   ![API Access page with API access enabled, the Download spec and Disable API access buttons, and the interactive OpenAdmin API reference](/img/openadmin-screenshots/settings/api_server-enabled.png#gh-light-mode-only)
   ![API Access page with API access enabled, the Download spec and Disable API access buttons, and the interactive OpenAdmin API reference](/img/openadmin-screenshots/settings/api_server-enabled_dark.png#gh-dark-mode-only)

3. **Authorize**

   Most endpoints require a Bearer token. Expand **POST /api/** in the Authentication section, click **Try it out**, enter an admin username and password, and click **Execute** to get a token. Then click **Authorize**, paste the token, and confirm. Tokens expire after 15 minutes.

4. **Send a request**

   Expand any endpoint, click **Try it out**, fill in its parameters, and click **Execute**.

5. **View the response and curl command**

   Below the request you'll see the equivalent curl command, the request URL, and the server response with its status code, body and headers.

   ![An API endpoint expanded in the API reference after Try it out and Execute, with the curl command, request URL and server response](/img/openadmin-screenshots/settings/api_server-try.png#gh-light-mode-only)
   ![An API endpoint expanded in the API reference after Try it out and Execute, with the curl command, request URL and server response](/img/openadmin-screenshots/settings/api_server-try_dark.png#gh-dark-mode-only)

For full API reference and additional endpoints, please refer to [OpenAdmin API Specification](/docs/articles/dev-experience/openadmin-api).
