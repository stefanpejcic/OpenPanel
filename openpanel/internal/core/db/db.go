// Package db provides the panel's own MySQL connection pool, reading
// credentials from /etc/my.cnf (the standard MySQL option-file format).
package db

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const OptionFile = "/etc/my.cnf"

// clientOptions parses the [client] section of a MySQL option file
// (ini-style: "key = value" lines under a "[section]" header).
func clientOptions(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
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
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		opts[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return opts, scanner.Err()
}

// Open builds a *sql.DB pool from the [client] section of optionFile,
// matching connect_to_database()'s pool_size=10 / database=panel setup.
func Open(optionFile string) (*sql.DB, error) {
	opts, err := clientOptions(optionFile)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", optionFile, err)
	}

	database := opts["database"]
	if database == "" {
		database = "panel"
	}

	port := opts["port"]
	if port == "" {
		port = "3306"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		opts["user"], opts["password"], opts["host"], port, database)

	pool, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	pool.SetMaxOpenConns(10) // matches pool_size=10
	pool.SetMaxIdleConns(10)
	pool.SetConnMaxLifetime(time.Hour)

	return pool, nil
}
