package vt

const modelDefaultTemplate = `//nolint:dupl
package {{.Package}}

import ({{if .HasImports}}{{range .Imports}}
    "{{.}}"{{end}}
{{end}}
	"{{.ModelPackage}}"
)
{{range $model := .Entities}}
type {{.Name}} struct { {{range .ModelColumns}}
	{{.Name}} {{.GoType}} {{.Tag}} {{.Comment}}{{end}}{{if .HasModelRelations}}
	{{range .ModelRelations}}
	{{.Name}} *{{.Type}}{{if ne .Type "Status"}}Summary{{end}} {{.Tag}}{{end}}{{end}}
}

func ({{.ShortVarName}} *{{.Name}}) ToDB() *db.{{.Name}} {
	if {{.ShortVarName}} == nil {
		return nil
	}{{range .ModelColumns}}{{if ne .ToDBName ""}}
		{{.ToDBFunc}}
	{{end}}{{end}}

	{{.VarName}} := &db.{{.Name}}{ {{range .ModelColumns}}{{if not .IsParams}}
		{{.Name}}: {{if ne .ToDBName ""}}{{.ToDBName}},{{else}}{{$model.ShortVarName}}.{{.FieldName}},{{end}}{{end}}{{end}}
	}
	{{range .ModelColumns}}{{if .IsParams}}{{if .NilCheck}}
	if {{$model.ShortVarName}}.{{.FieldName}} != nil {
		{{$model.VarName}}.{{.Name}} = {{$model.ShortVarName}}.{{.FieldName}}.ToDB()
	}{{else}}
	if {{$model.VarName}}{{.Name}} := {{$model.ShortVarName}}.{{.FieldName}}.ToDB(); {{$model.VarName}}{{.Name}} != nil {
		{{$model.VarName}}.{{.Name}} = *{{$model.VarName}}{{.Name}}
	}
	{{end}}{{end}}{{end}}

	return {{.VarName}}
}

type {{.Name}}Search struct {
	{{range .SearchColumns}}
	{{.Name}} {{.GoType}} {{.Tag}}{{end}}
}

func ({{.ShortVarName}}s *{{.Name}}Search) ToDB() *db.{{.Name}}Search {
	if {{.ShortVarName}}s == nil {
		return nil
	}{{range .SearchColumns}}{{if ne .ToDBName ""}}
		{{.ToDBFunc}}
	{{end}}{{end}}

	return &db.{{.Name}}Search{ {{range .SearchColumns}}
		{{.FieldName}}: {{if ne .ToDBName ""}}{{.ToDBName}},{{else}}{{$model.ShortVarName}}s.{{.Name}},{{end}}{{end}}
	}
}

type {{.Name}}Summary struct { {{range .SummaryColumns}}{{if ne .Name "StatusID"}}
	{{.Name}} {{.GoType}} {{.Tag}} {{.Comment}}{{end}}{{end}}{{if .HasSummaryRelations}}
	{{range .SummaryRelations}}
	{{.Name}} *{{.Type}}{{if ne .Name "Status"}}Summary{{end}} {{.Tag}}{{end}}{{end}}
}{{if .HasParams}}{{range .Params}}

type {{.Name}} struct {
}

func ({{.ShortVarName}} *{{.Name}}) ToDB() *db.{{.FieldName}} {
	return &db.{{.FieldName}}{}
}
{{end}}{{end}}
{{end}}`

const converterDefaultTemplate = `package {{.Package}}

import (
	"{{.ModelPackage}}"
)
{{range $model := .Entities}}
func New{{.Name}}(in *db.{{.Name}}) *{{.Name}} {
	if in == nil {
		return nil
	}{{range .ModelColumns}}{{if ne .FromDBName ""}}
		{{.FromDBFunc}}
	{{end}}{{end}}

	{{.VarName}} := &{{.Name}}{ {{range .ModelColumns}}{{if .IsParams}}{{if .NilCheck}}
		{{.Name}}: New{{.ParamsName}}(in.{{.Name}}),{{end}}{{else}}
		{{.Name}}: {{if ne .FromDBName ""}}{{.FromDBName}},{{else}}in.{{.Name}},{{end}}{{end}}{{end}}{{if .HasModelRelations}}
		{{range .ModelRelations}}
		{{.Name}}:   New{{.Type}}{{if ne .Name "Status"}}Summary{{end}}(in.{{.Name}}{{if eq .Name "Status"}}ID{{end}}),{{end}}{{end}}
	}
	{{range .ModelColumns}}{{if .IsParams}}{{if not .NilCheck}}
	if {{$model.VarName}}{{.Name}} := New{{.ParamsName}}(&in.{{.Name}}); {{$model.VarName}}{{.Name}} != nil {
		{{$model.VarName}}.{{.Name}} = *{{$model.VarName}}{{.Name}}
	}{{end}}{{end}}{{end}}

	return {{.VarName}}
}

func New{{.Name}}Summary(in *db.{{.Name}}) *{{.Name}}Summary {
	if in == nil {
		return nil
	}{{range .SummaryColumns}}{{if ne .FromDBName ""}}
		{{.FromDBFunc}}
	{{end}}{{end}}

	return &{{.Name}}Summary{ {{range .SummaryColumns}}{{if ne .Name "StatusID"}}{{if .IsParams}}
		{{.Name}}: New{{.ParamsName}}(in.{{.Name}}),{{else}}
		{{.Name}}: {{if ne .FromDBName ""}}{{.FromDBName}},{{else}}in.{{.Name}},{{end}}{{end}}{{end}}{{end}}{{if .HasSummaryRelations}}
		{{range .SummaryRelations}}
		{{.Name}}:   New{{.Type}}{{if ne .Name "Status"}}Summary{{end}}(in.{{.Name}}{{if eq .Name "Status"}}ID{{end}}),{{end}}{{end}}
	}
}{{if .HasParams}}{{range .Params}}

func New{{.Name}}(in *db.{{.Name}}) *{{.Name}} {
	return &{{.Name}}{
	}
}
{{end}}{{end}}
{{end}}`

const serviceDefaultTemplate = `package {{.Package}}

import (
	"context"{{if .HasImports}}{{range .Imports}}
    "{{.}}"{{end}}
{{end}}

	"{{.EmbedLogPackage}}"
	"{{.ModelPackage}}"

	"github.com/vmkteam/zenrpc/v2"
)

{{- range $model := .Entities }}
type {{.Name}}Service struct {
	zenrpc.Service
	embedlog.Logger
	{{$.VarName}}Repo db.{{$.Name}}Repo
    {{- if .HasRelations }}
    {{- range .UniqueRelations }}
    {{- if ne $.VarName .NameSpace }}
    {{.NameSpace}}Repo db.{{.NameSpace | title}}Repo
    {{- end }}
    {{- end}}
    {{- end}}
}

func New{{.Name}}Service(dbo db.DB, logger embedlog.Logger) *{{.Name}}Service {
	return &{{.Name}}Service{
		Logger:   logger,
		{{$.VarName}}Repo: db.New{{$.Name}}Repo(dbo),
        {{- if .HasRelations }}
        {{- range .UniqueRelations }}
        {{- if ne $.VarName .NameSpace }}
        {{.NameSpace}}Repo: db.New{{.NameSpace | title}}Repo(dbo),
        {{- end }}
        {{- end}}
        {{- end}}
	}
}

func (s {{.Name}}Service) dbSort(ops *ViewOps) db.OpFunc {
	v := s.{{$.VarName}}Repo.Default{{.Name}}Sort()
	if ops == nil {
		return v
	}{{if .HasSortColumns}}

	switch ops.SortColumn {
	case {{range $i, $e := .SortColumns}}{{if $i}}, {{end}}db.Columns.{{$model.Name}}.{{.}}{{end}}:
		v = db.WithSort(db.NewSortField(ops.SortColumn, ops.SortDesc))
	}
{{end}}
	return v
}

// Count returns count {{.NamePlural}} according to conditions in search params.
//
//zenrpc:search {{.Name}}Search
//zenrpc:return int
//zenrpc:500 Internal Error
func (s {{.Name}}Service) Count(ctx context.Context, search *{{.Name}}Search) (int, error) {
	count, err := s.{{$.VarName}}Repo.Count{{.NamePlural}}(ctx, search.ToDB())
	if err != nil {
		return 0, InternalError(err)
	}
	return count, nil
}

// Get returns а list of {{.NamePlural}} according to conditions in search params.
//
//zenrpc:search {{.Name}}Search
//zenrpc:viewOps ViewOps
//zenrpc:return []{{.Name}}Summary
//zenrpc:500 Internal Error
func (s {{.Name}}Service) Get(ctx context.Context, search *{{.Name}}Search, viewOps *ViewOps) ([]{{.Name}}Summary, error) {
	list, err := s.{{$.VarName}}Repo.{{.NamePlural}}ByFilters(ctx, search.ToDB(), viewOps.Pager(), s.dbSort(viewOps), s.{{$.VarName}}Repo.Full{{.Name}}())
	if err != nil {
		return nil, InternalError(err)
	}
	{{.VarNamePlural}} := make([]{{.Name}}Summary, 0, len(list))
	for i := 0; i {{$.Raw "<"}} len(list); i++ {
		if {{.VarName}} := New{{.Name}}Summary(&list[i]); {{.VarName}} != nil {
			{{.VarNamePlural}} = append({{.VarNamePlural}}, *{{.VarName}})
		}
	}
	return {{.VarNamePlural}}, nil
}

// GetByID returns a {{.Name}} by its ID.{{range .PKs}}
//
//zenrpc:{{.Arg}} {{.Type}}{{end}}
//zenrpc:return {{.Name}}
//zenrpc:500 Internal Error
//zenrpc:404 Not Found
func (s {{.Name}}Service) GetByID(ctx context.Context{{range .PKs}}, {{.Arg}} {{.Type}}{{end}}) (*{{.Name}}, error) {
	db, err := s.byID(ctx{{range .PKs}}, {{.Arg}}{{end}})
	if err != nil {
		return nil, err
	}
	return New{{.Name}}(db), nil
}

func (s {{.Name}}Service) byID(ctx context.Context{{range .PKs}}, {{.Arg}} {{.Type}}{{end}}) (*db.{{.Name}}, error) {
	db, err := s.{{$.VarName}}Repo.{{.Name}}ByID(ctx{{range .PKs}}, {{.Arg}}{{end}}, s.{{$.VarName}}Repo.Full{{.Name}}())
	if err != nil {
		return nil, InternalError(err)
	} else if db == nil {
		return nil, ErrNotFound
	}
	return db, nil
}{{if not .ReadOnly}}

// Add adds a {{.Name}} from the query.
//
//zenrpc:{{.VarName}} {{.Name}}
//zenrpc:return {{.Name}}
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
func (s {{.Name}}Service) Add(ctx context.Context, {{.VarName}} {{.Name}}) (*{{.Name}}, error) {
	if ve := s.isValid(ctx, {{.VarName}}, false); ve.HasErrors() {
		return nil, ve.Error()
	}

	db, err := s.{{$.VarName}}Repo.Add{{.Name}}(ctx, {{.VarName}}.ToDB())
	if err != nil {
		return nil, InternalError(err)
	}
	return New{{.Name}}(db), nil
}

// Update updates the {{.Name}} data identified by id from the query.
//
//zenrpc:{{.VarNamePlural}} {{.Name}}
//zenrpc:return {{.Name}}
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s {{.Name}}Service) Update(ctx context.Context, {{.VarName}} {{.Name}}) (bool, error) {
	if _, err := s.byID(ctx{{range .PKs}}, {{$model.VarName}}.{{.Field}}{{end}}); err != nil {
		return false, err
	}

	if ve := s.isValid(ctx, {{.VarName}}, true); ve.HasErrors() {
		return false, ve.Error()
	}

	ok, err := s.{{$.VarName}}Repo.Update{{.Name}}(ctx, {{.VarName}}.ToDB())
	if err != nil {
		return false, InternalError(err)
	}
	return ok, nil
}

// Delete deletes the {{.Name}} by its ID.{{range .PKs}}
//
//zenrpc:{{.Arg}} {{.Type}}{{end}}
//zenrpc:return isDeleted
//zenrpc:500 Internal Error
//zenrpc:400 Validation Error
//zenrpc:404 Not Found
func (s {{.Name}}Service) Delete(ctx context.Context{{range .PKs}}, {{.Arg}} {{.Type}}{{end}}) (bool, error) {
	if _, err := s.byID(ctx{{range .PKs}}, {{.Arg}}{{end}}); err != nil {
		return false, err
	}

	ok, err := s.{{$.VarName}}Repo.Delete{{.Name}}(ctx{{range .PKs}}, {{.Arg}}{{end}})
	if err != nil {
		return false, InternalError(err)
	}
	return ok, err
}

// Validate verifies that {{.Name}} data is valid.
//
//zenrpc:{{.VarName}} {{.Name}}
//zenrpc:return []FieldError
//zenrpc:500 Internal Error
func (s {{.Name}}Service) Validate(ctx context.Context, {{.VarName}} {{.Name}}) ([]FieldError, error) {
	isUpdate := {{range $i, $e := .PKs}}{{if $i}} && {{end}} {{$model.VarName}}.{{.Field}} != {{.Zero}} {{end}}
	if isUpdate {
		_, err := s.byID(ctx{{range .PKs}}, {{$model.VarName}}.{{.Field}}{{end}})
		if err != nil {
			return nil, err
		}
	}

	ve := s.isValid(ctx, {{.VarName}}, isUpdate)
	if ve.HasInternalError() {
		return nil, ve.Error()
	}

	return ve.Fields(), nil
}

func (s {{.Name}}Service) isValid(ctx context.Context, {{.VarName}} {{.Name}}, isUpdate bool) Validator {
	var v Validator

	if v.CheckBasic(ctx, {{.VarName}}); v.HasInternalError() {
		return v
	}

	{{if .HasAlias}}
	//check alias unique
	search := &db.{{.Name}}Search{ 
		{{.AliasArg}}: &{{$model.VarName}}.{{.AliasField}},{{range .PKSearches}}
		{{.Arg}}: &{{$model.VarName}}.{{.Field}},{{end}}
	}
	item, err := s.{{$.VarName}}Repo.One{{.Name}}(ctx, search)
	if err != nil {
		v.SetInternalError(err)
	} else if item != nil {
		v.Append("alias", FieldErrorUnique)
	}
	{{end}}

{{if .HasRelations}}
	// check fks{{range .Relations}}{{if .IsArray}}
		if len({{$model.VarName}}.{{.Name}}) != 0 {
		items, err := s.{{.NameSpace}}Repo.{{.PluralFK}}ByFilters(ctx, &db.{{.FK}}Search{IDs:{{$model.VarName}}.{{.Name}}},db.PagerNoLimit)
		if err != nil {
			v.SetInternalError(err)
		} else if len(items) != len({{$model.VarName}}.{{.Name}}) {
			v.Append("{{.JSONName}}", FieldErrorIncorrect)
		}
	}{{else}}
	if {{$model.VarName}}.{{.Name}} != {{if .Nullable}}nil{{else}}0{{end}} {
		item, err := s.{{.NameSpace}}Repo.{{.FK}}ByID(ctx, {{if .Nullable}}*{{end}}{{$model.VarName}}.{{.Name}})
		if err != nil {
			v.SetInternalError(err)
		} else if item == nil {
			v.Append("{{.JSONName}}", FieldErrorIncorrect)
		}
	}
	{{end}}{{end}}{{end}}
	//custom validation starts here
	return v
}

{{end}}{{end}}`

const serverDefaultTemplate = `
	Put this into your server code:

	const (
		NSAuth = "auth"
		NSUser = "user"

		{{range .Entities}}
		NS{{.Name}} = "{{.VarName}}"{{end}}
	)
	
	// services
	rpc.RegisterAll(map[string]zenrpc.Invoker{
		NSAuth: NewAuthService(dbo, logger),
		NSUser: NewUserService(dbo, logger),

		{{range .Entities}}
		NS{{.Name}}: New{{.Name}}Service(dbo, logger),{{end}}
	})
`
