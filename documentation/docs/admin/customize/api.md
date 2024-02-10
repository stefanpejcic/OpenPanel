---
sidebar_position: 9
---

# API

:::danger
THIS FEATURE IS STILL EXPERIMENTAL AND NOT YET PRODUCTION READY!
:::

## Basic Authentication

OpenPanel API uses JWT tokens for authentication.

### Generate access token

Send POST request to /api with your username and password:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"username":"USERNAME_HERE", "password":"PASSWORD_HERE"}' https://OPENPANEL:2087/api
```
Example response:
  
```json 
{
    "name": "string"
}
```

### Check if token is valid

```bash
curl -H "Authorization: Bearer JWT_TOKEN" https://OPENPANEL:2087/api/whoami
```

Example response:
```json 
{
  "logged_in_as": "user"
}
```

## Users

### List All Users

```bash
curl -H "Authorization: Bearer JWT_TOKEN" https://OPENPANEL:2087/api/users
```

Example response:
```json
{
{
  "users": [
    [
      12, 
      "SUSPENDED_20240103122300_jasta", 
      "pbkdf2:sha256:260000$G26Nw1zBQ6yywbiG$6a0d321490842b58068ad44e6e21b96e89f42cb6daecebe3afc06ad447ef3785", 
      "jasta", 
      "1,2,3,4,5,6,7,8,9,10,11,12", 
      "", 
      0, 
      null, 
      null, 
      "Tue, 02 Jan 2024 19:44:25 GMT", 
      2, 
      "cloud_4_nginx"
    ], 
    [
      13, 
      "example_user", 
      "pbkdf2:sha256:260000$EIkp3AkT9OBPDtr1$956594ccc6aaa2c24c528f98eacadee7ac35193578b32dc415b2402fa9a22595", 
      "user@example.com", 
      "1,2,3,4,5,6,7,8,9,10,11,12", 
      "", 
      0, 
      null, 
      null, 
      "Tue, 09 Jan 2024 14:18:56 GMT", 
      2, 
      "cloud_4_apache"
    ]
  ]
}
}
```

### List Single User

```bash
curl -H "Authorization: Bearer JWT_TOKEN" https://OPENPANEL:2087/api/users/stefan
```

Example response:
```json
{
{
  "users": [
    [
      15, 
      "stefan", 
      "pbkdf2:sha256:260000$EIkp3AkT9OBPDtr1$956594ccc6aaa2c24c528f98eacadee7ac35193578b32dc415b2402fa9a22595", 
      "user@example.com", 
      "1,2,3,4,5,6,7,8,9,10,11,12", 
      "", 
      0, 
      null, 
      null, 
      "Tue, 09 Jan 2024 14:18:56 GMT", 
      2, 
      "cloud_4_apache"
    ]
  ]
}
}
```

### Create User

