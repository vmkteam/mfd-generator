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
	"github.com/go-pg/pg{{.GoPGVer}}/orm"
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
	_, err := dbo.ExecContext(t.Context(),` + "`" + `
SELECT setval(
pg_get_serial_sequence('?2', ?0),
(SELECT MAX(?1) FROM ?2) + 1, false);
` + "`" + `, columnName, pg.Ident(columnName), pg.Ident(tableName))

	return err
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
//nolint:dupl
package {{.Package}}

import (
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
	{{- range $i, $e := .PKs}}
    {{- if $e.IsCustom }}
    var def{{$e.Field}} {{$e.Type}}
    {{- end}}
    {{- end}}
    if {{ range $i, $e := .PKs}}
    {{- if gt $i 0 }} && {{ end -}} {{- if $e.IsCustom }}in.{{$e.Field}} != def{{$e.Field}}{{else}}in.{{$e.Field}} != {{$e.Zero}}{{- end}} 
	{{- end}} {
		// Fetch the entity by PK
		{{.VarName}}, err := repo.{{.Name}}ByID(t.Context(){{range .PKs}}, in.{{.Field}}{{end}}, repo.Full{{$.Name}}())
		if err != nil {
			t.Fatal(err)
		}

		{{- if .AddIfNotFoundByPKFlow }}
		// Return if found without real cleanup
		if {{.VarName}} != nil {
			return {{.VarName}}, emptyClean
		}

		// If we're here, we don't find the entity by PKs. Just try to add the entity by provided PK
		t.Logf("the entity {{.Name}} is not found by provided PKs,
		{{- range $i, $e := .PKs}} {{.Field}}=%v
		{{- if gt $i 0 }}, {{ end -}} 
		{{- end}}. Trying to create one"{{- range .PKs}}, in.{{.Field}}{{- end}})
		{{- else }}

		// We must find the entity by PK
		if {{.VarName}} == nil {
			t.Fatalf("the entity {{.Name}} is not found by provided PKs
			{{- range $i, $e := .PKs}} {{.Field}}=%v
			{{- if gt $i 0 }}, {{ end -}} 
			{{- end}}"{{- range .PKs}}, in.{{.Field}}{{- end}})
		}

		// Return if found without real cleanup
		return {{.VarName}}, emptyClean
		{{- end }}
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

	{{- if .NeedInitDependedRelsFromRoot }}
	// Prepare nested relations which have the same relations
	{{- range .InitDependedRelsFromRoot }}{{.}}
	{{- end }}
	{{- end }}

	// Check if all FKs are provided. Fill them into the main struct rels
	{{- $entity := . }}{{- range $entity.FillingPKs }}
	{{.}}
	{{- end }}

	{{- if .NeedPreparingDependedRelsFromRoot }}
	// Inject relation IDs into relations which have the same relations
	{{- range .PreparingDependedRelsFromRoot }}
	{{.}}
	{{- end}}
	{{- end}}

	{{- range .Relations }}
	{{- $relation := .}}
	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		{{- if $relation.IsArray}}
		for i := range in.{{$relation.Name}} {
			{{- $pk := index $relation.Entity.PKs 0 }}
			_, relatedCleaner := {{.Type}}(t, dbo, &db.{{.Type}}{ {{ $pk.Field }}: in.{{$relation.Name}}[i] }
			{{- if .Entity.HasRelations }}, With{{.Type}}Relations {{ end }}, {{ if .Entity.NeedFakeFilling }} WithFake{{.Type}}{{ end -}})
			{{- if $entity.NeedPreparingFillingSameAsRootRels }}
			{{- range $relName, $vals := $entity.PreparingFillingSameAsRootRels }}
			{{- if eq $relName $relation.Name}}
			// Fill the same relations as in {{$relation.Name}}
			{{- range $vals }}
			{{.}}
			{{- end }}
			{{- end }}
			{{- end }}
			{{- end }}

			cleaners = append(cleaners, relatedCleaner)
		}
		{{- else}}
		rel, relatedCleaner := {{.Type}}(t, dbo, in.{{$relation.Name}}
		{{- if .Entity.HasRelations }}, With{{.Type}}Relations {{ end }}, {{ if .Entity.NeedFakeFilling }} WithFake{{.Type}}{{ end -}})
		{{- range .Entity.FillingCreatedOrFoundRels }}
		{{.}}
		{{- end }}
		{{- if $entity.NeedPreparingFillingSameAsRootRels }}
		{{- range $relName, $vals := $entity.PreparingFillingSameAsRootRels }}
		{{- if eq $relName $relation.Name}}
		// Fill the same relations as in {{$relation.Name}}
		{{- range $vals }}
		{{.}}
		{{- end }}
		{{- end }}
		{{- end }}
		{{- end }}

		cleaners = append(cleaners, relatedCleaner)
		{{- end}}
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
