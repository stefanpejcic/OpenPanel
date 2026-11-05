package postgresmanager

import (
	"fmt"
	"testing"

	"github.com/lib/pq"
)

func TestDSNQuoteKeepsInjectedKeysInsideTheValue(t *testing.T) {
	db := "x host=/home/victim/sockets/postgres dbname=victimdb"
	pass := `p'w\d host=/evil`
	dsn := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable",
		dsnQuote("postgres"), dsnQuote(pass), dsnQuote("/home/alice/sockets/postgres"), dsnQuote(db))
	cfg, err := pq.NewConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "/home/alice/sockets/postgres" || cfg.Database != db || cfg.Password != pass {
		t.Fatalf("host=%q db=%q pass=%q", cfg.Host, cfg.Database, cfg.Password)
	}
}
