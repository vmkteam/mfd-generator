//nolint:all
package test

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata/actual/db"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type Cleaner func()

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

// Setup logger
type testDBLogQuery struct{}

func (d testDBLogQuery) BeforeQuery(ctx context.Context, _ *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (d testDBLogQuery) AfterQuery(_ context.Context, q *pg.QueryEvent) error {
	fq, err := q.FormattedQuery()
	if err != nil {
		return err
	}
	log.Println(string(fq))

	return nil
}

func Setup(t *testing.T) db.DB {
	// Connect to DB
	conn, err := setup()
	if err != nil {
		if t == nil {
			panic(err)
		}
		t.Fatal(err)
	}

	// Cleanup after testing
	if t != nil {
		t.Cleanup(func() {
			if err := conn.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}

	return db.New(conn)
}

func RefreshPK(t *testing.T, dbo orm.DB, tableName, columnName string) error {
	_, err := dbo.ExecContext(t.Context(), `
SELECT setval(
pg_get_serial_sequence('?2', ?0),
(SELECT MAX(?1) FROM ?2) + 1, false);
`, columnName, pg.Ident(columnName), pg.Ident(tableName))

	return err
}

func setup() (*pg.DB, error) {
	u := env("DB_CONN", "postgresql://localhost:5432/newsportal?sslmode=disable")
	cfg, err := pg.ParseURL(u)
	if err != nil {
		return nil, err
	}
	conn := pg.Connect(cfg)

	if r := env("DB_LOG_QUERY", "true"); r == "true" {
		conn.AddQueryHook(testDBLogQuery{})
	}

	return conn, nil
}

func env(v, def string) string {
	if r := os.Getenv(v); r != "" {
		return r
	}

	return def
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
