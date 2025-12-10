//nolint:all
package test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata/actual/db"

	"github.com/go-pg/pg/v10"
	"github.com/vmkteam/embedlog"
)

type Cleaner func()

// For creating unique IDs.
var (
	logger     embedlog.Logger
	existsIds  sync.Map
	emptyClean Cleaner = func() {}
)

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func Setup(t *testing.T) (db.DB, embedlog.Logger) {
	// Create db connection
	conn, err := setup()
	if err != nil {
		if t == nil {
			panic(err)
		}
		t.Fatal(err)
	}

	// Cleanup after tests.
	if t != nil {
		t.Cleanup(func() {
			if err := conn.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}

	logger = embedlog.NewLogger(true, true)
	return db.New(conn), logger
}

func setup() (*pg.DB, error) {
	var (
		pghost = getenv("PGHOST", "localhost")
		pgport = getenv("PGPORT", "5432")
		pgdb   = getenv("PGDATABASE", "test-apisrv")
		pguser = getenv("PGUSER", "postgres")
		pgpass = getenv("PGPASSWORD", "postgres")
	)

	url := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable", pguser, pgpass, net.JoinHostPort(pghost, pgport), pgdb)

	cfg, err := pg.ParseURL(url)
	if err != nil {
		return nil, err
	}
	conn := pg.Connect(cfg)

	if r := getenv("DB_LOG_QUERY", "false"); r == "true" {
		conn.AddQueryHook(testDBLogQuery{})
	}

	return conn, nil
}

type testDBLogQuery struct{}

func (d testDBLogQuery) BeforeQuery(ctx context.Context, _ *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (d testDBLogQuery) AfterQuery(ctx context.Context, q *pg.QueryEvent) error {
	if fm, err := q.FormattedQuery(); err == nil {
		logger.Print(ctx, string(fm))
	}
	return nil
}

func val[T any, P *T](p P) T {
	if p != nil {
		return *p
	}
	var def T
	return def
}

func cutS(str string, maxLen int) string {
	if maxLen == 0 {
		return str
	}
	return string([]rune(str)[:min(len(str), maxLen)])
}

func cutB(str string, maxLen int) []byte {
	if maxLen == 0 {
		return []byte(str)
	}
	return []byte(str)[:min(len(str), maxLen)]
}

// NextID Helps to generate unique IDs
func NextID() int {
	for {
		id := rand.Int31n(1<<30 - 1)
		if _, found := existsIds.LoadOrStore(id, struct{}{}); found {
			continue
		}
		return 1<<30 | int(id)
	}
}

// NextStringID The same as NextID, but converts the result to string
func NextStringID() string {
	return strconv.Itoa(NextID())
}
