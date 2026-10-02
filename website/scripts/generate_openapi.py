#!/usr/bin/env python3
"""Generates the OpenAPI spec from internal/core/apidocs/api_endpoints.json,
the same structured data source that drives the app's own MCP tool
registry. Writes two copies:
  - website/static/openpanel-openapi.yaml (public docs site download link,
    with an example.com placeholder server since it's downloaded standalone)
  - openpanel/static/openapi.yaml (same-origin copy the in-app Swagger UI
    at /account/api loads - a relative server URL so it resolves against
    whatever domain/port the panel is actually running on)

Usage: python3 website/scripts/generate_openapi.py
Requires: pip install pyyaml

The QUERY_PARAMS/FILE_RESPONSES/MULTIPART_SCHEMAS maps below encode detail
that api_endpoints.json doesn't capture (query strings, file-download
content types, multipart/form bodies) - cross-checked by hand against the
Go handlers and docs/panel/999_api.md. Update those maps, not the
generated YAML, when adding new endpoints of these kinds.
"""
import json
import os
import re
from collections import OrderedDict

import yaml

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
WEBSITE_DIR = os.path.dirname(SCRIPT_DIR)
REPO_ROOT = os.path.dirname(WEBSITE_DIR)
APIDOCS = os.path.join(REPO_ROOT, "openpanel", "internal", "core", "apidocs", "api_endpoints.json")
OUT_PATHS = [
    os.path.join(WEBSITE_DIR, "static", "openpanel-openapi.yaml"),
    os.path.join(REPO_ROOT, "openpanel", "static", "openapi.yaml"),
]

with open(APIDOCS) as f:
    groups = json.load(f)

PARAM_RE = re.compile(r'<([^<>]+)>')


def to_openapi_path(path):
    return PARAM_RE.sub(lambda m: '{' + m.group(1) + '}', path)


def path_params(path):
    return PARAM_RE.findall(path)


def json_schema_for_value(v):
    if isinstance(v, bool):
        return {"type": "boolean", "example": v}
    if isinstance(v, int):
        return {"type": "integer", "example": v}
    if isinstance(v, float):
        return {"type": "number", "example": v}
    if isinstance(v, list):
        item_schema = json_schema_for_value(v[0]) if v else {"type": "string"}
        return {"type": "array", "items": item_schema, "example": v}
    if isinstance(v, dict):
        return {"type": "object", "properties": {k: json_schema_for_value(val) for k, val in v.items()}, "example": v}
    return {"type": "string", "example": v}


def body_schema(body):
    props = {k: json_schema_for_value(v) for k, v in body.items()}
    return {"type": "object", "properties": props, "example": body}


# Query params documented via curl examples in docs/panel/999_api.md, keyed by (method, openapi_path)
QUERY_PARAMS = {
    ("GET", "/api/mysql/databases/{db_name}/export"): ["format"],
    ("GET", "/api/mysql/size"): ["unit", "show_all"],
    ("DELETE", "/api/mysql/users/{db_user}"): ["host"],
    ("GET", "/api/mysql/users/{db_user}/privileges/{db_name}"): ["host"],
    ("GET", "/api/domains/{domain}/logs"): ["page", "show_all"],
    ("GET", "/api/emails/configuration/{config_type}/{account}"): ["ssl"],
    ("GET", "/api/crons/log"): ["lines", "job"],
    ("GET", "/api/containers/{service}/logs"): ["lines"],
    ("GET", "/api/waf/log/{domain}"): ["page", "per_page"],
    ("GET", "/api/waf/stats/{domain}"): ["seconds"],
    ("GET", "/api/php/{version}/extensions/install/status"): ["install_id"],
    ("GET", "/api/usage/history"): ["page", "per_page"],
    ("GET", "/api/files/{path_param}"): ["hidden_files", "page"],
    ("GET", "/api/files"): ["hidden_files", "page"],
    ("DELETE", "/api/file-manager/delete"): ["filename", "path_param", "item_type", "mode"],
    ("POST", "/api/file-manager/copy"): ["item_name", "item_type", "path_param", "destination_path"],
    ("POST", "/api/file-manager/move"): ["item_name", "path_param", "destination_path"],
    ("POST", "/api/trash/restore"): ["filename"],
    ("DELETE", "/api/trash/delete"): ["filename"],
    ("GET", "/api/sites/{domain}/visitors"): ["seconds"],
    ("GET", "/api/pm2/{domain}/logs"): ["lines"],
    ("GET", "/api/file-manager/download-file/{filename}"): ["path_param"],
    ("GET", "/api/file-manager/view-file/{filename}"): ["path_param"],
    ("GET", "/api/account/activity"): ["page", "search", "show_all"],
    ("GET", "/api/search/{what}"): ["q", "folder", "ext"],
}

# File-download responses, keyed by (method, openapi_path) -> content type.
# Cross-checked against each handler's Content-Type header.
FILE_RESPONSES = {
    ("GET", "/api/emails/export"): "text/csv",
    ("GET", "/api/emails/configuration/{config_type}/{account}"): "application/octet-stream",
    ("GET", "/api/file-manager/download-file/{filename}"): "application/octet-stream",
    ("GET", "/api/backup-wizard/download/{filename}"): "application/gzip",
    ("GET", "/api/mysql/databases/{db_name}/export"): "application/octet-stream",
    ("GET", "/api/domains/{domain}/dns/export"): "text/plain",
    ("GET", "/api/postgresql/databases/{db_name}/export"): "application/sql",
    ("GET", "/api/ftp/configuration/{config_type}/{account}"): "application/xml",
    ("GET", "/api/file-manager/view-file/{filename}"): "text/plain",
    ("POST", "/api/backups/download"): "application/gzip",
}

# Form fields appinstall.HandleInstall reads, shared by the Node.js/Python/Java/Ruby installers
APP_INSTALL_FORM = {
    "type": "object", "properties": {
        "domain_id": {"type": "string"}, "service_name": {"type": "string"},
        "startup_file": {"type": "string"}, "cpu_limit": {"type": "string"},
        "mem_limit": {"type": "string"}, "pids_limit": {"type": "integer"}, "port": {"type": "integer"},
        "subdirectory": {"type": "string"}, "version": {"type": "string", "default": "latest"},
        "custom_cmd": {"type": "string"}, "requirements": {"type": "string"},
        "git_repo_url": {"type": "string", "format": "uri"},
    }, "required": ["domain_id", "service_name"]}

# n8n always listens on 5678 so port is ignored, and the owner account fields are required
N8N_INSTALL_FORM = {
    "type": "object", "properties": {
        **{k: v for k, v in APP_INSTALL_FORM["properties"].items() if k != "port"},
        "owner_email": {"type": "string", "format": "email"}, "owner_first_name": {"type": "string"},
        "owner_last_name": {"type": "string"}, "owner_password": {"type": "string", "minLength": 8, "maxLength": 64},
    }, "required": ["domain_id", "service_name", "owner_email", "owner_first_name", "owner_last_name", "owner_password"]}

# Multipart/form endpoints, keyed by openapi_path -> (content_type, schema)
MULTIPART_SCHEMAS = {
    "/api/mysql/databases/{db_name}/import": ("multipart/form-data", {
        "type": "object", "properties": {"file": {"type": "string", "format": "binary"}}, "required": ["file"]}),
    "/api/postgresql/databases/{db_name}/import": ("multipart/form-data", {
        "type": "object", "properties": {"file": {"type": "string", "format": "binary"}}, "required": ["file"]}),
    "/api/emails/import": ("multipart/form-data", {
        "type": "object", "properties": {"file": {"type": "string", "format": "binary"}}, "required": ["file"]}),
    "/api/file-manager/upload": ("multipart/form-data", {
        "type": "object", "properties": {
            "files": {"type": "array", "items": {"type": "string", "format": "binary"}},
            "path_param": {"type": "string"},
        }, "required": ["files"]}),
    "/api/file-manager/wget": ("multipart/form-data", {
        "type": "object", "properties": {
            "url": {"type": "string", "format": "uri"},
            "path_param": {"type": "string"},
        }, "required": ["url"]}),
    "/api/nodejs/install": ("application/x-www-form-urlencoded", APP_INSTALL_FORM),
    "/api/python/install": ("application/x-www-form-urlencoded", APP_INSTALL_FORM),
    "/api/java/install": ("application/x-www-form-urlencoded", APP_INSTALL_FORM),
    "/api/ruby/install": ("application/x-www-form-urlencoded", APP_INSTALL_FORM),
    "/api/n8n/install": ("application/x-www-form-urlencoded", N8N_INSTALL_FORM),
}


# Per-operation detail api_endpoints.json can't hold, keyed by (method, openapi_path), cross-checked against the Go handlers.
# Keys: description, operationId (pins a published id), param_types, body_enum, success (200 description), success_schema, errors.
OPERATION_EXTRAS = {
    ("POST", "/api/mysql/processlist/{id}/kill"): {
        "description": "Only entries whose command is Query can be killed. System threads are refused.",
        "operationId": "mysql_kill_query",
        "param_types": {"id": "integer"},
        "success": "Query killed.",
        "errors": {"400": "Invalid process ID.", "403": "System threads can not be killed.",
                   "404": "Process not found or no longer running a query."},
    },
    ("GET", "/api/mysql/configuration/recommendations"): {
        "description": "Checks the live server (memory limit, data per storage engine, status counters, connections "
                       "and error log) and returns suggested values for the configurable keys, with the reason for "
                       "each and the facts they are based on.",
        "operationId": "get_mysql_configuration_recommendations",
        "errors": {"503": "The database service could not be reached."},
    },
    ("POST", "/api/postgresql/processlist/{pid}/kill"): {
        "description": "Only active client backend queries can be cancelled. The connection stays open.",
        "operationId": "postgresql_kill_query",
        "param_types": {"pid": "integer"},
        "success": "Query cancelled.",
        "errors": {"400": "Invalid process ID.", "403": "Only user queries can be cancelled.",
                   "404": "Process not found or no longer running a query."},
    },
    ("GET", "/api/postgresql/configuration/recommendations"): {
        "description": "Checks the live server (memory and CPU limits, data size, statistics, connections and log) "
                       "and returns suggested values for the configurable keys, with the reason for each and the "
                       "facts they are based on.",
        "operationId": "get_postgresql_configuration_recommendations",
        "errors": {"503": "The database service could not be reached."},
    },
    ("GET", "/api/php/{version}/options/recommendations"): {
        "description": "Checks the PHP service (memory limit, workers, log), the domains and WordPress sites using "
                       "this version and the current options, and returns suggested values with the reason for each.",
        "operationId": "get_php_options_recommendations",
        "errors": {"400": "Invalid PHP version.", "404": "PHP version not installed."},
    },
    ("GET", "/api/malware-scanner/logs"): {
        "description": "Scans from the panel, the API and cron, newest first. The last 50 are kept.",
    },
    ("GET", "/api/crons/timezone"): {
        "description": "The time zone the cron service runs schedules in, from its TZ, the server zone when "
                       "/etc/localtime is mounted, or UTC.",
        "success_schema": {"type": "object", "properties": {"timezone": {"type": "string", "example": "Europe/Belgrade"}}},
    },
    ("PUT", "/api/crons/timezone"): {
        "description": "Sets TZ on the cron service in docker-compose.yml and recreates the service when a job is enabled.",
    },
    ("POST", "/api/waf/{domain}"): {
        "body_enum": {"status": ["On", "Off", "DetectionOnly"]},
    },
    ("POST", "/api/fix-permissions"): {
        "operationId": "post_fix_permissions",
    },
    ("POST", "/api/sites/bulk"): {
        "body_enum": {"action": ["detach", "update", "backup", "delete"]},
        "errors": {"400": "Invalid request body or no sites selected.",
                   "501": "Only the detach action is available via the API."},
    },
}


def slug_for_path(path):
    slug = PARAM_RE.sub(lambda m: m.group(1), path)
    if slug.startswith('/api/'):
        slug = slug[len('/api/'):]
    slug = slug.strip('/')
    slug = re.sub(r'[/\-.]', '_', slug)
    slug = re.sub(r'[^a-zA-Z0-9_]', '', slug)
    return slug or "root"


def build_spec(same_origin=False):
    paths = OrderedDict()
    tags = []
    seen_tags = set()

    slug_method_count = {}
    for g in groups:
        for e in g['endpoints']:
            oapi_path = to_openapi_path(e['path'])
            slug = slug_for_path(oapi_path)
            slug_method_count[slug] = slug_method_count.get(slug, 0) + 1

    operation_id_used = set()

    for g in groups:
        group_name = g['group']
        feature = g['feature']
        feature_str = feature if isinstance(feature, str) else " or ".join(feature)
        if group_name not in seen_tags:
            seen_tags.add(group_name)
            tags.append({
                "name": group_name,
                "description": f"Requires the `{feature_str}` feature to be enabled on the account's plan.",
            })

        for e in g['endpoints']:
            method = e['method'].lower()
            oapi_path = to_openapi_path(e['path'])
            description = e['description']

            if oapi_path not in paths:
                paths[oapi_path] = OrderedDict()

            extras = OPERATION_EXTRAS.get((e['method'], oapi_path), {})
            slug = slug_for_path(oapi_path)
            op_id = extras.get("operationId") or (f"{method}_{slug}" if slug_method_count[slug] > 1 else slug)
            base_op_id = op_id
            n = 2
            while op_id in operation_id_used:
                op_id = f"{base_op_id}_{n}"
                n += 1
            operation_id_used.add(op_id)

            operation = OrderedDict()
            operation["summary"] = description
            if "description" in extras:
                operation["description"] = extras["description"]
            operation["operationId"] = op_id
            operation["tags"] = [group_name]

            parameters = []
            for p in path_params(e['path']):
                parameters.append({
                    "name": p, "in": "path", "required": True,
                    "schema": {"type": extras.get("param_types", {}).get(p, "string")},
                })
            for qp in QUERY_PARAMS.get((e['method'], oapi_path), []):
                parameters.append({
                    "name": qp, "in": "query", "required": False,
                    "schema": {"type": "string"},
                })
            if parameters:
                operation["parameters"] = parameters

            if oapi_path in MULTIPART_SCHEMAS:
                content_type, schema = MULTIPART_SCHEMAS[oapi_path]
                operation["requestBody"] = {
                    "required": True,
                    "content": {content_type: {"schema": schema}},
                }
            elif "body" in e:
                schema = body_schema(e["body"])
                for field, values in extras.get("body_enum", {}).items():
                    schema["properties"][field]["enum"] = values
                operation["requestBody"] = {
                    "required": True,
                    "content": {"application/json": {"schema": schema}},
                }

            content_type = FILE_RESPONSES.get((e['method'], oapi_path))
            if content_type:
                success_response = {
                    "description": "File download.",
                    "content": {content_type: {"schema": {"type": "string", "format": "binary"}}},
                }
            else:
                success_response = {
                    "description": extras.get("success", "Success."),
                    "content": {"application/json": {"schema": extras.get(
                        "success_schema", {"type": "object", "additionalProperties": True})}},
                }

            responses = [("200", success_response)]
            responses += [(code, {"description": desc}) for code, desc in extras.get("errors", {}).items()]
            responses.append(("default", {"$ref": "#/components/responses/Error"}))
            operation["responses"] = OrderedDict(responses)

            operation["security"] = [{"bearerAuth": []}]

            paths[oapi_path][method] = operation

    spec = OrderedDict()
    spec["openapi"] = "3.0.3"
    spec["info"] = OrderedDict([
        ("title", "OpenPanel API"),
        ("version", "2.0.2"),
        ("description",
         "REST API for OpenPanel, a self-hosted web hosting control panel. "
         "Every feature available in the panel's web UI can also be accessed programmatically.\n\n"
         "All endpoints require a Bearer token obtained from `POST /api/login`. "
         "Beyond authentication, each endpoint also requires its documented feature to be "
         "enabled on the account's hosting plan.\n\n"
         "Success responses are JSON objects (`200`) unless noted otherwise (a handful of "
         "endpoints stream a file download instead). Error responses always include an "
         "`error` field and use standard HTTP status codes (400, 401, 403, 404, 409, 500, "
         "503, 504)."),
        ("contact", OrderedDict([("name", "OpenPanel"), ("url", "https://openpanel.com")])),
        ("license", OrderedDict([("name", "OpenPanel EULA"), ("url", "https://openpanel.com/LICENSE")])),
    ])
    if same_origin:
        # In-app copy is loaded by Swagger UI from the same host it runs
        # on, so a relative server URL resolves against whatever domain
        # and port the panel is actually reachable on instead of a
        # hardcoded placeholder.
        spec["servers"] = [
            {"url": "/", "description": "This OpenPanel installation"},
        ]
    else:
        spec["servers"] = [
            {"url": "https://panel.example.com", "description": "Your OpenPanel installation (default port 2083)"},
        ]
    spec["tags"] = tags
    spec["paths"] = paths
    spec["components"] = OrderedDict([
        ("securitySchemes", OrderedDict([
            ("bearerAuth", OrderedDict([
                ("type", "http"),
                ("scheme", "bearer"),
                ("bearerFormat", "JWT"),
                ("description", "Obtain a token via `POST /api/login`. Tokens expire after 24 hours."),
            ])),
        ])),
        ("responses", OrderedDict([
            ("Error", OrderedDict([
                ("description", "An error occurred. See the `error` field for details."),
                ("content", OrderedDict([
                    ("application/json", OrderedDict([
                        ("schema", OrderedDict([
                            ("type", "object"),
                            ("properties", OrderedDict([
                                ("error", {"type": "string", "example": "You do not own this domain"}),
                                ("hint", {"type": "string"}),
                            ])),
                            ("required", ["error"]),
                        ])),
                    ])),
                ])),
            ])),
        ])),
    ])
    spec["security"] = [{"bearerAuth": []}]
    return spec


class NoAliasDumper(yaml.SafeDumper):
    def ignore_aliases(self, data):
        return True


def represent_ordereddict(dumper, data):
    return dumper.represent_mapping('tag:yaml.org,2002:map', data.items())


def represent_str(dumper, data):
    if '\n' in data:
        return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='|')
    return dumper.represent_scalar('tag:yaml.org,2002:str', data)


NoAliasDumper.add_representer(OrderedDict, represent_ordereddict)
NoAliasDumper.add_representer(str, represent_str)


def main():
    website_out, app_out = OUT_PATHS
    specs = [
        (website_out, build_spec(same_origin=False)),
        (app_out, build_spec(same_origin=True)),
    ]
    for out, spec in specs:
        with open(out, 'w') as f:
            f.write("# Auto-generated from internal/core/apidocs/api_endpoints.json by generate_openapi.py.\n")
            f.write("# To update: regenerate rather than hand-editing this file directly.\n")
            yaml.dump(spec, f, Dumper=NoAliasDumper, sort_keys=False, allow_unicode=True, default_flow_style=False, width=100)
        print(f"Wrote {out}")

    total_ops = sum(len(v) for v in specs[0][1]["paths"].values())
    print(f"paths: {len(specs[0][1]['paths'])}, operations: {total_ops}, tags: {len(specs[0][1]['tags'])}")


if __name__ == "__main__":
    main()
