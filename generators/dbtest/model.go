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
	Package        string
	DBPackage      string
	DBPackageAlias string

	ProjectName string
	GoPGVer     string
}

// PackFuncRenderData packs mfd namespace to template data
func PackFuncRenderData(options Options) FuncFileRenderData {
	var goPGVer string
	if options.GoPGVer != mfd.GoPG8 {
		goPGVer = fmt.Sprintf("/v%d", options.GoPGVer)
	}
	pkgParts := strings.Split(options.DBPackage, "/")
	return FuncFileRenderData{
		GoPGVer:        goPGVer,
		Package:        options.Package,
		DBPackage:      options.DBPackage,
		DBPackageAlias: pkgParts[len(pkgParts)-1],
		ProjectName:    options.ProjectName,
	}
}

// PKPair stores primary keys with type for template
type PKPair struct {
	Field    string
	Arg      string
	Type     string
	FK       *mfd.Entity
	Nullable bool
	IsCustom bool
	Zero     template.HTML
}

func PackPKPair(column model.AttributeData) PKPair {
	arg := util.LowerFirst(column.Name)
	if column.Name == util.ID {
		arg = "id"
	}

	zv, found := mfd.MakeZeroValue2(column.GoType)

	return PKPair{
		Field:    column.Name,
		Arg:      arg,
		Type:     column.GoType,
		FK:       column.ForeignEntity,
		Nullable: column.Nullable(),
		Zero:     template.HTML(zv),
		IsCustom: !found,
	}
}

// SortPair stores sort columns with direction for template
type SortPair struct {
	Field, Dir string
}

// NamespaceData stores namespace info for template
type NamespaceData struct {
	Package        string
	DBPackage      string
	DBPackageAlias string

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
	pkgParts := strings.Split(options.DBPackage, "/")

	return NamespaceData{
		Package:        options.Package,
		DBPackage:      options.DBPackage,
		DBPackageAlias: pkgParts[len(pkgParts)-1],

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

	VarName        string
	VarNamePlural  string
	DBPackageAlias string

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
	relationByName map[string]RelationData

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
func PackEntity(entity mfd.Entity, namespace string, options Options, previous ...string) EntityData {
	// base template entity - repo depends on int
	te := model.PackEntity(entity, model.Options{ArrayAsRelation: true})

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
		if !column.Nullable() && !column.PrimaryKey && column.ForeignKey == "" && !hasSameRel {
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
	relByNamesMap := make(map[string]RelationData, len(te.Relations))
	relByNamesTypesMap := make(map[string]RelationData, len(te.Relations))
	for i := range te.Relations {
		if te.Relations[i].Entity.Name == te.Name {
			continue
		}

		// Skip if the same entity appeared before
		if slices.Contains(previous, te.Relations[i].Entity.Name) {
			continue
		}

		// Skip if a relation can be nil
		if te.Relations[i].Nullable {
			continue
		}

		relationData := PackRelationData(te.Relations[i], namespace, options, append(previous, te.Name)...)
		relNames = append(relNames, relationData)
		relByNamesMap[relationData.Name] = relationData
		relByNamesTypesMap[relationData.Name] = relationData
		if relationData.IsArray {
			relByNamesTypesMap[relationData.Type] = relationData
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

	pkgParts := strings.Split(options.DBPackage, "/")
	res := EntityData{
		Name:       te.Name,
		NamePlural: goNamePlural,

		Namespace: namespace,

		HasImports: imports.Len() > 0,
		Imports:    imports.Elements(),

		VarName:        varName,
		VarNamePlural:  varNamePlural,
		DBPackageAlias: pkgParts[len(pkgParts)-1],

		HasStatus:             hasStatus,
		PKs:                   pks,
		HasPKs:                len(pks) > 0,
		AddIfNotFoundByPKFlow: te.AreNotNullablePKs(),

		SortField: sortField,
		SortDir:   sortDir,

		Relations:      relNames,
		HasRelations:   len(relNames) > 0,
		relationByName: relByNamesTypesMap,

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
	res.InitRels = initRels(relByNamesMap, res.DBPackageAlias)
	res.NeedPreparingFillingSameAsRootRels = len(res.PreparingFillingSameAsRootRels) > 0
	res.fillingRels(te, relByNamesMap)
	res.fillingRelPKs(relByNamesMap)

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
		if curEntity.IsArray {
			continue
		}
		init, filling := walkThroughDependedEntities(curEntity.Entity.Relations, root, embeddedRels+"."+curEntity.Name, embeddedRels+"."+curEntity.Type, alreadyPrepared, assignNestedRelsByRel)
		initNestedRels = append(initNestedRels, init...) // Fill from the end to the start
		fillNestedRels = append(fillNestedRels, filling...)

		if rootRel, ok := root.relationByName[curEntity.Name]; ok && embeddedRels != "" {
			fillNestedRels = append(fillNestedRels, prepareFillingConsideringIsArr(rootRel, curEntity, embeddedRels)...)
		}

		if embeddedRels != "" {
			if !hasAlreadyPrepared {
				// Split the chain of relations by dots. We need to extract the last element
				relsChain := strings.Split(embeddedRelTypes, ".")
				if len(relsChain) > 2 {
					zero := fmt.Sprintf("&%s.%s{}", root.DBPackageAlias, relsChain[len(relsChain)-1])
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
		// Skip arrays
		if curEntity.IsArray {
			continue
		}

		// Skip same assignment
		if curEmbeddedPosition == embeddedRels {
			continue
		}

		res = append(res, walkThroughRels(curEntity.Entity.Relations, curEmbeddedPosition, embeddedRels+"."+curEntity.Name, targetEntityName)...)

		if targetEntityName == curEntity.Name {
			res = append(res, embeddedRels)
		}
	}

	return
}

func prepareFillingConsideringIsArr(rootRel, curRel RelationData, embeddedRels string) (res []template.HTML) {
	if curRel.IsArray && rootRel.IsArray {
		return []template.HTML{template.HTML(fmt.Sprintf("in%[1]s.%[2]s = &in.%[2]s", embeddedRels, curRel.Name))}
	}

	if curRel.IsArray && !rootRel.IsArray {
		if curRel.NilCheck {
			return nil
		}

		for _, pk := range rootRel.Entity.PKs {
			res = append(res, template.HTML(fmt.Sprintf("in%[1]s.%[2]s = append(in%[1]s.%[2]s, in.%[3]s%[4]s)", embeddedRels, curRel.Name, rootRel.Name, pk.Field)))
		}

		return
	}

	if !curRel.IsArray && rootRel.IsArray {
		pk := curRel.Entity.PKs[0]
		needAmpersand := rootRel.NilCheck
		var assign string
		switch {
		case needAmpersand:
			assign = fmt.Sprintf("in%[1]s.%[2]s%[3]s = &in.%[4]s[0]", embeddedRels, curRel.Type, pk.Field, rootRel.Name)
		default:
			assign = fmt.Sprintf("in%[1]s.%[2]s%[3]s = in.%[4]s[0]", embeddedRels, curRel.Type, pk.Field, rootRel.Name)
		}

		condition := mustWrapFilling("in."+rootRel.Name, "nil", "0", template.HTML(assign), true, false, true)
		return []template.HTML{condition}
	}

	for _, pk := range curRel.Entity.PKs {
		needAmpersand := curRel.NilCheck && !rootRel.NilCheck
		needVal := !curRel.NilCheck && rootRel.NilCheck
		switch {
		case needAmpersand:
			res = append(res, template.HTML(fmt.Sprintf("in%[1]s.%[2]s%[3]s = &in.%[2]s%[3]s", embeddedRels, curRel.Name, pk.Field)))
		case needVal:
			res = append(res, template.HTML(fmt.Sprintf("in%[1]s.%[2]s%[3]s = val(in.%[2]s%[3]s)", embeddedRels, curRel.Name, pk.Field)))
		default:
			res = append(res, template.HTML(fmt.Sprintf("in%[1]s.%[2]s%[3]s = in.%[2]s%[3]s", embeddedRels, curRel.Name, pk.Field)))
		}
	}

	return append(res, template.HTML(fmt.Sprintf("in%[1]s.%[2]s = in.%[2]s", embeddedRels, curRel.Name)))
}

func initRels(relByName map[string]RelationData, dbPkgAlias string) []template.HTML {
	res := make([]template.HTML, 0, len(relByName))
	for relName, rel := range relByName {
		zero := fmt.Sprintf("&%s.%s{}", dbPkgAlias, rel.Type)
		if rel.IsArray {
			zero = fmt.Sprintf("%s{}", rel.GoType)
		}
		assign := fmt.Sprintf("in.%s = %s", relName, zero)
		str := mustWrapFilling("in."+relName, "nil", "nil", template.HTML(assign), false, false, false)
		res = append(res, str)
	}

	slices.Sort(res)

	return res
}

func (e *EntityData) fillingRelPKs(relByName map[string]RelationData) {
	for _, rel := range e.Relations {
		if rel.IsArray {
			continue
		}

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

		if rel.IsArray {
			byRelName[r.Name] = append(byRelName[r.Name], template.HTML(fmt.Sprintf("in.%s = append(in.%s, rel.%s)", r.Name, r.Name, rel.Entity.PKs[0].Field)))
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
	GoType   string
	VarName  string
	NilCheck bool
	IsArray  bool
	Entity   EntityData

	Tag     template.HTML
	Comment template.HTML
}

func PackRelationData(in model.RelationData, namespace string, options Options, previous ...string) RelationData {
	res := RelationData{
		Name:     in.Name,
		Type:     in.Type,
		GoType:   in.GoType,
		VarName:  mfd.VarName(in.Name),
		Tag:      in.Tag,
		Comment:  in.Comment,
		NilCheck: in.Nullable,
		IsArray:  in.IsArray,
	}

	if in.ForeignEntity != nil {
		res.Entity = PackEntity(*in.Entity, namespace, Options{GoPGVer: options.GoPGVer}, previous...)
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
