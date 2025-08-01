package db

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/types"
)

type jsonField struct {
	DBName      string
	FullPath    string
	LastElement string
}

// prepareJSON prepares SQL where-condition for json field filtering
func (f Filter) prepareJSON(st string) (field, value types.ValueAppender) {
	jf := f.jsonField(f.Field)
	switch f.SearchType {
	case SearchTypeArrayContains:
		if f.Exclude {
			jf.DBName = "not " + jf.DBName
		}
		st = fmt.Sprintf(`@> '{"%s": [%v]}'`, jf.LastElement, f.jsonArrayValue(f.Value))
		return pg.Safe(jf.DBName), pg.Safe(st)
	case SearchTypeEquals, SearchTypeArray:
		st = searchTypes[f.Exclude][SearchTypeArray]
		f.Value = pg.In(f.jsonValue(f.Value))
	}
	return pg.Safe(jf.FullPath), pg.SafeQuery(st, f.Value)
}

// jsonField prepares json/jsonb field name for postgresql json filters
func (f Filter) jsonField(field string) jsonField {
	var (
		result jsonField
	)
	str := strings.Split(field, "->")
	for i, elem := range str {
		sep := "->"
		// last element use ->>
		if elem == str[len(str)-1] {
			sep = "->>"
			result.LastElement = elem
		}
		// first element must be wrapped in double quotes
		if i == 0 {
			sep = ""
			elem = `"` + strings.Join(strings.Split(elem, "."), `"."`) + `"`
			result.DBName = elem
		} else {
			elem = "'" + elem + "'"
		}
		result.FullPath += sep + elem
	}
	return result
}

// jsonValue convert json field value to []string
func (f Filter) jsonValue(value interface{}) []string {
	var res []string

	switch v := value.(type) {
	case bool:
		return []string{strconv.FormatBool(v)}
	case int:
		return []string{strconv.Itoa(v)}
	case int64:
		return []string{strconv.FormatInt(v, 10)}
	case uint:
		return []string{strconv.FormatUint(uint64(v), 10)}
	case uint64:
		return []string{strconv.FormatUint(v, 10)}
	case float64, float32:
		return []string{fmt.Sprintf("%f", v)}
	case string:
		return []string{v}
	case []int:
		for _, k := range v {
			res = append(res, strconv.Itoa(k))
		}
		return res
	case []int64:
		for _, k := range v {
			res = append(res, strconv.FormatInt(k, 10))
		}
		return res
	case []uint:
		for _, k := range v {
			res = append(res, strconv.FormatUint(uint64(k), 10))
		}
		return res
	case []uint64:
		for _, k := range v {
			res = append(res, strconv.FormatUint(k, 10))
		}
		return res
	case []string:
		return v
	case []float64:
		for _, k := range v {
			res = append(res, fmt.Sprintf("%f", k))
		}
		return res
	case []float32:
		for _, k := range v {
			res = append(res, fmt.Sprintf("%f", k))
		}
		return res
	case []bool:
		for _, k := range v {
			res = append(res, strconv.FormatBool(k))
		}
		return res
	default:
		return []string{fmt.Sprint(v)}
	}
}

// jsonArrayValue convert json field value to string
func (f Filter) jsonArrayValue(value interface{}) string {
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float64, float32:
		return fmt.Sprintf("%f", v)
	case string:
		return strconv.Quote(v)
	default:
		return strconv.Quote(fmt.Sprint(v))
	}
}
