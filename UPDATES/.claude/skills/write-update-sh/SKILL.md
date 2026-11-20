---
name: write-update-sh
description: Rules for writing or changing UPDATES/<version>/UPDATE.sh migrations that edit config files on existing OpenPanel servers (docker-compose.yml, .env, ini/conf files in the template and in every /home/<user>). Use whenever an UPDATE.sh adds or changes services, env vars, mounts or other lines in files that admins and users may have edited.
---

# Writing UPDATE.sh migrations

UPDATE.sh runs once on every existing server during the update. The files it
edits were written by older installers, edited by admins, or re-dumped by a
yaml library, so **never assume the file looks like the template in
openpanel-configuration**. Look at what the file actually is and add to it in
the same style.

## Indentation and format: detect, never hardcode

- Do **not** match services with a fixed indent like `grep "^  php-fpm-8.5:"`.
  Some accounts have compose files re-dumped with 4-space indentation (and
  sorted keys, `command:` first, no blank lines), and a fixed 2-space pattern
  silently skips them. It happened in 2.0.28: `testinguser` was skipped
  on two servers.
- Decide whether a service exists, and read its image or other fields, by
  **parsing the yaml** (`python3` + `yaml.safe_load`, available on every
  server), not with grep:
  ```bash
  # image of service $2 in compose file $1, "none" when there's no such service
  svc_image() {
      python3 -c 'import sys, yaml; s = (yaml.safe_load(open(sys.argv[1])) or {}).get("services") or {}; print((s[sys.argv[2]] or {}).get("image", "") if sys.argv[2] in s else "none")' "$1" "$2" 2>/dev/null || echo "none"
  }
  ```
- Edit the file **as text** (don't re-dump the yaml, that rewrites the whole
  file, drops comments and changes formatting). When inserting:
  - find the indent the file uses (indent of the first key under `services:`)
    and build regexes and new lines with that indent;
  - a block ends at the next non-empty, non-comment line whose indent is
    `<=` the service indent;
  - prefer **cloning an existing block from the same file** (e.g. copy
    `php-fpm-8.5` as `php-fpm-8.6` and substitute the version) over pasting a
    block from the template. The clone keeps that file's indentation, quoting,
    key order and setup (e.g. shinsenter vs openpanel images, which need
    different env, mounts and `command`).
- After substitutions, check that nothing unexpected is left (e.g. no `8.5`
  left in the cloned 8.6 block) and the image is one we know. Otherwise skip
  that file with a message: it's a custom setup.
- `.env`: add new keys after the related existing ones and copy their values
  where that makes sense (e.g. `PHP_FPM_8_6_*` from the user's `PHP_FPM_8_5_*`),
  match the file's quoting.

## Safety

- **Idempotent**: running the script twice must change nothing the second
  time. Check that the thing is missing before adding it (for every file:
  compose service, env keys, ini file, list entry).
- Only touch what the change is for. Respect the admin's choices: change a
  value only when it still has the old factory value (e.g.
  `DEFAULT_PHP_VERSION` only when it is exactly `8.5`), never touch users' own
  settings unless asked.
- Validate before and after: skip a file that already fails
  `podman-compose config` / yaml parse, back it up (`cp -p f f.bak-<version>`),
  and restore it if the check fails after the change. Remove the backup on
  success.
- Write in place with `cat tmp > file` (not `mv`), to keep owner,
  permissions and bind mounts. New files in a user's home get
  `chown --reference=<dir>`.
- The panel memoizes data in redis (`openpanel_cache_<key>`, see
  `cache.Memoize` calls in openpanel/ and `opencli/lib/redis.sh`). When the
  update changes something the panel caches (e.g. the account's PHP versions
  come from compose services, cached for an hour as
  `openpanel_cache_fetch_php_versions:<user>`), drop **only those exact keys**
  for the accounts that changed:
  `podman exec openpanel_redis redis-cli DEL "<key>"`. Never FLUSHALL or
  delete `openpanel_cache_*`, the same redis holds sessions and rate limits.
- Don't restart or start containers unless the change needs it.
- Print one line per change or skip, so the update log says what happened.

## Testing before release

1. Test locally against copies in the scratchpad: the current template, an
   old template from git history (e.g. the shinsenter one, `git show
   fdbb728^:docker/compose/1.0/docker-compose.yml` in openpanel-configuration),
   the same file re-dumped by `yaml.dump` with 2 and 4 space indent, a file
   without the service, and a file with a custom image. Run twice.
2. Then on a test server: back up the touched files to `/root/*.tgz` first,
   run only the new section fetched from GitHub raw (`?$(date +%s)` to dodge
   the cache), run it twice, check the result with the yaml parser, and start
   the new container for one account if it adds a service. Ask the user for
   server access each time and never write addresses or passwords into
   this repo.
