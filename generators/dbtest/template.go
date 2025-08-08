package dbtest

const connTemplate = `//nolint:all
package {{.Package}}

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"

	"{{.DBPackage}}"

	"github.com/go-pg/pg{{.GoPGVer}}"
	"github.com/google/uuid"
	"github.com/vmkteam/embedlog"
)

var logger embedlog.Logger

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

type Cleaner func()

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

// For creating unique IDs.
var (
	existsIds  sync.Map
	emptyClean Cleaner = func() {}
)

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

func Ptr[T any](v T) *T {
	return &v
}
`
