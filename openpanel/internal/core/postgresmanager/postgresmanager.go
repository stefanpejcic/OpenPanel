// Package postgresmanager maintains a per-(user, database) connection pool
// to that user's own PostgreSQL
// instance, reached over a unix socket at /home/<context>/sockets/postgres
// using credentials from that user's own env file - a completely separate
// database from the panel's own (internal/core/db), one pool per
// (hosting account, database) pair, since unlike MySQL a Postgres
// connection is bound to a single database for its lifetime.
package postgresmanager

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
)

type poolKey struct {
	context  string
	database string
}

var (
	poolsMu sync.Mutex
	pools   = map[poolKey]*sql.DB{}
)

// creds reads POSTGRES_USER (default "postgres") and POSTGRES_PASSWORD
// from /home/<context>/.env, erroring if the password is missing (there is
// no sensible default for it).
func creds(userContext string) (user, password string, err error) {
	user = webserver.GetEnvFileValue(userContext, "POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}
	password = webserver.GetEnvFileValue(userContext, "POSTGRES_PASSWORD")
	if password == "" {
		return "", "", fmt.Errorf("PostgreSQL password not found in /home/%s/.env", userContext)
	}
	return user, password, nil
}

// openPool opens one *sql.DB per (context, database), connecting over the
// account's own postgres unix socket directory.
func openPool(userContext, database string) (*sql.DB, error) {
	user, password, err := creds(userContext)
	if err != nil {
		return nil, err
	}

	socketDir := "/home/" + userContext + "/sockets/postgres"
	if _, statErr := os.Stat(socketDir); statErr != nil {
		return nil, fmt.Errorf("PostgreSQL socket not found: %s", socketDir)
	}

	dsn := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable connect_timeout=10",
		user, password, socketDir, database)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// A handful of connections at most, reused across light panel-side
	// queries against this one account's own database.
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func getPool(userContext, database string) (*sql.DB, error) {
	poolsMu.Lock()
	defer poolsMu.Unlock()

	key := poolKey{userContext, database}
	if db, ok := pools[key]; ok {
		return db, nil
	}

	db, err := openPool(userContext, database)
	if err != nil {
		return nil, err
	}
	pools[key] = db
	return db, nil
}

// InvalidatePool drops and closes the cached pool(s) for a user - either
// just one database, or every database for that user when database is "".
func InvalidatePool(userContext, database string) {
	poolsMu.Lock()
	var toClose []*sql.DB
	if database != "" {
		key := poolKey{userContext, database}
		if db, ok := pools[key]; ok {
			toClose = append(toClose, db)
			delete(pools, key)
		}
	} else {
		for key, db := range pools {
			if key.context == userContext {
				toClose = append(toClose, db)
				delete(pools, key)
			}
		}
	}
	poolsMu.Unlock()

	for _, db := range toClose {
		_ = db.Close()
	}
}

// getConnection gets the pool, validates it with SELECT 1, and
// transparently reopens once if that fails (covers the
// postgres-container-restarted case without the caller needing to know).
func getConnection(ctx context.Context, userContext, database string) (*sql.DB, error) {
	db, err := getPool(userContext, database)
	if err == nil {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pingErr := db.PingContext(pingCtx)
		cancel()
		if pingErr == nil {
			return db, nil
		}
	}

	InvalidatePool(userContext, database)
	return getPool(userContext, database)
}

// Exec runs one statement against the given database (defaulting to
// "postgres" when unset) and returns every row as a slice of columns.
func Exec(ctx context.Context, userContext, query string, database string, args ...any) ([][]any, error) {
	if database == "" {
		database = "postgres"
	}
	db, err := getConnection(ctx, userContext, database)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result [][]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		result = append(result, vals)
	}
	return result, rows.Err()
}

// ToInt converts one Exec result cell to an int, tolerating the different
// concrete types the driver may hand back (int64, []byte, etc.).
func ToInt(v any) int {
	switch t := v.(type) {
	case int64:
		return int(t)
	case int:
		return t
	case []byte:
		n, _ := strconv.Atoi(string(t))
		return n
	case string:
		n, _ := strconv.Atoi(t)
		return n
	default:
		return 0
	}
}
