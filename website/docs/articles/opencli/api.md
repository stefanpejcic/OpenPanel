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

`opencli api list` displays all available api endpoints for the installed OpenAdmin version.

```bash
opencli api list
```

<details>
  <summary>Example output</summary>

```bash
# opencli api list
Endpoint: /api/
Description: On GET returns api status, on POST returns access token
Type: GET POST
Examples:
  curl -X GET http://localhost:2087/api/
  curl -X POST http://localhost:2087/api/ -H "Content-Type: application/json" -d '{"username":"admin", "password":"kQsUFhwkzBCw3M57"}'

--------------------------------------------------------------------------------

Endpoint: /api/whoami
Description: protected route that can be accessed only with a token and returns a username
Type: GET
Examples:

--------------------------------------------------------------------------------
...
```
</details>

