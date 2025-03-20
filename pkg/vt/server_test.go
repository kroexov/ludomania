package vt

import (
	"context"
	"fmt"
	"os"
	"testing"

	"apisrv/pkg/db"

	"github.com/go-pg/pg/v10"
)

var (
	showSQL = false
)

var dbConn = env("DB_CONN", "postgresql://localhost:5432/apisrv?sslmode=disable")

func env(v, def string) string {
	if r := os.Getenv(v); r != "" {
		return r
	}

	return def
}

var testDb db.DB

type testdbLogger struct{}

func (d testdbLogger) BeforeQuery(ctx context.Context, _ *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (d testdbLogger) AfterQuery(_ context.Context, q *pg.QueryEvent) error {
	fmt.Println(q.FormattedQuery())
	return nil
}

func TestMain(m *testing.M) {
	testDb = NewTestDb()
	runTests := m.Run()
	os.Exit(runTests)
}

func NewTestDb() db.DB {
	cfg, err := pg.ParseURL(dbConn)
	if err != nil {
		panic(err)
	}
	dbc := pg.Connect(cfg)
	if showSQL {
		dbc.AddQueryHook(testdbLogger{})
	}
	return db.New(dbc)

}
