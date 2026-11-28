package apidocs

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// typedGroups must have a body_schema on every write route, new routes in them fail this test until they get one
var typedGroups = map[string]bool{
	"MySQL": true, "PostgreSQL": true, "Domains": true, "DNS Zone Editor": true, "Cron Jobs": true,
	"Docker / Containers": true, "Git": true, "Node.js / Python Apps": true, "App Installer": true, "File Manager": true, "App Deploy": true, "WordPress": true,
}

type endpoint struct {
	Method      string          `json:"method"`
	Path        string          `json:"path"`
	Description string          `json:"description"`
	Body        json.RawMessage `json:"body"`
	Params      string          `json:"params"`
	BodySchema  json.RawMessage `json:"body_schema"`
	QuerySchema json.RawMessage `json:"query_schema"`
}

type group struct {
	Group     string     `json:"group"`
	Endpoints []endpoint `json:"endpoints"`
}

func loadGroups(t *testing.T) []group {
	t.Helper()
	var groups []group
	if err := json.Unmarshal(EndpointsJSON, &groups); err != nil {
		t.Fatalf("api_endpoints.json: %v", err)
	}
	return groups
}

var schemaKeys = map[string]bool{
	"type": true, "description": true, "enum": true, "default": true, "pattern": true, "format": true,
	"minLength": true, "maxLength": true, "minimum": true, "maximum": true, "minItems": true, "maxItems": true,
	"items": true, "properties": true, "required": true, "additionalProperties": true,
}

var jsonTypes = map[string]bool{"string": true, "integer": true, "number": true, "boolean": true, "array": true, "object": true, "null": true}

var contentTypes = map[string]bool{"application/x-www-form-urlencoded": true, "multipart/form-data": true}

func schemaTypes(s map[string]any) []string {
	switch v := s["type"].(type) {
	case string:
		return []string{v}
	case []any:
		var out []string
		for _, x := range v {
			if str, ok := x.(string); ok {
				out = append(out, str)
			}
		}
		return out
	}
	return nil
}

// checkProperty reports problems with one property schema, recursing into items and nested objects
func checkProperty(where string, s map[string]any) []string {
	var errs []string
	for k := range s {
		if !schemaKeys[k] {
			errs = append(errs, fmt.Sprintf("%s: unknown keyword %q", where, k))
		}
	}
	types := schemaTypes(s)
	if len(types) == 0 && s["enum"] == nil {
		errs = append(errs, where+": needs a type or an enum")
	}
	for _, ty := range types {
		if !jsonTypes[ty] {
			errs = append(errs, fmt.Sprintf("%s: unknown type %q", where, ty))
		}
	}
	if d, _ := s["description"].(string); strings.TrimSpace(d) == "" {
		errs = append(errs, where+": needs a description")
	}
	if items, ok := s["items"].(map[string]any); ok {
		errs = append(errs, checkItems(where+"[]", items)...)
	}
	if props, ok := s["properties"].(map[string]any); ok {
		for name, p := range props {
			ps, ok := p.(map[string]any)
			if !ok {
				errs = append(errs, where+"."+name+": not an object")
				continue
			}
			errs = append(errs, checkProperty(where+"."+name, ps)...)
		}
	}
	return errs
}

// array items don't need their own description
func checkItems(where string, s map[string]any) []string {
	if _, ok := s["description"]; !ok {
		s = copyWith(s, "description", "item")
	}
	return checkProperty(where, s)
}

func copyWith(m map[string]any, k string, v any) map[string]any {
	out := map[string]any{k: v}
	for kk, vv := range m {
		out[kk] = vv
	}
	return out
}

// checkObjectSchema validates a body_schema or query_schema and returns it parsed
func checkObjectSchema(where string, raw json.RawMessage) (map[string]any, []string) {
	var s map[string]any
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, []string{where + ": " + err.Error()}
	}
	var errs []string
	if s["type"] != "object" {
		errs = append(errs, where+": type must be object")
	}
	for k := range s {
		if k != "type" && k != "properties" && k != "required" && k != "x-content-type" && k != "additionalProperties" && k != "description" && k != "minProperties" && k != "anyOf" {
			errs = append(errs, fmt.Sprintf("%s: unknown top-level keyword %q", where, k))
		}
	}
	if ct, ok := s["x-content-type"].(string); ok && !contentTypes[ct] {
		errs = append(errs, fmt.Sprintf("%s: x-content-type %q", where, ct))
	}
	props, ok := s["properties"].(map[string]any)
	if !ok {
		return s, append(errs, where+": needs properties, {} when the route takes none")
	}
	for name, p := range props {
		ps, ok := p.(map[string]any)
		if !ok {
			errs = append(errs, where+"."+name+": not an object")
			continue
		}
		errs = append(errs, checkProperty(where+"."+name, ps)...)
	}
	// anyOf is only used for "at least one of these", so each branch is a required list of known properties
	if alts, ok := s["anyOf"].([]any); ok {
		for _, alt := range alts {
			m, _ := alt.(map[string]any)
			req, _ := m["required"].([]any)
			if len(m) != 1 || len(req) == 0 {
				errs = append(errs, where+": anyOf branches must only hold required")
			}
			for _, r := range req {
				if name, _ := r.(string); props[name] == nil {
					errs = append(errs, fmt.Sprintf("%s: anyOf requires %v, which isn't a property", where, r))
				}
			}
		}
	}
	if req, ok := s["required"]; ok {
		list, ok := req.([]any)
		if !ok {
			errs = append(errs, where+": required must be a list")
		}
		for _, r := range list {
			if name, _ := r.(string); props[name] == nil {
				errs = append(errs, fmt.Sprintf("%s: required %v isn't a property", where, r))
			}
		}
	}
	return s, errs
}

func valueMatches(v any, s map[string]any) bool {
	types := schemaTypes(s)
	if len(types) == 0 {
		return true
	}
	for _, ty := range types {
		switch val := v.(type) {
		case string:
			if ty == "string" {
				return true
			}
		case float64:
			if ty == "number" || (ty == "integer" && val == float64(int64(val))) {
				return true
			}
		case bool:
			if ty == "boolean" {
				return true
			}
		case []any:
			if ty == "array" {
				return true
			}
		case map[string]any:
			if ty == "object" {
				return true
			}
		case nil:
			if ty == "null" {
				return true
			}
		}
	}
	return false
}

// checkExample makes sure the documented example only uses known fields with the right types and has every required one
func checkExample(where string, schema map[string]any, example map[string]any, form bool) []string {
	var errs []string
	props, _ := schema["properties"].(map[string]any)
	for k, v := range example {
		p, ok := props[k].(map[string]any)
		if !ok {
			errs = append(errs, fmt.Sprintf("%s: example field %q isn't in the schema", where, k))
			continue
		}
		// form and query values are strings on the wire whatever the handler parses them into
		if _, isStr := v.(string); form && isStr {
			continue
		}
		if !valueMatches(v, p) {
			errs = append(errs, fmt.Sprintf("%s: example %s=%v doesn't match type %v", where, k, v, p["type"]))
		}
		if enum, ok := p["enum"].([]any); ok {
			found := false
			for _, e := range enum {
				if fmt.Sprint(e) == fmt.Sprint(v) {
					found = true
				}
			}
			if !found {
				errs = append(errs, fmt.Sprintf("%s: example %s=%v isn't one of %v", where, k, v, enum))
			}
		}
	}
	if req, ok := schema["required"].([]any); ok {
		for _, r := range req {
			if _, ok := example[r.(string)]; !ok {
				errs = append(errs, fmt.Sprintf("%s: example is missing required %q", where, r))
			}
		}
	}
	return errs
}

func TestSchemasAreValid(t *testing.T) {
	var errs []string
	typed := 0
	for _, g := range loadGroups(t) {
		for _, ep := range g.Endpoints {
			where := ep.Method + " " + ep.Path
			isWrite := ep.Method == "POST" || ep.Method == "PUT" || ep.Method == "PATCH"
			if len(ep.BodySchema) > 0 {
				typed++
				if ep.Method == "GET" || ep.Method == "HEAD" {
					errs = append(errs, where+": body_schema on a "+ep.Method)
				}
				s, e := checkObjectSchema(where+" body_schema", ep.BodySchema)
				errs = append(errs, e...)
				ct, _ := s["x-content-type"].(string)
				desc := strings.ToLower(ep.Description)
				if strings.Contains(desc, "form-encoded") && ct != "application/x-www-form-urlencoded" {
					errs = append(errs, where+": description says form-encoded but body_schema doesn't")
				}
				if strings.Contains(desc, "multipart/form-data") && ct != "multipart/form-data" {
					errs = append(errs, where+": description says multipart but body_schema doesn't")
				}
				if len(ep.Body) > 0 && s != nil {
					var example map[string]any
					if json.Unmarshal(ep.Body, &example) == nil {
						errs = append(errs, checkExample(where+" body", s, example, ct != "")...)
					}
				}
			} else if isWrite && typedGroups[g.Group] {
				errs = append(errs, where+": write route in a typed group without body_schema")
			}
			if len(ep.QuerySchema) > 0 {
				s, e := checkObjectSchema(where+" query_schema", ep.QuerySchema)
				errs = append(errs, e...)
				if ep.Params != "" && s != nil {
					// examples are written for people (?host=%), so split by hand instead of url-decoding
					example := map[string]any{}
					for _, pair := range strings.Split(strings.TrimPrefix(ep.Params, "?"), "&") {
						if k, v, _ := strings.Cut(pair, "="); k != "" {
							example[k] = v
						}
					}
					// a params example shows one way to call it, not every required key
					s2 := copyWith(s, "required", []any{})
					errs = append(errs, checkExample(where+" params", s2, example, true)...)
				}
			}
		}
	}
	sort.Strings(errs)
	for _, e := range errs {
		t.Error(e)
	}
	if typed == 0 {
		t.Error("no endpoint has a body_schema")
	}
}

func TestInstallExamplesUseStrongPasswords(t *testing.T) {
	// installers reject weak admin passwords at the default password_strength, so the examples have to pass it
	for _, g := range loadGroups(t) {
		for _, ep := range g.Endpoints {
			var body map[string]any
			if !strings.HasSuffix(ep.Path, "/install") || json.Unmarshal(ep.Body, &body) != nil {
				continue
			}
			for _, f := range []string{"admin_password", "owner_password"} {
				if pw, ok := body[f].(string); ok && !validators.IsPasswordStrongEnough(pw, 50) {
					t.Errorf("%s %s: example %s %q is too weak for the default strength", ep.Method, ep.Path, f, pw)
				}
			}
		}
	}
}
