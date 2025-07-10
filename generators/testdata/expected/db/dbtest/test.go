//nolint:all
package dbtest

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata/expected/db"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
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
	log.Println(q.FormattedQuery())
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

func setup() (*pg.DB, error) {
	url := "postgresql://localhost:5432/newsportal?sslmode=disable"
	if r := os.Getenv("DB_CONN"); r != "" {
		url = r
	}
	cfg, err := pg.ParseURL(url)
	if err != nil {
		return nil, err
	}
	conn := pg.Connect(cfg)

	if r := os.Getenv("DB_LOG_QUERY"); r == "true" {
		conn.AddQueryHook(testDBLogQuery{})
	}

	return conn, nil
}
