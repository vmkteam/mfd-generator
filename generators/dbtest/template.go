package dbtest

const baseFileTemplate = `//nolint:all
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

func Setup(t *testing.T) ({{.DBPackageAlias}}.DB, embedlog.Logger) {
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
	return {{.DBPackageAlias}}.New(conn), logger
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
`

const funcFileTemplate = `
//nolint:dupl,funlen
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

const opFuncTypeTemplate = `type {{.Name}}OpFunc func(t *testing.T, dbo orm.DB, in *{{.DBPackageAlias}}.{{.Name}}) Cleaner
`

const funcTemplate = `func {{.Name}}(t *testing.T, dbo orm.DB, in *{{.DBPackageAlias}}.{{.Name}}, ops ...{{.Name}}OpFunc) (*{{.DBPackageAlias}}.{{.Name}}, Cleaner) {
	repo := {{.DBPackageAlias}}.New{{.Namespace}}Repo(dbo)
	var cleaners []Cleaner

	// Fill the incoming entity
	if in == nil {
		in = &{{.DBPackageAlias}}.{{.Name}}{}
	}

	{{if .HasPKs}}
	// Check if PKs are provided
	{{- range $i, $e := .PKs}}
    {{- if $e.IsCustom }}
    var def{{$e.Field}} {{$e.Type}}
    {{- end}}
    {{- end}}
    if {{ range $i, $e := .PKs}}
    {{- if and (gt $i 0) (ne $e.Type "bool") }} && {{ end -}} {{- if $e.IsCustom }}in.{{$e.Field}} != def{{$e.Field}}{{else if eq $e.Type "bool" }}{{else if eq $e.Type "time.Time" }}!in.{{$e.Field}}.IsZero(){{else}}in.{{$e.Field}} != {{$e.Zero}}{{- end}} 
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
		t.Logf("the entity {{.Name}} is not found by provided PKs:
		{{- range $i, $e := .PKs}}{{- if gt $i 0 }}, {{ end -}}{{.Field}}=%v{{- end}}. Trying to create one"
		{{- range .PKs}}, in.{{.Field}}{{- end}})
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
		if _, err := dbo.ModelContext(t.Context(), &{{.DBPackageAlias}}.{{.Name}}{ 
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
func With{{.Name}}Relations(t *testing.T, dbo orm.DB, in *{{.DBPackageAlias}}.{{.Name}}) Cleaner {
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

	{{- $entity := .}}
	{{- range .Relations }}
	{{- $relation := .}}
	// Fetch the relation. It creates if the FKs are provided it fetch from DB by PKs. Else it creates new one.
	{
		{{- if $relation.IsArray}}
		for i := range in.{{$relation.Name}} {
			{{- $pk := index $relation.Entity.PKs 0 }}
			_, relatedCleaner := {{.Type}}(t, dbo, &{{$entity.DBPackageAlias}}.{{.Type}}{ {{ $pk.Field }}: in.{{$relation.Name}}[i] }
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
func WithFake{{.Name}}(t *testing.T, dbo orm.DB, in *{{.DBPackageAlias}}.{{.Name}}) Cleaner {
	{{- range .FakeFilling }}{{.}}{{ end }}

	return emptyClean
}

{{- end}}`
