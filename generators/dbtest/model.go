package dbtest

import (
	"cmp"
	"fmt"
	"html/template"
	"io"
	"slices"
	"strings"

	"github.com/vmkteam/mfd-generator/generators/model"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/util"
)

// FuncFileRenderData stores data for generating functions template
type FuncFileRenderData struct {
	Package   string
	DBPackage string

	ProjectName string
	GoPGVer     string
}

// PackFuncRenderData packs mfd namespace to template data
func PackFuncRenderData(options Options) FuncFileRenderData {
	var goPGVer string
	if options.GoPGVer != mfd.GoPG8 {
		goPGVer = fmt.Sprintf("/v%d", options.GoPGVer)
	}

	return FuncFileRenderData{
		GoPGVer:     goPGVer,
		Package:     options.Package,
		DBPackage:   options.DBPackage,
		ProjectName: options.ProjectName,
	}
}

// PKPair stores primary keys with type for template
type PKPair struct {
	Field string
	Arg   string
	Type  string
	Zero  template.HTML
}

func PackPKPair(column model.AttributeData) PKPair {
	arg := util.LowerFirst(column.Name)
	if column.Name == util.ID {
		arg = "id"
	}
	return PKPair{
		Field: column.Name,
		Arg:   arg,
		Type:  column.GoType,
		Zero:  template.HTML(mfd.MakeZeroValue(column.GoType)),
	}
}

// SortPair stores sort columns with direction for template
type SortPair struct {
	Field, Dir string
}

// NamespaceData stores namespace info for template
type NamespaceData struct {
	Package   string
	DBPackage string

	Name         string
	ShortVarName string

	HasImports bool
	Imports    []string

	GoPGVer string

	Entities []EntityData
}

// PackNamespace packs mfd namespace to template data
func PackNamespace(namespace *mfd.Namespace, options Options) NamespaceData {
	imports := mfd.NewSet()
	entities := make([]EntityData, len(namespace.Entities))
	name := util.CamelCased(util.Sanitize(namespace.Name))
	for i, entity := range namespace.Entities {
		packed := PackEntity(*entity, name, options)
		entities[i] = packed

		for _, imp := range packed.Imports {
			imports.Append(imp)
		}
	}

	goPGVer := ""
	if options.GoPGVer != mfd.GoPG8 {
		goPGVer = fmt.Sprintf("/v%d", options.GoPGVer)
	}

	return NamespaceData{
		Package:   options.Package,
		DBPackage: options.DBPackage,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		Name:         name,
		ShortVarName: mfd.ShortVarName(name),

		GoPGVer: goPGVer,

		Entities: entities,
	}
}

// EntityData stores entity info for template
type EntityData struct {
	Name       string
	NamePlural string

	Namespace string

	HasImports bool
	Imports    []string

	VarName       string
	VarNamePlural string

	HasStatus bool
	HasPKs    bool
	PKs       []PKPair

	SortField string
	SortDir   string

	HasRelations bool
	Relations    []RelationData

	// Helpers for filling NeedPreparingDependedRelsFromRoot and PreparingDependedRelsFromRoot
	relationByName            map[string]RelationData
	relationNamesHasRelations map[string]struct{}

	HasNestedSameRelations bool
	NestedSameRelations    []string

	NeedFakeFilling bool
	FakeFilling     []template.HTML

	NeedPreparingDependedRelsFromRoot bool
	PreparingDependedRelsFromRoot     []template.HTML
	InitDependedRelsFromRoot          []template.HTML

	NeedPreparingFillingSameAsRootRels bool
	PreparingFillingSameAsRootRels     []string

	Columns []AttributeData

	HasNotAddable bool
	NotAddable    []string

	HasNotUpdatable bool
	NotUpdatable    []string
}

// PackEntity packs mfd entity to template data
//
//nolint:funlen
func PackEntity(entity mfd.Entity, namespace string, options Options) EntityData {
	// base template entity - repo depends on int
	te := model.PackEntity(entity, model.Options{})

	var (
		notAddable      []string
		notUpdatable    []string
		pks             []PKPair
		hasStatus       bool
		hasNotAddable   bool
		hasNotUpdatable bool
	)

	imports := mfd.NewSet()
	columns := make([]AttributeData, 0, len(te.Columns))
	fakeFiller := NewFakeFiller()
	fakeFillingData := make([]template.HTML, 0, len(te.Columns))

	for _, column := range te.Columns {
		// if it has status - generate soft delete
		if mfd.IsStatus(column.DBName) {
			hasStatus = true
		}

		// if a key - generate arg(s) for GetByID function
		if column.PrimaryKey {
			if imp := mfd.Import(&column.Attribute, options.GoPGVer, options.CustomTypes); imp != "" {
				imports.Append(imp)
			}
			pks = append(pks, PackPKPair(column))
		}

		if !column.IsAddable() {
			hasNotAddable = true
			notAddable = append(notAddable, column.Name)
		}

		if !column.IsUpdatable() {
			hasNotUpdatable = true
			notUpdatable = append(notUpdatable, column.Name)
		}

		columns = append(columns, AttributeData{
			Name:          column.Name,
			ForeignKey:    column.ForeignKey,
			ForeignEntity: column.ForeignEntity,
			Addable:       column.IsAddable(),
			Updatable:     column.IsUpdatable(),
		})

		// Filling OpFunc which generates fake data
		if !column.Nullable() && !column.PrimaryKey && column.ForeignEntity == nil {
			// Check if the column has a known field name
			byFieldName, ok := fakeFiller.ByNameAndType(column.Name, column.GoType, column.Max)
			if ok {
				// If it is, it generates more for the field data
				condition := mustWrapFilling(column.Name, column.GoType, template.HTML(mfd.MakeZeroValue(column.GoType)), byFieldName)
				fakeFillingData = append(fakeFillingData, condition)
			} else {
				// By default, generates something depending on a field type
				condition := mustWrapFilling(column.Name, column.GoType, template.HTML(mfd.MakeZeroValue(column.GoType)), fakeFiller.ByType(column.Name, column.GoType, column.Max))
				fakeFillingData = append(fakeFillingData, condition)
			}
		}
	}

	imports.Append(fakeFiller.Imports()...)

	// store all relation names for join field
	relNames := make([]RelationData, len(te.Relations))
	relNamesMap := make(map[string]RelationData, len(te.Relations))
	relNamesWhichHasRels := make(map[string]struct{}, len(te.Relations))
	for i := range te.Relations {
		relNames[i] = PackRelationData(te.Relations[i], namespace, options)
		relNamesMap[relNames[i].Name] = relNames[i]
		if relNames[i].Entity.HasRelations {
			relNamesWhichHasRels[relNames[i].Name] = struct{}{}
		}
	}

	// Sort it from the most relation count to the least. Need to correctly order nested relations in rendering
	slices.SortFunc(relNames, func(a, b RelationData) int {
		return cmp.Compare(len(b.Entity.Relations), len(a.Entity.Relations))
	})

	// getting default sorts
	sortField, sortDir := sort(entity)

	// getting plural name for function name (eg CategoriesList)
	goNamePlural := mfd.MakePlural(te.Name)

	// getting var name for variable name (eg categories, newsList)
	varName := mfd.VarName(te.Name)
	varNamePlural := mfd.VarName(goNamePlural)
	if varName == varNamePlural {
		varNamePlural = fmt.Sprintf("%sList", varNamePlural)
	}

	res := EntityData{
		Name:       te.Name,
		NamePlural: goNamePlural,

		Namespace: namespace,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		VarName:       varName,
		VarNamePlural: varNamePlural,

		HasStatus: hasStatus,
		PKs:       pks,
		HasPKs:    len(pks) > 0,

		SortField: sortField,
		SortDir:   sortDir,

		Relations:                 relNames,
		HasRelations:              len(relNames) > 0,
		relationByName:            relNamesMap,
		relationNamesHasRelations: relNamesWhichHasRels,

		NeedFakeFilling: len(fakeFillingData) > 0,
		FakeFilling:     fakeFillingData,

		Columns:       columns,
		HasNotAddable: hasNotAddable,
		NotAddable:    notAddable,

		HasNotUpdatable: hasNotUpdatable,
		NotUpdatable:    notUpdatable,
	}

	curRel := res
	res.InitDependedRelsFromRoot, res.PreparingDependedRelsFromRoot = walkThroughDependedEntities(curRel.Relations, curRel, "in", "in")
	res.NeedPreparingDependedRelsFromRoot = len(res.PreparingDependedRelsFromRoot) > 0

	res.PreparingFillingSameAsRootRels = packPrepareSameAsRootRels(relNamesMap)
	res.NeedPreparingFillingSameAsRootRels = len(res.PreparingFillingSameAsRootRels) > 0

	return res
}

// walkThroughDependedEntities Walks through all relations of current relations recursively and finds same relations.
// Returns prepared strings to inject in a layout.
// The first value is inici
// The second one is filling root PKs to same nested relations.
func walkThroughDependedEntities(curRels []RelationData, parent EntityData, embeddedRels, root string) (initNestedRels, fillNestedRels []template.HTML) {
	var hasAlreadyPrepared bool
	for _, curEntity := range curRels {
		if _, ok := parent.relationNamesHasRelations[curEntity.Name]; ok {
			init, filling := walkThroughDependedEntities(curEntity.Entity.Relations, parent, embeddedRels+"."+curEntity.Name, root)
			initNestedRels = append(initNestedRels, init...) // Fill from the end to the start
			fillNestedRels = append(fillNestedRels, filling...)
		}

		if parentRel, ok := parent.relationByName[curEntity.Name]; ok && embeddedRels != root {
			for _, pk := range curEntity.Entity.PKs {
				needAmpersand := curEntity.NilCheck && !parentRel.NilCheck
				needVal := !curEntity.NilCheck && parentRel.NilCheck
				switch {
				case needAmpersand:
					fillNestedRels = append(fillNestedRels, template.HTML(fmt.Sprintf("%[1]s.%[2]s%[3]s = &%[4]s.%[2]s%[3]s", embeddedRels, curEntity.Name, pk.Field, root)))
				case needVal:
					fillNestedRels = append(fillNestedRels, template.HTML(fmt.Sprintf("%[1]s.%[2]s%[3]s = val(%[4]s.%[2]s%[3]s)", embeddedRels, curEntity.Name, pk.Field, root)))
				default:
					fillNestedRels = append(fillNestedRels, template.HTML(fmt.Sprintf("%[1]s.%[2]s%[3]s = %[4]s.%[2]s%[3]s", embeddedRels, curEntity.Name, pk.Field, root)))
				}
			}

			if !hasAlreadyPrepared {
				// Split the chain of relations by dots. We need to extract the last element
				relsChain := strings.Split(embeddedRels, ".")

				str := template.HTML(fmt.Sprintf(`
	if %[1]s == nil {
	%[1]s = &db.%[2]s{}
}`, embeddedRels, relsChain[len(relsChain)-1]))
				// Fill from the end to the start
				initNestedRels = append([]template.HTML{(str)}, initNestedRels...) // Fill from the end to the start
				hasAlreadyPrepared = true
			}
		}
	}

	return
}

func packPrepareSameAsRootRels(sameRelNames map[string]RelationData) []string {
	res := make([]string, 0, len(sameRelNames))
	for name := range sameRelNames {
		res = append(res, fmt.Sprintf("in.%[1]s = rel.%[1]s", name))
	}

	return res
}

// AttributeData stores attribute info for template
type AttributeData struct {
	Name          string
	ForeignKey    string
	ForeignEntity *mfd.Entity
	Addable       bool
	Updatable     bool
}

type RelationData struct {
	Name     string
	Type     string
	VarName  string
	NilCheck bool
	Entity   EntityData

	Tag     template.HTML
	Comment template.HTML
}

func PackRelationData(in model.RelationData, namespace string, options Options) RelationData {
	res := RelationData{
		Name:     in.Name,
		Type:     in.Type,
		VarName:  mfd.VarName(in.Name),
		Tag:      in.Tag,
		Comment:  in.Comment,
		NilCheck: in.Nullable,
	}

	if in.ForeignEntity != nil {
		res.Entity = PackEntity(*in.Entity, namespace, Options{GoPGVer: options.GoPGVer})
	}

	return res
}

func sort(entity mfd.Entity) (string, string) {
	presets := []SortPair{
		{"createdAt", "SortDesc"},
		{"title", "SortAsc"},
	}

	for _, preset := range presets {
		for _, attribute := range entity.Attributes {
			if attribute.DBName == preset.Field {
				return util.ColumnName(attribute.Name), preset.Dir
			}
		}
	}

	for _, attribute := range entity.Attributes {
		if attribute.PrimaryKey {
			return util.ColumnName(attribute.Name), "SortDesc"
		}
	}

	return "", ""
}

type OpFuncLayoutBuilder interface {
	Name(entity string) string
	Render(w io.Writer, data any) error
}

type OpFuncWithRelations struct{}

func (op OpFuncWithRelations) Name(entity string) string {
	return fmt.Sprintf("With%sRelations", entity)
}

func (op OpFuncWithRelations) Render(w io.Writer, data any) error {
	return loadAndParseTemplate(w, funcOpWithRelTemplate, data)
}

type OpFuncWithFake struct{}

func (op OpFuncWithFake) Name(entity string) string {
	return fmt.Sprintf("WithFake%s", entity)
}

func (op OpFuncWithFake) Render(w io.Writer, data any) error {
	return loadAndParseTemplate(w, funcOpWithFakeTemplate, data)
}

func loadAndParseTemplate(w io.Writer, tmpl string, data any) error {
	return Render(w, tmpl, data)
}
