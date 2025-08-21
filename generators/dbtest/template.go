package dbtest

const baseFileTemplate = `//nolint:all
package {{.Package}}

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"

	"{{.DBPackage}}"

	"github.com/go-pg/pg{{.GoPGVer}}"
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
	url := "postgresql://localhost:5432/{{.ProjectName}}?sslmode=disable"
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
`

const funcFileTemplate = `
package {{.Package}}

import (
	"errors"
	"fmt"
	"testing"

	"{{.DBPackage}}"
)

var (
	errNotFound = errors.New("not found")
)

`

const funcTemplate = `// {{.Entity.Name}} creates and returns a {{.Entity.Name}} entity with cleanup function
func {{.Entity.Name}}(t *testing.T, dbo db.DB, in *db.{{.Entity.Name}}) (*db.{{.Entity.Name}}, Cleaner) {
	repo := db.New{{.Namespace}}Repo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.{{.Entity.Name}}{}
	}

	{{if .Entity.HasPKs}}
	// Check if PKs are provided
	if {{ range $i, $e := .Entity.PKs}}
    {{- if gt $i 0 }} && {{ end -}} in.{{$e.Field}} != {{$e.Zero}}
	{{- end}} {
		// Fetch the entity by PK
		{{.Entity.VarName}}, err := repo.{{.Entity.Name}}ByID(t.Context(){{range .Entity.PKs}}, in.{{.Field}}{{end}}, repo.Full{{$.Entity.Name}}())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if {{.Entity.VarName}} == nil {
			t.Fatal(fmt.Errorf("fetch the main entity {{.Entity.Name}} by
			{{- range $i, $e := .Entity.PKs}} {{.Field}}=%v
			{{- if gt $i 0 }}, {{ end -}} 
			{{- end}}, err=%w"{{- range .Entity.PKs}}, in.{{.Field}}{{- end}}, errNotFound))
		}

		// Return if found without real cleanup
		return {{.Entity.VarName}}, emptyClean
	}
	{{- end}}

	{{if .Entity.HasRelations}}
	// Check embedded entities by PK
	{{- range .Entity.Relations}}
	if in.{{.Name}} == nil {
		rel, relatedCleaner := {{.Name}}(t, dbo, nil)
		in.{{.Name}} = rel
		cleaners = append(cleaners, relatedCleaner)
	}
	{{end}}{{end}}

	// Create the main entity
	{{.Entity.VarName}}, err := repo.Add{{.Entity.Name}}(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return {{.Entity.VarName}}, func() {
		if _, err := dbo.ModelContext(t.Context(), &db.{{.Entity.Name}}{ {{range .Entity.PKs}}{{.Field}}: {{$.Entity.VarName}}.{{.Field}}{{end}} }).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

`
