package dbtest

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/model"
)

const (
	maxWordCount       = 10
	maxWordLen         = 10
	minSentenceLen     = 30
	defaultSentenceLen = 100
)

type FakeFiller struct {
	imports map[string]struct{}
}

func NewFakeFiller() FakeFiller { return FakeFiller{imports: make(map[string]struct{})} }

// ByNameAndType Checks column name if it is a known name with a special fake func substitution
//
//nolint:funlen
func (ff FakeFiller) ByNameAndType(columnName, gotype string, maxFiledLen int) (res template.HTML, found bool) {
	switch columnName {
	case "StatusID":
		//nolint:gocritic
		switch gotype {
		case model.TypeInt, model.TypeInt32, model.TypeInt64, model.TypeFloat32, model.TypeFloat64:
			return FakeIt("1").assign(columnName).Tmpl(), true
		}

		return "", false
	case "Phone":
		switch gotype {
		case model.TypeInt, model.TypeInt32, model.TypeInt64, model.TypeFloat32, model.TypeFloat64:
			ff.imports["strconv"] = struct{}{}
			return template.HTML(fmt.Sprintf("in.Phone, _ = strconv.Atoi(%s)", fakePhone.cutString(maxFiledLen))), true
		case model.TypeString:
			return fakePhone.cutString(maxFiledLen).assign(columnName).Tmpl(), true
		}

		return "", false
	case "Alias":
		//nolint:gocritic
		switch gotype {
		case model.TypeString:
			ff.imports["strings"] = struct{}{}
			switch {
			case maxFiledLen == 0:
				return fakeEmpty.sentence(defaultSentenceLen).cutString(maxFiledLen).replaceAll(" ", "-").assign(columnName).Tmpl(), true
			case maxFiledLen >= minSentenceLen:
				return fakeEmpty.sentence(maxFiledLen).cutString(maxFiledLen).replaceAll(" ", "-").assign(columnName).Tmpl(), true
			}

			return fakeWord.cutString(maxFiledLen).assign(columnName).Tmpl(), true
		}

		return "", false
	case "Email":
		//nolint:gocritic
		switch gotype {
		case model.TypeString:
			return fakeEmail.cutString(maxFiledLen).assign(columnName).Tmpl(), true
		}

		return "", false
	case "Login":
		switch gotype {
		case model.TypeInt, model.TypeInt32, model.TypeInt64, model.TypeFloat32, model.TypeFloat64:
			return fakeIntRange.assign(columnName).Tmpl(), true
		case model.TypeString:
			return fakeWord.cutString(maxFiledLen).assign(columnName).Tmpl(), true
		}

		return "", false
	case "Password":
		//nolint:gocritic
		switch gotype {
		case model.TypeString:
			return fakePassword.cutString(maxFiledLen).assign(columnName).Tmpl(), true
		}

		return "", false
	case "CreatedAt":
		switch gotype {
		case model.TypeTime:
			ff.imports["time"] = struct{}{}
			return fakeNow.assign(columnName).Tmpl(), true
		case model.TypeString:
			ff.imports["time"] = struct{}{}
			return fakeNow.formatRFC3339().assign(columnName).Tmpl(), true
		}

		return "", false
	case "ModifiedAt", "ModifiedDate", "ModifyAt", "ModifyDate",
		"UpdatedAt", "UpdatedDate", "UpdateAt", "UpdateDate",
		"StartedAt", "StartedDate", "StartAt", "StartDate",
		"DeletedAt", "DeletedDate", "DeleteAt", "DeleteDate",
		"PublishedAt", "PublishedDate", "PublishDate", "PublishAt":
		switch gotype {
		case model.TypeTime:
			ff.imports["time"] = struct{}{}
			return fakeRangeDateFuture.assign(columnName).Tmpl(), true
		case model.TypeString:
			ff.imports["time"] = struct{}{}
			return fakeRangeDateFuture.formatRFC3339().assign(columnName).Tmpl(), true
		}

		return "", false
	}

	return "", false
}

func (ff FakeFiller) ByType(columnName, gotype string, isArray bool, maxFiledLen int) (res template.HTML, found bool) {
	switch gotype {
	case model.TypeInt, model.TypeInt32, model.TypeInt64:
		return fakeIntRange.assign(columnName).Tmpl(), true
	case model.TypeFloat32:
		return fakeFloat32Range.assign(columnName).Tmpl(), true
	case model.TypeFloat64:
		return fakeFloat64Range.assign(columnName).Tmpl(), true
	case model.TypeString:
		switch {
		case maxFiledLen == 0:
			return fakeEmpty.sentence(defaultSentenceLen).cutString(maxFiledLen).assign(columnName).Tmpl(), true
		case maxFiledLen >= minSentenceLen:
			return fakeEmpty.sentence(maxFiledLen).cutString(maxFiledLen).assign(columnName).Tmpl(), true
		}
		return fakeWord.cutString(maxFiledLen).assign(columnName).Tmpl(), true
	case model.TypeByteSlice:
		switch {
		case maxFiledLen == 0:
			return fakeEmpty.sentence(defaultSentenceLen).cutBytes(maxFiledLen).assign(columnName).Tmpl(), true
		case maxFiledLen >= minSentenceLen:
			return fakeEmpty.sentence(maxFiledLen).cutBytes(maxFiledLen).assign(columnName).Tmpl(), true
		}
		return fakeWord.cutBytes(maxFiledLen).assign(columnName).Tmpl(), true
	case model.TypeBool:
		return fakeBool.assign(columnName).Tmpl(), true
	case model.TypeTime:
		ff.imports["time"] = struct{}{}
		return fakeRangeDateFuture.assign(columnName).Tmpl(), true
	case model.TypeDuration:
		return FakeIt(fmt.Sprintf("gofakeit.IntRange(%d, %d)", time.Second.Nanoseconds(), (24 * time.Hour).Nanoseconds())).assign(columnName).Tmpl(), true
	case model.TypeMapInterface:
		return FakeIt("map[string]interface{}{gofakeit.InputName(): gofakeit.Word()}").assign(columnName).Tmpl(), true
	case model.TypeMapString:
		return FakeIt("map[string]string{gofakeit.InputName(): gofakeit.Word()}").assign(columnName).Tmpl(), true
	case model.TypeIP:
		ff.imports["net"] = struct{}{}
		return fakeEmpty.ipv4().assign(columnName).Tmpl(), true
	case model.TypeIPNet:
		ff.imports["net"] = struct{}{}
		return fakeEmpty.ipv4Net().assign(columnName).Tmpl(), true
	case model.TypeInterface:
		return fakeWord.cutString(maxFiledLen).assign(columnName).Tmpl(), true
	}

	if isArray {
		return FakeIt(gotype + "{}").assign(columnName).Tmpl(), true
	}

	// By default, we don't know what the type and what package it belongs.
	// Skip it, because the original struct has already defaulted value.
	return "", false
}

func (ff FakeFiller) Imports() []string {
	res := make([]string, 0, len(ff.imports))
	for i := range ff.imports {
		res = append(res, i)
	}

	return res
}

const (
	fakeEmpty           FakeIt = ""
	fakeIntRange        FakeIt = "gofakeit.IntRange(1, 10)"
	fakeFloat32Range    FakeIt = "gofakeit.Float32Range(1, 10)"
	fakeFloat64Range    FakeIt = "gofakeit.Float64Range(1, 10)"
	fakeByte            FakeIt = "byte(gofakeit.UintRange(0, 255))"
	fakeBool            FakeIt = "gofakeit.Bool()"
	fakeWord            FakeIt = "gofakeit.Word()"
	fakePhone           FakeIt = "gofakeit.Phone()"
	fakeEmail           FakeIt = "gofakeit.Email()"
	fakePassword        FakeIt = "gofakeit.Password(true, true, true, false, false, 12)"
	fakeNow             FakeIt = "time.Now()"
	fakeRangeDateFuture FakeIt = "gofakeit.DateRange(time.Now().Add(5*time.Minute), time.Now().Add(1*time.Hour))"
)

type FakeIt string

func (fi FakeIt) String() string {
	return string(fi)
}

func (fi FakeIt) sentence(maxFiledLen int) FakeIt {
	return FakeIt(fmt.Sprintf("gofakeit.Sentence(%d)", min(maxWordCount, maxFiledLen/maxWordLen)))
}

func (fi FakeIt) ipv4() FakeIt {
	return FakeIt(fmt.Sprintf("net.IPv4(%[1]s, %[1]s, %[1]s, %[1]s", fakeByte))
}

func (fi FakeIt) ipv4Net() FakeIt {
	return FakeIt(fmt.Sprintf("net.IPNet{IP: %s, Mask: net.IPv4Mask(255, 255, 255, 0)}", fakeEmpty.ipv4()))
}

func (fi FakeIt) cutString(maxFiledLen int) FakeIt {
	return FakeIt(fmt.Sprintf("cutS(%s, %d)", fi, maxFiledLen))
}

func (fi FakeIt) cutBytes(maxFiledLen int) FakeIt {
	return FakeIt(fmt.Sprintf("cutB(%s, %d)", fi, maxFiledLen))
}

func (fi FakeIt) formatRFC3339() FakeIt {
	return FakeIt(fmt.Sprintf("%s.Format(time.RFC3339)", fi))
}

func (fi FakeIt) replaceAll(from, to string) FakeIt {
	return FakeIt(fmt.Sprintf("strings.ReplaceAll(%s, \"%s\", \"%s\")", fi, from, to))
}

func (fi FakeIt) assign(columnName string) FakeIt {
	return FakeIt(fmt.Sprintf("in.%s = %s", columnName, fi))
}

func (fi FakeIt) Tmpl() template.HTML {
	return template.HTML(fi.String())
}

const (
	wrapperTmpl = `
	if {{.Condition}} {
		{{.Filling}}
	}
`
)

type conditionData struct {
	Name string
	Zero template.HTML
}

type wrapperTemplateData struct {
	Condition template.HTML
	Filling   template.HTML
}

func mustWrapFilling(columnName, goType string, zeroVal, filling template.HTML, isArray bool) template.HTML {
	if isArray {
		zeroVal = "nil"
	}

	condition := "{{.Name}} == {{.Zero}}"
	//nolint:gocritic
	switch goType {
	case model.TypeTime:
		condition = "{{.Name}}.IsZero()"
	}

	var conditionBuf bytes.Buffer
	err := mfd.Render(&conditionBuf, condition, conditionData{columnName, zeroVal})
	if err != nil {
		panic(fmt.Errorf("cannot make a condition, column=%s, GoType=%s, err=%w", columnName, zeroVal, err))
	}

	var wrapperBuff bytes.Buffer
	err = mfd.Render(&wrapperBuff, wrapperTmpl, wrapperTemplateData{Condition: template.HTML(conditionBuf.String()), Filling: filling})
	if err != nil {
		panic(fmt.Errorf("cannot wrap, column=%s, GoType=%s, err=%w", columnName, zeroVal, err))
	}

	return template.HTML(wrapperBuff.String())
}
