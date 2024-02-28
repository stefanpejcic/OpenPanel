---
title: OpenPanel BETA is out!
description: Public beta is released!
slug: openpanel-public-beta-released
authors: yildiray
tags: [OpenPanel, news, tutorial]
image: https://refine.ams3.cdn.digitaloceanspaces.com/website/static/img/placeholder.png
hide_table_of_contents: false
---

:::caution

Something
:::

Link to blog..

<!--truncate-->

All the steps described are in this [repo](https://github.com/refinedev/refine/tree/master/examples/blog-job-posting).

## Intro

[NestJS](https://github.com/nestjs/nest) is a framework for building efficient, scalable Node.js server-side applications. With [nestjsx/crud](https://github.com/nestjsx/crud) we can add CRUD functions quickly and effortlessly on this framework.

## NestJS Rest Api

To start playing with NestJS you should have node (>= 10.13.0, except for v13) and [npm](https://nodejs.org) installed.

**Create Project Folder**

```bash
mkdir job-posting-app
cd job-posting-app
```

Setting up a new project is quite simple with the [Nest CLI](https://docs.nestjs.com/cli/overview). With npm installed, you can create a new Nest project with the following commands in your OS terminal:

```bash
npm i -g @nestjs/cli
nest new api
```

[TypeORM](https://github.com/typeorm/typeorm) is definitely the most mature ORM available in the node.js world. Since it's written in TypeScript, it works pretty well with the Nest framework. I chose mysql as database. TypeORM supports many databases (MySQL, MariaDB, Postgres etc.)

To start with this library we have to install all required dependencies:

```bash
npm install --save @nestjs/typeorm @nestjs/config typeorm mysql2
```

- Create an [.env.example](https://github.com/refinedev/refine-hackathon/tree/main/job-posting-app/blob/master/api/.env.example) file. Here we will save the database information.
- Create and configured a [docker-compose](https://github.com/refinedev/refine-hackathon/tree/main/job-posting-app/blob/master/api/docker-compose.yml) file for MySQL.
- Create a [ormconfig.ts](https://github.com/refinedev/refine-hackathon/tree/main/job-posting-app/blob/master/api/ormconfig.ts) file for migrations.
- Add the following scripts to the `package.json` file for migrations.

```bash
"typeorm": "ts-node -r tsconfig-paths/register ./node_modules/typeorm/cli.js",
"db:migration:generate": "npm run typeorm -- migration:generate",
"db:migration:run": "npm run typeorm -- migration:run",
"db:migration:revert": "npm run typeorm -- migration:revert",
"db:refresh": "npm run typeorm schema:drop && npm run db:migration:run"
```


<img src="https://refine.ams3.cdn.digitaloceanspaces.com/blog/2021-10-4-admin-panel-with-nestjs/refine_job.png" alt="refine_job" />
<br />
