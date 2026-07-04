#!/bin/bash
# ======================================================================
# Shared Redis cache-invalidation helpers, sourced by scripts that mutate
# data the OpenPanel UI (app.py / modules/*) has memoized in Redis.
#
# NEVER FLUSHALL or wildcard-DEL 'openpanel_cache_*' from a script. That
# nukes every memoized function for every user in one shot, and FLUSHALL
# additionally wipes login sessions and rate-limiter state that live in
# the same Redis instance. Always target the specific memver key(s) for
# the function(s) whose underlying data actually changed.
#
# Flask-Caching's @cache.memoize keeps one version key per function (not
# per set of arguments - so it can't be invalidated per-user/per-id),
# named:
#   openpanel_cache_<python.module.path>.<function_name>_memver
# Deleting that key invalidates every cached result of that function on
# its next read; it does not touch any other function's cache.
#
# This file only defines functions - it has no top-level logic/exit, so
# it is safe to `source` from any script.
# ======================================================================

REDIS_CONTAINER="openpanel_redis"

# Runs a redis-cli command inside the redis container.
redis_cli() {
    docker --context=default exec "$REDIS_CONTAINER" redis-cli "$@"
}

# Deletes one or more exact keys. No-op if called with no args.
redis_drop_key() {
    [ "$#" -eq 0 ] && return 0
    redis_cli DEL "$@" >/dev/null 2>&1
}

# Invalidates one or more @cache.memoize'd functions by dropping their
# "_memver" key, e.g.:
#   redis_drop_memver "app.get_user_details_with_plan" "modules.json.helpers.query_plan_details_by_id"
redis_drop_memver() {
    [ "$#" -eq 0 ] && return 0

    local func keys=()
    for func in "$@"; do
        keys+=( "openpanel_cache_${func}_memver" )
    done
    redis_drop_key "${keys[@]}"
}

# Terminates all active sessions for a given user id (forces re-login),
# used when a user's password/IP/account is changed or removed.
redis_drop_user_sessions() {
    local user_id="$1"
    [ -z "$user_id" ] && return 0

    local session_keys
    session_keys=$(redis_cli --scan --pattern "session:${user_id}:*")
    [ -z "$session_keys" ] && return 0

    local key
    while IFS= read -r key; do
        [ -n "$key" ] && redis_cli UNLINK "$key" >/dev/null 2>&1
    done <<< "$session_keys"
}
