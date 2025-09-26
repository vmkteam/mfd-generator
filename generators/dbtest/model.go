package dbtest

import (
	"cmp"
	"fmt"
	"html/template"
	"io"
	"regexp"
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
	Field    string
	Arg      string
	Type     string
	FK       *mfd.Entity
	Nullable bool
	Zero     template.HTML
}

func PackPKPair(column model.AttributeData) PKPair {
	arg := util.LowerFirst(column.Name)
	if column.Name == util.ID {
		arg = "id"
	}

	return PKPair{
		Field:    column.Name,
		Arg:      arg,
		Type:     column.GoType,
		FK:       column.ForeignEntity,
		Nullable: column.Nullable(),
		Zero:     template.HTML(mfd.MakeZeroValue(column.GoType)),
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
		packed := PackEntity(*entity, nil, name, options)
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

func (n NamespaceData) HasAllOfProvidedEntities(provided []string) bool {
	if len(provided) == 0 {
		return false
	}

	byName := make(map[string]struct{}, len(provided))
	for i := range provided {
		byName[provided[i]] = struct{}{}
	}

	for i := range n.Entities {
		if _, ok := byName[n.Entities[i].Name]; !ok {
			return false
		}
	}

	return true
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

	HasStatus             bool
	HasPKs                bool
	AddIfNotFoundByPKFlow bool
	PKs                   []PKPair
	FillingPKs            []template.HTML

	SortField string
	SortDir   string

	HasRelations              bool
	Relations                 []RelationData
	InitRels                  []template.HTML
	FillingCreatedOrFoundRels []template.HTML

	// Helpers for filling NeedPreparingDependedRelsFromRoot and PreparingDependedRelsFromRoot
	relationByName            map[string]RelationData
	relationNamesHasRelations map[string]struct{}

	HasNestedSameRelations bool
	NestedSameRelations    []string

	NeedFakeFilling bool
	FakeFilling     []template.HTML

	NeedPreparingDependedRelsFromRoot bool
	PreparingDependedRelsFromRoot     []template.HTML

	NeedInitDependedRelsFromRoot bool
	InitDependedRelsFromRoot     []template.HTML

	NeedPreparingFillingSameAsRootRels bool
	PreparingFillingSameAsRootRels     map[string][]template.HTML

	Columns []AttributeData

	HasNotAddable bool
	NotAddable    []string

	HasNotUpdatable bool
	NotUpdatable    []string
}

// PackEntity packs mfd entity to template data
//
//nolint:funlen,gocognit
func PackEntity(entity mfd.Entity, parentEntity *model.EntityData, namespace string, options Options) EntityData {
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

		// Calculate if the column relates to itself
		var hasSameRel bool
		if column.ForeignEntity != nil {
			for _, rel := range te.Relations {
				if rel.Type == column.GoType {
					hasSameRel = true
				}
			}
		}

		// Filling OpFunc which generates fake data
		if !column.Nullable() && !column.PrimaryKey && !hasSameRel {
			// Check if the column has a known field name
			byFieldName, ok := fakeFiller.ByNameAndType(column.Name, column.GoType, column.Max)
			if ok {
				// If it is, it generates more for the field data
				condition := mustWrapFilling("in."+column.Name, column.GoType, template.HTML(mfd.MakeZeroValue(column.GoType)), byFieldName, column.IsArray, false, false)
				fakeFillingData = append(fakeFillingData, condition)
			} else if byType, found := fakeFiller.ByType(column.Name, column.GoType, column.DBType, column.IsArray, column.Max); found {
				// By default, generates something depending on a field type
				condition := mustWrapFilling("in."+column.Name, column.GoType, template.HTML(mfd.MakeZeroValue(column.GoType)), byType, column.IsArray, false, false)
				fakeFillingData = append(fakeFillingData, condition)
			}
		}
	}

	imports.Append(fakeFiller.Imports()...)

	// store all relation names for join field
	relNames := make([]RelationData, 0, len(te.Relations))
	relNamesMap := make(map[string]RelationData, len(te.Relations))
	sameRelNamesMap := make(map[string]RelationData, len(te.Relations))
	relNamesWhichHasRels := make(map[string]struct{}, len(te.Relations))
	for i := range te.Relations {
		if te.Relations[i].Entity.Name == te.Name {
			continue
		}

		relationData := PackRelationData(te.Relations[i], &te, namespace, options)
		relNames = append(relNames, relationData)
		relNamesMap[relationData.Name] = relationData

		// Check what if we have the same relations as a parent has
		if parentEntity != nil && parentEntity.HasRelations {
			for _, rel := range parentEntity.Relations {
				if rel.Name == relationData.Name {
					sameRelNamesMap[relationData.Name] = relationData
				}
			}
		}

		if relationData.Entity.HasRelations {
			relNamesWhichHasRels[relationData.Name] = struct{}{}
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

		HasStatus:             hasStatus,
		PKs:                   pks,
		HasPKs:                len(pks) > 0,
		AddIfNotFoundByPKFlow: te.AreNotNullablePKs(),

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

	res.PreparingFillingSameAsRootRels = make(map[string][]template.HTML)
	res.InitDependedRelsFromRoot, res.PreparingDependedRelsFromRoot = walkThroughDependedEntities(res.Relations, res, "", "", make(map[string]struct{}), res.PreparingFillingSameAsRootRels)
	res.NeedPreparingDependedRelsFromRoot = len(res.PreparingDependedRelsFromRoot) > 0
	res.NeedInitDependedRelsFromRoot = len(res.InitDependedRelsFromRoot) > 0
	res.InitRels = initRels(relNamesMap)
	res.NeedPreparingFillingSameAsRootRels = len(res.PreparingFillingSameAsRootRels) > 0
	res.fillingRels(te, relNamesMap)
	res.fillingRelPKs(relNamesMap)

	return res
}

var (
	firstEntity = regexp.MustCompile(`^\.[^.]+`)
)

// walkThroughDependedEntities Walks through all relations of current relations recursively and finds same relations.
// Returns prepared strings to inject in a layout.
// The first value is inici
// The second one is filling root PKs to same nested relations.
func walkThroughDependedEntities(curRels []RelationData, root EntityData, embeddedRels, embeddedRelTypes string, alreadyPrepared map[string]struct{}, assignNestedRelsByRel map[string][]template.HTML) (initNestedRels, fillNestedRels []template.HTML) {
	rootRelName := firstEntity.FindString(embeddedRels)

	var hasAlreadyPrepared bool
	for _, curEntity := range curRels {
		init, filling := walkThroughDependedEntities(curEntity.Entity.Relations, root, embeddedRels+"."+curEntity.Name, embeddedRels+"."+curEntity.Type, alreadyPrepared, assignNestedRelsByRel)
		initNestedRels = append(initNestedRels, init...) // Fill from the end to the start
		fillNestedRels = append(fillNestedRels, filling...)

		if rootRel, ok := root.relationByName[curEntity.Name]; ok && embeddedRels != "" {
			for _, pk := range curEntity.Entity.PKs {
				needAmpersand := curEntity.NilCheck && !rootRel.NilCheck
				needVal := !curEntity.NilCheck && rootRel.NilCheck
				switch {
				case needAmpersand:
					fillNestedRels = append(fillNestedRels, template.HTML(fmt.Sprintf("in%[1]s.%[2]s%[3]s = &in.%[2]s%[3]s", embeddedRels, curEntity.Name, pk.Field)))
				case needVal:
					fillNestedRels = append(fillNestedRels, template.HTML(fmt.Sprintf("in%[1]s.%[2]s%[3]s = val(in.%[2]s%[3]s)", embeddedRels, curEntity.Name, pk.Field)))
				default:
					fillNestedRels = append(fillNestedRels, template.HTML(fmt.Sprintf("in%[1]s.%[2]s%[3]s = in.%[2]s%[3]s", embeddedRels, curEntity.Name, pk.Field)))
				}
			}
		}

		if embeddedRels != "" {
			if !hasAlreadyPrepared {
				// Split the chain of relations by dots. We need to extract the last element
				relsChain := strings.Split(embeddedRelTypes, ".")
				if len(relsChain) > 2 {
					zero := fmt.Sprintf("&db.%s{}", relsChain[len(relsChain)-1])
					assign := fmt.Sprintf("in%s = %s", embeddedRels, zero)
					str := mustWrapFilling("in"+embeddedRels, "nil", "nil", template.HTML(assign), false, false, false)
					initNestedRels = append([]template.HTML{str}, initNestedRels...) // Fill from the end to the start
					hasAlreadyPrepared = true
				}
			}

			if _, ok := alreadyPrepared[curEntity.Name]; !ok {
				chainsToFill := walkThroughRels(root.Relations, embeddedRels, "", curEntity.Name)
				alreadyPrepared[curEntity.Name] = struct{}{}
				skipFirstRel := strings.Replace(embeddedRels, rootRelName, "", 1)
				for _, chain := range chainsToFill {
					cleaned := strings.TrimLeft(rootRelName, ".")
					assignNestedRelsByRel[cleaned] = append(assignNestedRelsByRel[cleaned], template.HTML(fmt.Sprintf("in%[1]s.%[2]s = rel%[3]s.%[2]s", chain, curEntity.Name, skipFirstRel)))
				}
			}
		}
	}

	return
}

func walkThroughRels(curRels []RelationData, curEmbeddedPosition, embeddedRels, targetEntityName string) (res []string) {
	for _, curEntity := range curRels {
		res = append(res, walkThroughRels(curEntity.Entity.Relations, curEmbeddedPosition, embeddedRels+"."+curEntity.Name, targetEntityName)...)

		if curEmbeddedPosition == embeddedRels {
			continue
		}

		if targetEntityName == curEntity.Name {
			res = append(res, embeddedRels)
		}
	}

	return
}

func initRels(relByName map[string]RelationData) []template.HTML {
	res := make([]template.HTML, 0, len(relByName))
	for relName, rel := range relByName {
		zero := fmt.Sprintf("&db.%s{}", rel.Type)
		assign := fmt.Sprintf("in.%s = %s", relName, zero)
		str := mustWrapFilling("in."+relName, "nil", "nil", template.HTML(assign), false, false, false)
		res = append(res, str)
	}

	slices.Sort(res)

	return res
}

func (e *EntityData) fillingRelPKs(relByName map[string]RelationData) {
	for _, rel := range e.Relations {
		for _, pk := range relByName[rel.Name].Entity.PKs {
			var res string
			needVal := rel.NilCheck
			fieldName := rel.Name + pk.Field
			for _, origPK := range e.PKs {
				if origPK.FK != nil && origPK.FK.Name == rel.Name {
					fieldName = pk.Field
					needVal = false // Consider that its pk as fk is not nil
				}
			}
			switch {
			case needVal:
				res = fmt.Sprintf("in.%s.%s = val(in.%s)", rel.Name, pk.Field, fieldName)
			default:
				res = fmt.Sprintf("in.%s.%s = in.%s", rel.Name, pk.Field, fieldName)
			}

			e.FillingPKs = append(e.FillingPKs, mustWrapFilling("in."+fieldName, pk.Type, pk.Zero, template.HTML(res), false, needVal, true))
		}
	}
}

func (e *EntityData) fillingRels(rawEntity model.EntityData, relNamesMap map[string]RelationData) {
	byRelName := make(map[string][]template.HTML, len(relNamesMap)*2)
	for _, r := range rawEntity.Relations {
		rel, ok := relNamesMap[r.Name]
		if !ok {
			continue
		}

		byRelName[r.Name] = append(byRelName[r.Name], template.HTML(fmt.Sprintf("in.%s = rel", r.Name)))
		for _, pk := range rel.Entity.PKs {
			fieldName := rel.Name + pk.Field
			needAmpersand := rel.NilCheck
			for _, origPK := range e.PKs {
				if origPK.FK != nil && origPK.FK.Name == rel.Name {
					fieldName = pk.Field
					needAmpersand = false // Consider that its pk as fk is not nil
				}
			}
			switch {
			case needAmpersand:
				byRelName[r.Name] = append(byRelName[r.Name], template.HTML(fmt.Sprintf("in.%s = &rel.%s", fieldName, pk.Field)))
			default:
				byRelName[r.Name] = append(byRelName[r.Name], template.HTML(fmt.Sprintf("in.%s = rel.%s", fieldName, pk.Field)))
			}
		}
	}

	for i, rel := range e.Relations {
		e.Relations[i].Entity.FillingCreatedOrFoundRels = byRelName[rel.Name]
	}
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

func PackRelationData(in model.RelationData, parentEntity *model.EntityData, namespace string, options Options) RelationData {
	res := RelationData{
		Name:     in.Name,
		Type:     in.Type,
		VarName:  mfd.VarName(in.Name),
		Tag:      in.Tag,
		Comment:  in.Comment,
		NilCheck: in.Nullable,
	}

	if in.ForeignEntity != nil {
		res.Entity = PackEntity(*in.Entity, parentEntity, namespace, Options{GoPGVer: options.GoPGVer})
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

type FuncLayoutRenderer interface {
	Render(w io.Writer, data any) error
}

type MainFunc struct{}

func (op MainFunc) Render(w io.Writer, data any) error {
	return loadAndParseTemplate(w, funcTemplate, data)
}

type OpFuncWithRelations struct{}

func (op OpFuncWithRelations) Render(w io.Writer, data any) error {
	return loadAndParseTemplate(w, funcOpWithRelTemplate, data)
}

type OpFuncType struct{}

func (op OpFuncType) Render(w io.Writer, data any) error {
	return loadAndParseTemplate(w, opFuncTypeTemplate, data)
}

type OpFuncWithFake struct{}

func (op OpFuncWithFake) Render(w io.Writer, data any) error {
	return loadAndParseTemplate(w, funcOpWithFakeTemplate, data)
}

func loadAndParseTemplate(w io.Writer, tmpl string, data any) error {
	return mfd.Render(w, tmpl, data)
}
