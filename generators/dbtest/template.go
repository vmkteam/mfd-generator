package dbtest

const baseFileTemplate = `//nolint:all
package {{.Package}}

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"

	"{{.DBPackage}}"

	"github.com/go-pg/pg{{.GoPGVer}}"
)

var (
	errNotFound = errors.New("not found")
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

func setup() (*pg.DB, error) {
	u := env("DB_CONN", "postgresql://localhost:5432/{{.ProjectName}}?sslmode=disable")
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
`

const funcFileTemplate = `
package {{.Package}}

import (
	"fmt"
	"testing"
	{{- if .HasImports}}{{- range .Imports}}
	"{{.}}"
	{{- end }}
	{{- end }}

	"{{.DBPackage}}"

	"github.com/go-pg/pg{{.GoPGVer}}/orm"
	"github.com/brianvoe/gofakeit/v7"
)

`

const opFuncTypeTemplate = `type {{.Name}}OpFunc func(t *testing.T, dbo orm.DB, in *db.{{.Name}}) Cleaner
`

const funcTemplate = `func {{.Name}}(t *testing.T, dbo orm.DB, in *db.{{.Name}}, ops ...{{.Name}}OpFunc) (*db.{{.Name}}, Cleaner) {
	repo := db.New{{.Namespace}}Repo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &db.{{.Name}}{}
	}

	{{if .HasPKs}}
	// Check if PKs are provided
	if {{ range $i, $e := .PKs}}
    {{- if gt $i 0 }} && {{ end -}} in.{{$e.Field}} != {{$e.Zero}}
	{{- end}} {
		// Fetch the entity by PK
		{{.VarName}}, err := repo.{{.Name}}ByID(t.Context(){{range .PKs}}, in.{{.Field}}{{end}}, repo.Full{{$.Name}}())
		if err != nil {
			t.Fatal(err)
		}

		// We must find the entity by PK
		if {{.VarName}} == nil {
			t.Fatal(fmt.Errorf("fetch the main entity {{.Name}} by
			{{- range $i, $e := .PKs}} {{.Field}}=%v
			{{- if gt $i 0 }}, {{ end -}} 
			{{- end}}, err=%w"{{- range .PKs}}, in.{{.Field}}{{- end}}, errNotFound))
		}

		// Return if found without real cleanup
		return {{.VarName}}, emptyClean
	}
	{{- end}}

	for _, op := range ops {
		if cl := op(t, dbo, in); cl != nil {
			cleaners = append(cleaners, cl)
		}
	}

	// Create the main entity
	{{.VarName}}, err := repo.Add{{.Name}}(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}

	return {{.VarName}}, func() {
		{{- if .HasPKs}}
		if _, err := dbo.ModelContext(t.Context(), &db.{{.Name}}{ 
		{{- range $i, $e := .PKs}}
		{{- if gt $i 0 }}, {{ end -}}
		{{.Field}}: {{$.VarName}}.{{.Field}}{{end}} }).WherePK().Delete(); err != nil {
			t.Fatal(err)
		}
		{{- end}}

		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

`

const funcOpWithRelTemplate = `{{- if .HasRelations }}
func With{{.Name}}Relations(t *testing.T, dbo orm.DB, in *db.{{.Name}}) Cleaner {
	var cleaners []Cleaner

	// Prepare main relations
	{{- range .InitRels }}{{.}}{{ end }}

	{{- if .NeedPreparingDependedRelsFromRoot }}
	// Prepare nested relations which have the same relations
	{{- range .InitDependedRelsFromRoot }}{{.}}{{ end }}

	// Inject relation IDs into relations which have the same relations
	{{- range .PreparingDependedRelsFromRoot }}
	{{.}}
	{{- end}}
	{{- end}}

	// Check embedded entities by FK
	{{- $entity := . }}
	{{- range .Relations }}

	// {{.Name}}. Check if all FKs are provided.
	{{- $relation := .}}
	{{- range $entity.FillingPKs }}
	{{.}}
	{{- end }}
	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		rel, relatedCleaner := {{.Type}}(t, dbo, in.{{$relation.Name}}
		{{- if .Entity.HasRelations }}, With{{.Type}}Relations {{ end -}}
		, {{- if .Entity.NeedFakeFilling }} WithFake{{.Type}}{{ end -}}) 
		{{- range .Entity.FillingCreatedOrFoundRels }}
		{{.}}
		{{- end }}
		{{- if .Entity.NeedPreparingFillingSameAsRootRels }}
		// Fill the same relations as in {{$relation.Name}}
		{{- range .Entity.PreparingFillingSameAsRootRels }}
		{{.}}
		{{- end }}
		{{- end }}
		
		cleaners = append(cleaners, relatedCleaner)
	}
	{{end}}

	return func() {
		// Clean up related entities from the last to the first
		for i := len(cleaners) - 1; i >= 0; i-- {
			cleaners[i]()
		}
	}
}

{{- end}}`

const funcOpWithFakeTemplate = `{{- if .NeedFakeFilling }}
func WithFake{{.Name}}(t *testing.T, dbo orm.DB, in *db.{{.Name}}) Cleaner {
	{{- range .FakeFilling }}{{.}}{{ end }}
	
	return emptyClean
}

{{- end}}`
