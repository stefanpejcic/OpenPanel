// Package mongomanager maintains a per-user *mongo.Client, connected over that account's own
// mongod unix socket - separate from the panel's own DB. Unlike Postgres/MySQL a Mongo client
// isn't bound to one database, so the cache key is just the userContext, and callers select a
// database per call via client.Database(name).
package mongomanager

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
)

var (
	clientsMu sync.Mutex
	clients   = map[string]*mongo.Client{}
)

// systemDatabases are never listed, created, or dropped through this package.
var systemDatabases = map[string]bool{"admin": true, "local": true, "config": true}

func IsSystemDatabase(name string) bool {
	return systemDatabases[name]
}

// creds reads MONGODB_ROOT_USER (default "admin") and MONGODB_ROOT_PASSWORD from /home/<context>/.env
func creds(userContext string) (user, password string, err error) {
	user = webserver.GetEnvFileValue(userContext, "MONGODB_ROOT_USER")
	if user == "" {
		user = "admin"
	}
	password = webserver.GetEnvFileValue(userContext, "MONGODB_ROOT_PASSWORD")
	if password == "" {
		return "", "", fmt.Errorf("MongoDB password not found in /home/%s/.env", userContext)
	}
	return user, password, nil
}

// openClient connects to the account's own mongod over its unix domain socket
func openClient(ctx context.Context, userContext string) (*mongo.Client, error) {
	user, password, err := creds(userContext)
	if err != nil {
		return nil, err
	}

	socketPath := "/home/" + userContext + "/sockets/mongodb/mongodb-27017.sock"
	if _, statErr := os.Stat(socketPath); statErr != nil {
		return nil, fmt.Errorf("MongoDB socket not found: %s", socketPath)
	}

	uri := fmt.Sprintf("mongodb://%s:%s@%s/?authSource=admin",
		url.QueryEscape(user), url.QueryEscape(password), url.QueryEscape(socketPath))

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if pingErr := client.Ping(connectCtx, nil); pingErr != nil {
		_ = client.Disconnect(context.Background())
		return nil, pingErr
	}
	return client, nil
}

func getClient(ctx context.Context, userContext string) (*mongo.Client, error) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	if client, ok := clients[userContext]; ok {
		return client, nil
	}

	client, err := openClient(ctx, userContext)
	if err != nil {
		return nil, err
	}
	clients[userContext] = client
	return client, nil
}

// InvalidateClient drops and disconnects the cached client for a user, if any
func InvalidateClient(userContext string) {
	clientsMu.Lock()
	client, ok := clients[userContext]
	if ok {
		delete(clients, userContext)
	}
	clientsMu.Unlock()

	if ok {
		_ = client.Disconnect(context.Background())
	}
}

// GetClient gets the cached client, pings it, and transparently reconnects once if that fails -
// covers a restarted mongod container without the caller needing to know
func GetClient(ctx context.Context, userContext string) (*mongo.Client, error) {
	client, err := getClient(ctx, userContext)
	if err == nil {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pingErr := client.Ping(pingCtx, nil)
		cancel()
		if pingErr == nil {
			return client, nil
		}
	}

	InvalidateClient(userContext)
	return getClient(ctx, userContext)
}

// DatabaseInfo is one entry of ListDatabases.
type DatabaseInfo struct {
	Name      string
	SizeBytes int64
}

// ListDatabases returns every non-system database, sorted by name.
func ListDatabases(ctx context.Context, userContext string) ([]DatabaseInfo, error) {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return nil, err
	}

	result, err := client.ListDatabases(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	dbs := make([]DatabaseInfo, 0, len(result.Databases))
	for _, d := range result.Databases {
		if IsSystemDatabase(d.Name) {
			continue
		}
		dbs = append(dbs, DatabaseInfo{Name: d.Name, SizeBytes: d.SizeOnDisk})
	}
	return dbs, nil
}

// CreateDatabase creates a database by inserting a placeholder collection - Mongo has no
// explicit CREATE DATABASE, a database only exists once it holds a collection.
func CreateDatabase(ctx context.Context, userContext, name string) error {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return err
	}
	return client.Database(name).CreateCollection(ctx, "_init")
}

// DropDatabase drops a database entirely.
func DropDatabase(ctx context.Context, userContext, name string) error {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return err
	}
	return client.Database(name).Drop(ctx)
}

// UserRole is one {role, db} pair granted to a Mongo user.
type UserRole struct {
	Role string `bson:"role"`
	DB   string `bson:"db"`
}

// UserInfo is one entry of ListUsers.
type UserInfo struct {
	Username string     `bson:"user"`
	Roles    []UserRole `bson:"roles"`
}

// ListUsers returns every user defined in the admin database.
func ListUsers(ctx context.Context, userContext string) ([]UserInfo, error) {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return nil, err
	}

	var result struct {
		Users []UserInfo `bson:"users"`
	}
	if runErr := client.Database("admin").RunCommand(ctx, bson.D{{Key: "usersInfo", Value: 1}}).Decode(&result); runErr != nil {
		return nil, runErr
	}
	return result.Users, nil
}

// CreateUser creates a Mongo user with no roles - access is granted afterward via GrantRole.
func CreateUser(ctx context.Context, userContext, username, password string) error {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return err
	}
	cmd := bson.D{
		{Key: "createUser", Value: username},
		{Key: "pwd", Value: password},
		{Key: "roles", Value: bson.A{}},
	}
	return client.Database("admin").RunCommand(ctx, cmd).Err()
}

// DropUser removes a Mongo user entirely.
func DropUser(ctx context.Context, userContext, username string) error {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return err
	}
	cmd := bson.D{{Key: "dropUser", Value: username}}
	return client.Database("admin").RunCommand(ctx, cmd).Err()
}

// GrantRole grants a user a role scoped to one database, e.g. GrantRole(ctx, uc, "bob", "readWrite", "mydb").
func GrantRole(ctx context.Context, userContext, username, role, database string) error {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return err
	}
	cmd := bson.D{
		{Key: "grantRolesToUser", Value: username},
		{Key: "roles", Value: bson.A{bson.D{{Key: "role", Value: role}, {Key: "db", Value: database}}}},
	}
	return client.Database("admin").RunCommand(ctx, cmd).Err()
}

// RevokeRole revokes a previously granted role from a user.
func RevokeRole(ctx context.Context, userContext, username, role, database string) error {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return err
	}
	cmd := bson.D{
		{Key: "revokeRolesFromUser", Value: username},
		{Key: "roles", Value: bson.A{bson.D{{Key: "role", Value: role}, {Key: "db", Value: database}}}},
	}
	return client.Database("admin").RunCommand(ctx, cmd).Err()
}

// ChangeUserPassword sets a new password for an existing Mongo user.
func ChangeUserPassword(ctx context.Context, userContext, username, newPassword string) error {
	client, err := GetClient(ctx, userContext)
	if err != nil {
		return err
	}
	cmd := bson.D{
		{Key: "updateUser", Value: username},
		{Key: "pwd", Value: newPassword},
	}
	return client.Database("admin").RunCommand(ctx, cmd).Err()
}

// Ping makes a single connection attempt - no retry/polling loop, that behavior belongs to the install wizard.
func Ping(ctx context.Context, userContext string) bool {
	_, err := GetClient(ctx, userContext)
	return err == nil
}
