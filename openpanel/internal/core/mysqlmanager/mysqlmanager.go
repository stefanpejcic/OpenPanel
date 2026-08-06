// Package mysqlmanager maintains a per-user connection pool to that user's
// own MySQL/MariaDB instance, reached over
// a unix socket at /home/<context>/sockets/mysqld/mysqld.sock using
// credentials from that user's own /home/<context>/my.cnf - a completely
// separate database from the panel's own (internal/core/db), one pool per
// hosting account.
package mysqlmanager

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	socketWaitTimeout  = 15 * time.Second
	socketWaitInterval = 300 * time.Millisecond
)

var (
	poolsMu sync.Mutex
	pools   = map[string]*sql.DB{}
)

// mycnfCredentials reads the [client] section of /home/<context>/my.cnf.
// Same ini format as /etc/my.cnf (internal/core/db), but a distinct
// per-user file.
func mycnfCredentials(userContext string) (user, password string, err error) {
	path := fmt.Sprintf("/home/%s/my.cnf", userContext)
	f, err := os.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("MySQL config not readable: %s: %w", path, err)
	}
	defer f.Close()

	opts := map[string]string{}
	section := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		if section != "client" {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			opts[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", err
	}
	if _, ok := opts["user"]; !ok && section == "" {
		return "", "", fmt.Errorf("missing [client] section in %s", path)
	}
	return opts["user"], opts["password"], nil
}

// waitForSocket waits for a freshly-started mysqld to create its socket.
func waitForSocket(socketPath string) error {
	deadline := time.Now().Add(socketWaitTimeout)
	for {
		if _, err := os.Stat(socketPath); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("MySQL socket not found: %s", socketPath)
		}
		time.Sleep(socketWaitInterval)
	}
}

func openPool(userContext string) (*sql.DB, error) {
	user, password, err := mycnfCredentials(userContext)
	if err != nil {
		return nil, err
	}

	socketPath := fmt.Sprintf("/home/%s/sockets/mysqld/mysqld.sock", userContext)
	if err := waitForSocket(socketPath); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%s:%s@unix(%s)/?parseTime=true&timeout=10s",
		user, password, socketPath)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// One connection, reused: this hits a per-user mysqld meant for light
	// panel-side queries, not app traffic.
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func getPool(userContext string) (*sql.DB, error) {
	poolsMu.Lock()
	defer poolsMu.Unlock()

	if db, ok := pools[userContext]; ok {
		return db, nil
	}

	db, err := openPool(userContext)
	if err != nil {
		return nil, err
	}
	pools[userContext] = db
	return db, nil
}

// InvalidatePool drops and closes the cached pool for a user, e.g. after
// their mysqld container restarts.
func InvalidatePool(userContext string) {
	poolsMu.Lock()
	db, ok := pools[userContext]
	delete(pools, userContext)
	poolsMu.Unlock()

	if ok {
		_ = db.Close()
	}
}

// getConnection gets the pool, validates it with SELECT 1, and
// transparently reopens once if that fails (covers the
// mysqld-container-restarted case without the caller needing to know).
func getConnection(ctx context.Context, userContext string) (*sql.DB, error) {
	db, err := getPool(userContext)
	if err == nil {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pingErr := db.PingContext(pingCtx)
		cancel()
		if pingErr == nil {
			return db, nil
		}
	}

	InvalidatePool(userContext)
	return getPool(userContext)
}

// Exec runs one statement (optionally against a specific database first,
// via `USE `db“) and returns every row as a slice of columns - the same
// shape regardless of statement type, since callers range from SELECT
// COUNT(*) to SHOW GRANTS to CREATE DATABASE (which simply returns no
// rows).
func Exec(ctx context.Context, userContext, query string, database string) ([][]any, error) {
	db, err := getConnection(ctx, userContext)
	if err != nil {
		return nil, err
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if database != "" {
		safeDatabase := strings.ReplaceAll(database, "`", "``")
		if _, err := conn.ExecContext(ctx, "USE `"+safeDatabase+"`"); err != nil {
			return nil, err
		}
	}

	rows, err := conn.QueryContext(ctx, query)
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
// concrete types the driver may hand back for numeric aggregates ([]byte,
// int64, etc.) depending on the query - callers doing a COUNT(*)-style
// query just need `int(result[0][0])`.
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
