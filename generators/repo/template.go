package repo

const repoDefaultTemplate = `
package {{.Package}}

import (
	"context"
	"errors"{{if .HasImports}}{{range .Imports}}
	"{{.}}"{{end}}
	{{end}}

	"github.com/go-pg/pg{{.GoPGVer}}"
	"github.com/go-pg/pg{{.GoPGVer}}/orm"
)

type {{.Name}}Repo struct {
	db orm.DB
	filters map[string][]Filter
	sort    map[string][]SortField
	join    map[string][]string
}

// New{{.Name}}Repo returns new repository
func New{{.Name}}Repo(db orm.DB) {{.Name}}Repo {
	return {{.Name}}Repo{
		db:     db,
		filters: map[string][]Filter{
			{{- range .Entities}}{{if .HasStatus}}
			Tables.{{.Name}}.Name: {StatusFilter}, {{end}}{{end}} 
		},
		sort: map[string][]SortField{
			{{- range .Entities}}{{if ne .SortField ""}}
			Tables.{{.Name}}.Name: { {Column: Columns.{{.Name}}.{{.SortField}}, Direction: {{.SortDir}}} },{{end}}{{end}}
		},
		join: map[string][]string{
			{{- range $i, $e := .Entities}}
			Tables.{{$e.Name}}.Name: {TableColumns{{range .Relations}}, Columns.{{$e.Name}}.{{.Name}}{{end}} },{{end}} 
		},
	}
}

// WithTransaction is a function that wraps {{.Name}}Repo with pg.Tx transaction.
func ({{.ShortVarName}}r {{.Name}}Repo) WithTransaction(tx *pg.Tx) {{.Name}}Repo {
	{{.ShortVarName}}r.db = tx
	return {{.ShortVarName}}r
}

// WithEnabledOnly is a function that adds "statusId"=1 as base filter.
func ({{.ShortVarName}}r {{.Name}}Repo) WithEnabledOnly() {{.Name}}Repo {
	f := make(map[string][]Filter,len({{.ShortVarName}}r.filters))
	for i := range {{.ShortVarName}}r.filters {
    	f[i] = make([]Filter,len({{.ShortVarName}}r.filters[i]))
        copy(f[i], {{.ShortVarName}}r.filters[i])
        f[i] = append(f[i], StatusEnabledFilter)
	}
	{{.ShortVarName}}r.filters = f

	return {{.ShortVarName}}r
}

{{range $i, $e := .Entities}}/*** {{.Name}} ***/

// Full{{.Name}} returns full joins with all columns
func ({{$.ShortVarName}}r {{$.Name}}Repo) Full{{.Name}}() OpFunc {
	return WithColumns({{$.ShortVarName}}r.join[Tables.{{.Name}}.Name]...)
}

// Default{{.Name}}Sort returns default sort.
func ({{$.ShortVarName}}r {{$.Name}}Repo) Default{{.Name}}Sort() OpFunc {
	return WithSort({{$.ShortVarName}}r.sort[Tables.{{.Name}}.Name]...)
}
{{if .HasPKs}}
// {{.Name}}ByID is a function that returns {{.Name}} by ID(s) or nil.
func ({{$.ShortVarName}}r {{$.Name}}Repo) {{.Name}}ByID(ctx context.Context{{range .PKs}}, {{.Arg}} {{.Type}}{{end}}, ops ...OpFunc) (*{{.Name}}, error) {
	return {{$.ShortVarName}}r.One{{.Name}}(ctx, &{{.Name}}Search{ {{range $i, $e := .PKs}}{{if $i}}, {{end}}{{.Field}}: &{{.Arg}}{{end}} }, ops...)
}
{{end}}

// One{{.Name}} is a function that returns one {{.Name}} by filters. It could return pg.ErrMultiRows.
func ({{$.ShortVarName}}r {{$.Name}}Repo) One{{.Name}}(ctx context.Context, search *{{.Name}}Search, ops ...OpFunc) (*{{.Name}}, error) {
	obj := &{{.Name}}{}
	err := buildQuery(ctx, {{$.ShortVarName}}r.db, obj, search, {{$.ShortVarName}}r.filters[Tables.{{.Name}}.Name], PagerTwo, ops...).Select()

	if errors.Is(err, pg.ErrMultiRows) {
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}

	return obj, err
}

// {{.NamePlural}}ByFilters returns {{.Name}} list.
func ({{$.ShortVarName}}r {{$.Name}}Repo) {{.NamePlural}}ByFilters(ctx context.Context, search *{{.Name}}Search, pager Pager, ops ...OpFunc) ({{.VarNamePlural}} []{{.Name}}, err error) {
	err = buildQuery(ctx, {{$.ShortVarName}}r.db, &{{.VarNamePlural}}, search, {{$.ShortVarName}}r.filters[Tables.{{.Name}}.Name], pager, ops...).Select()
	return
}

// Count{{.NamePlural}} returns count
func ({{$.ShortVarName}}r {{$.Name}}Repo) Count{{.NamePlural}}(ctx context.Context, search *{{.Name}}Search, ops ...OpFunc) (int, error) {
	return buildQuery(ctx, {{$.ShortVarName}}r.db, &{{.Name}}{}, search, {{$.ShortVarName}}r.filters[Tables.{{.Name}}.Name], PagerOne, ops...).Count()
}

// Add{{.Name}} adds {{.Name}} to DB.
func ({{$.ShortVarName}}r {{$.Name}}Repo) Add{{.Name}}(ctx context.Context, {{.VarName}} *{{.Name}}, ops ...OpFunc) (*{{.Name}}, error) {
	q := {{$.ShortVarName}}r.db.ModelContext(ctx, {{.VarName}})
	{{- if .HasNotAddable }}
	if len(ops) == 0 {
		q = q.ExcludeColumn({{range .NotAddable}}Columns.{{$e.Name}}.{{.}},{{end}})
	}
	{{- end }}
	applyOps(q, ops...)
	_, err := q.Insert()

	return {{.VarName}}, err
}

// Update{{.Name}} updates {{.Name}} in DB.
func ({{$.ShortVarName}}r {{$.Name}}Repo) Update{{.Name}}(ctx context.Context, {{.VarName}} *{{.Name}}, ops ...OpFunc) (bool, error) {
	q := {{$.ShortVarName}}r.db.ModelContext(ctx, {{.VarName}}).WherePK()
	{{- if .HasNotUpdatable }}
	if len(ops) == 0 {
		q = q.ExcludeColumn({{range .NotUpdatable}}Columns.{{$e.Name}}.{{.}},{{end}})
    }
    {{- end }}
	applyOps(q, ops...)
	res, err := q.Update()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err
}
{{if .HasPKs}}
// Delete{{.Name}} {{if .HasStatus}}set statusId to deleted in DB{{else}}deletes {{.Name}} from DB{{end}}.
func ({{$.ShortVarName}}r {{$.Name}}Repo) Delete{{.Name}}(ctx context.Context{{range .PKs}}, {{.Arg}} {{.Type}}{{end}}) (deleted bool, err error) {
	{{.VarName}} := &{{.Name}}{ {{range $i, $e := .PKs}}{{if $i}}, {{end}}{{.Field}}: {{.Arg}}{{end}}{{if .HasStatus}}, StatusID: StatusDeleted,{{end}} }

{{if .HasStatus}}return {{$.ShortVarName}}r.Update{{.Name}}(ctx, {{.VarName}}, WithColumns(Columns.{{.Name}}.StatusID)){{else}}res, err := {{$.ShortVarName}}r.db.ModelContext(ctx, {{.VarName}}).WherePK().Delete()
	if err != nil {
		return false, err
	}

	return res.RowsAffected() > 0, err{{end}}
}{{end}}
{{end}}`

const repoRuntimeTemplate = `// Code generated by mfd-generator; DO NOT EDIT.

//nolint:all
package {{.Package}}

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

// Applier is the shared go-pg query transformation contract.
type Applier = func(query *orm.Query) (*orm.Query, error)

// RepoOption configures generated generic repositories.
type RepoOption func(*repoConfig)

type repoConfig struct {
	filters        []Applier
	insertDefaults []Applier
	updateDefaults []Applier
	deleteOptions  []Applier
	setDeleted     func(any)
}

// WithFilters adds filters applied to every read query.
func WithFilters(filters ...Applier) RepoOption {
	return func(config *repoConfig) {
		config.filters = append(config.filters, filters...)
	}
}

// WithInsertDefaults configures columns applied when no insert options are supplied.
func WithInsertDefaults(options ...Applier) RepoOption {
	return func(config *repoConfig) {
		config.insertDefaults = append(config.insertDefaults, options...)
	}
}

// WithUpdateDefaults configures columns applied when no update options are supplied.
func WithUpdateDefaults(options ...Applier) RepoOption {
	return func(config *repoConfig) {
		config.updateDefaults = append(config.updateDefaults, options...)
	}
}

// WithSoftDelete configures status-based deletion for a generated model.
func WithSoftDelete[T any](setDeleted func(*T), options ...Applier) RepoOption {
	return func(config *repoConfig) {
		config.setDeleted = func(model any) {
			setDeleted(model.(*T))
		}
		config.deleteOptions = append(config.deleteOptions, options...)
	}
}

// ApplyFilter adapts a generated Filter to a query applier.
func ApplyFilter(filter Filter) Applier {
	return func(query *orm.Query) (*orm.Query, error) {
		return filter.Apply(query), nil
	}
}

// ApplyPager adapts a generated Pager to a query applier.
func ApplyPager(pager Pager) Applier {
	return func(query *orm.Query) (*orm.Query, error) {
		return pager.Apply(query), nil
	}
}

// ApplyOp adapts a legacy query operation to a query applier.
func ApplyOp(operation OpFunc) Applier {
	return func(query *orm.Query) (*orm.Query, error) {
		operation(query)
		return query, nil
	}
}

// GenRepo provides generic repository operations for one generated model.
type GenRepo[T any] struct {
	db             orm.DB
	filters        []Applier
	insertDefaults []Applier
	updateDefaults []Applier
	deleteOptions  []Applier
	setDeleted     func(any)
}

// NewGenRepo creates a generic repository backed by go-pg/v10.
func NewGenRepo[T any](db orm.DB, options ...RepoOption) GenRepo[T] {
	config := repoConfig{}
	for _, option := range options {
		option(&config)
	}

	return GenRepo[T]{
		db:             db,
		filters:        cloneAppliers(config.filters),
		insertDefaults: cloneAppliers(config.insertDefaults),
		updateDefaults: cloneAppliers(config.updateDefaults),
		deleteOptions:  cloneAppliers(config.deleteOptions),
		setDeleted:     config.setDeleted,
	}
}

// WithTransaction returns a repository bound to tx.
func (repo GenRepo[T]) WithTransaction(tx *pg.Tx) GenRepo[T] {
	repo.db = tx
	return repo
}

// AppendFilter returns a repository with one additional base filter.
func (repo GenRepo[T]) AppendFilter(filter Applier) GenRepo[T] {
	repo.filters = append(cloneAppliers(repo.filters), filter)
	return repo
}

// One returns one model or nil when no row matches.
func (repo GenRepo[T]) One(ctx context.Context, search Applier, options ...Applier) (*T, error) {
	model := new(T)
	query, err := repo.readQuery(ctx, model, search, nil, options...)
	if err != nil {
		return nil, err
	}
	err = query.Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return model, nil
}

// List returns all models matching the query.
func (repo GenRepo[T]) List(ctx context.Context, search, pager Applier, options ...Applier) ([]T, error) {
	models := make([]T, 0)
	query, err := repo.readQuery(ctx, &models, search, pager, options...)
	if err != nil {
		return nil, err
	}
	if err := query.Select(); err != nil {
		return nil, err
	}
	return models, nil
}

// Count returns the number of matching models.
func (repo GenRepo[T]) Count(ctx context.Context, search Applier, options ...Applier) (int, error) {
	model := new(T)
	query, err := repo.readQuery(ctx, model, search, nil, options...)
	if err != nil {
		return 0, err
	}
	return query.Count()
}

// Add inserts model and returns the same pointer.
func (repo GenRepo[T]) Add(ctx context.Context, model *T, options ...Applier) (*T, error) {
	query := repo.db.ModelContext(ctx, model)
	if len(options) == 0 {
		options = repo.insertDefaults
	}
	query, err := applyAppliers(query, options...)
	if err != nil {
		return nil, err
	}
	_, err = query.Insert()
	if err != nil {
		return nil, err
	}
	return model, nil
}

// Update updates model by its primary key and reports whether a row changed.
func (repo GenRepo[T]) Update(ctx context.Context, model *T, options ...Applier) (bool, error) {
	query := repo.db.ModelContext(ctx, model).WherePK()
	if len(options) == 0 {
		options = repo.updateDefaults
	}
	query, err := applyAppliers(query, options...)
	if err != nil {
		return false, err
	}
	result, err := query.Update()
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

// Delete deletes model by its primary key and reports whether a row changed.
func (repo GenRepo[T]) Delete(ctx context.Context, model *T, options ...Applier) (bool, error) {
	if repo.setDeleted != nil {
		repo.setDeleted(model)
		options = append(cloneAppliers(options), repo.deleteOptions...)
		query, err := applyAppliers(repo.db.ModelContext(ctx, model).WherePK(), options...)
		if err != nil {
			return false, err
		}
		result, err := query.Update()
		if err != nil {
			return false, err
		}
		return result.RowsAffected() > 0, nil
	}

	query, err := applyAppliers(repo.db.ModelContext(ctx, model).WherePK(), options...)
	if err != nil {
		return false, err
	}
	result, err := query.Delete()
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

func (repo GenRepo[T]) readQuery(ctx context.Context, model any, search, pager Applier, options ...Applier) (*orm.Query, error) {
	query := repo.db.ModelContext(ctx, model)
	var err error
	if query, err = applyAppliers(query, repo.filters...); err != nil {
		return nil, err
	}
	if query, err = applyAppliers(query, search); err != nil {
		return nil, err
	}
	if query, err = applyAppliers(query, pager); err != nil {
		return nil, err
	}
	return applyAppliers(query, options...)
}

func applyAppliers(query *orm.Query, appliers ...Applier) (*orm.Query, error) {
	for _, applier := range appliers {
		if applier != nil {
			var err error
			query, err = applier(query)
			if err != nil {
				return query, err
			}
		}
	}
	return query, nil
}

func cloneAppliers(appliers []Applier) []Applier {
	return append([]Applier(nil), appliers...)
}
`

const repoGenericTemplate = `package {{.Package}}

import (
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type {{.Name}}Repo struct {
{{- range .Entities}}
	{{.Name}} GenRepo[{{.Name}}]
{{- end}}
}

// New{{.Name}}Repo returns a generic repository collection.
func New{{.Name}}Repo(db orm.DB) {{.Name}}Repo {
return {{.Name}}Repo{
{{- range .Entities}}
{{- $entity := .}}
		{{$entity.Name}}: NewGenRepo[{{$entity.Name}}](db{{if $entity.HasStatus}}, WithFilters(ApplyFilter(StatusFilter)), WithSoftDelete(func(model *{{$entity.Name}}) { model.StatusID = StatusDeleted }, ApplyOp(WithColumns(Columns.{{$entity.Name}}.StatusID)), ApplyOp(func(query *orm.Query) { query.Set("? = ?", pg.Ident(Columns.{{$entity.Name}}.StatusID), StatusDeleted) })){{end}}{{if $entity.HasNotAddable}}, WithInsertDefaults(ApplyOp(WithoutColumns({{range $entity.NotAddable}}Columns.{{$entity.Name}}.{{.}},{{end}}))){{end}}{{if $entity.HasNotUpdatable}}, WithUpdateDefaults(ApplyOp(WithoutColumns({{range $entity.NotUpdatable}}Columns.{{$entity.Name}}.{{.}},{{end}}))){{end}}),
{{- end}}
	}
}

// WithTransaction returns a repository collection bound to tx.
func (repo {{.Name}}Repo) WithTransaction(tx *pg.Tx) {{.Name}}Repo {
{{- range .Entities}}
	repo.{{.Name}} = repo.{{.Name}}.WithTransaction(tx)
{{- end}}
	return repo
}

// WithEnabledOnly returns a repository collection that selects enabled rows.
func (repo {{.Name}}Repo) WithEnabledOnly() {{.Name}}Repo {
{{- range .Entities}}{{if .HasStatus}}
	repo.{{.Name}} = repo.{{.Name}}.AppendFilter(ApplyFilter(StatusEnabledFilter))
{{- end}}{{- end}}
	return repo
}

{{range .Entities}}
{{- $entity := .}}
// Full{{.Name}} returns an operation that loads all generated relations.
func (repo {{$.Name}}Repo) Full{{.Name}}() Applier {
	return ApplyOp(WithColumns(TableColumns{{range $entity.Relations}}, Columns.{{$entity.Name}}.{{.Name}}{{end}}))
}

// Default{{.Name}}Sort returns an operation for the generated default sort.
func (repo {{$.Name}}Repo) Default{{.Name}}Sort() Applier {
{{- if ne .SortField ""}}
	return ApplyOp(WithSort(SortField{Column: Columns.{{.Name}}.{{.SortField}}, Direction: {{.SortDir}}}))
{{- else}}
	return func(query *orm.Query) (*orm.Query, error) { return query, nil }
{{- end}}
}

// {{.Name}}Query returns typed search and pager appliers.
func (repo {{$.Name}}Repo) {{.Name}}Query(search *{{.Name}}Search, pager Pager) (Applier, Applier) {
	var searchApply Applier
	if search != nil {
		searchApply = search.Q()
	}
	return searchApply, ApplyPager(pager)
}
{{end}}
`
