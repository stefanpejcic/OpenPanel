# API

:::info
API access is available only on [OpenPanel Enterprise edition](https://openpanel.com/beta/)
:::

## Enable

To enable OpenAdmin API access:

```bash
opencli api on
```

<details>
  <summary>Example output</summary>

```bash
# opencli api on
Updated api to on
```
</details>

## Disable

To disable OpenAdmin API access:

```bash
opencli api off
```

<details>
  <summary>Example output</summary>

```bash
# opencli api off
Updated api to off
```
</details>

## Status

To check current OpenAdmin API status:

```bash
opencli api status
```

<details>
  <summary>Example output</summary>

```bash
# opencli api status
on
```
</details>

## API list

`opencli api list` displays all API endpoints of the installed OpenAdmin version, with the access they need and a curl example for each. The list is built into the OpenAdmin binary from the same source as the API docs on **OpenAdmin > Settings > API**, and the examples use the OpenAdmin URL from `opencli admin`.

```bash
opencli api list
```

<details>
  <summary>Example output</summary>

```bash
# opencli api list
# Authentication

Endpoint: GET /api/
Description: Health check - confirms the API is reachable.
Access: No authentication.
Example:
  curl -X GET 'https://server.example.com:2087/api/'

--------------------------------------------------------------------------------

Endpoint: POST /api/
Description: Log in with username/password and obtain a JWT Bearer token, valid for 15 minutes.
Access: No authentication.
Example:
  curl -X POST 'https://server.example.com:2087/api/' -H "Content-Type: application/json" -d '{"username":"admin","password":"secret"}'

--------------------------------------------------------------------------------

Endpoint: GET /api/whoami
Description: Returns the username embedded in the caller's JWT access token. Reads the identity directly from the token claims rather than the database, so it still answers even if the user row backing the token has since been deleted.
Access: Any authenticated user.
Example:
  curl -X GET 'https://server.example.com:2087/api/whoami' -H "Authorization: Bearer TOKEN_HERE"

--------------------------------------------------------------------------------
...
```
</details>
