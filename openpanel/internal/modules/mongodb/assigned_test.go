package mongodb

import (
	"reflect"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
)

func TestAssignedDatabases(t *testing.T) {
	users := []mongomanager.UserInfo{
		{Username: "shopuser", Roles: []mongomanager.UserRole{{Role: "readWrite", DB: "shop"}, {Role: "dbAdmin", DB: "shop"}, {Role: "read", DB: "blog"}}},
		{Username: "app", Roles: []mongomanager.UserRole{{Role: "dbOwner", DB: "shop"}, {Role: "read", DB: "admin"}}},
		{Username: "lonely"},
	}
	want := []map[string]string{
		{"database": "blog", "users": "shopuser"},
		{"database": "shop", "users": "app,shopuser"},
	}
	if got := assignedDatabases(users); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if got := assignedDatabases(nil); got == nil || len(got) != 0 {
		t.Errorf("no users should give an empty list, got %#v", got)
	}
}
