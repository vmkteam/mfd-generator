package mfd

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/dizzyfool/genna/util"
	"github.com/jinzhu/inflection"
)

var mfdReserved = map[string]struct{}{
	"Columns":       {},
	"Tables":        {},
	"Searcher":      {},
	"ErrEmptyValue": {},
	"ErrMaxLength":  {},
	"ErrWrongValue": {},
	"Status":        {},
	"OpFunc":        {},
}

func MakeSearchName(name string, searchType SearchType) string {
	switch searchType {
	case SearchEquals:
		return fmt.Sprintf("%sEq", name)
	case SearchArray:
		return MakePlural(name)
	case SearchG:
		return fmt.Sprintf("%sG", name)
	case SearchGE:
		return fmt.Sprintf("%sGE", name)
	case SearchL:
		return fmt.Sprintf("%sL", name)
	case SearchLE:
		return fmt.Sprintf("%sLE", name)
	case SearchILike:
		return fmt.Sprintf("%sILike", name)
	case SearchLike:
		return fmt.Sprintf("%sLike", name)
	case SearchLeftILike:
		return fmt.Sprintf("%sLILike", name)
	case SearchLeftLike:
		return fmt.Sprintf("%sLLike", name)
	case SearchRightILike:
		return fmt.Sprintf("%sRILike", name)
	case SearchRightLike:
		return fmt.Sprintf("%sRLike", name)
	case SearchNotArray:
		return fmt.Sprintf("Not%s", MakePlural(name))
	case SearchNotEquals:
		return fmt.Sprintf("Not%s", name)
	case SearchNull:
		return fmt.Sprintf("%sNull", name)
	case SearchNotNull:
		return fmt.Sprintf("%sNotNull", name)
	}

	return name
}

func MakePlural(name string) string {
	if strings.HasSuffix(name, util.ID) {
		return name + "s"
	}

	if plural := inflection.Plural(name); plural != name {
		return plural
	}

	if !strings.HasSuffix(name, "s") {
		return name + "s"
	}

	return name
}

func JSONName(name string) string {
	return VarName(name)
}

func VarName(name string) string {
	name = util.ReplaceSuffix(name, util.ID, util.Id)
	name = util.ReplaceSuffix(name, util.IDs, util.Ids)

	return util.LowerFirst(name)
}

func ShortVarName(name string) string {
	var r []byte
	for i := 0; i < len(name); i++ {
		c := name[i]
		if i == 0 {
			r = append(r, util.ToLower(c))
		} else if util.IsUpper(c) && util.IsLower(name[i-1]) {
			r = append(r, util.ToLower(c))
		}
	}

	for i := 0; i < len(r); i++ {
		if varName := string(r[i:]); !IsReserved(varName) {
			return varName
		}
	}

	return VarName(name)
}

func URLName(name string) string {
	return strings.ReplaceAll(util.Underscore(name), "_", "-")
}

func FKName(name string) string {
	return util.ReplaceSuffix(util.ColumnName(name), util.ID, "")
}

func IsReserved(name string) bool {
	return token.Lookup(strings.ToLower(name)).IsKeyword()
}

func IsReservedByMFD(name string) bool {
	_, ok := mfdReserved[name]
	return ok
}
