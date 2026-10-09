package dbtest

import (
	"go/parser"
	"testing"

	"github.com/dizzyfool/genna/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFakeFiller_IntCast(t *testing.T) {
	Convey("int32 and int64 columns are cast before assignment", t, func() {
		ff := NewFakeFiller()

		cases := []struct {
			name, goType, want string
			byName             bool
		}{
			{"Count", model.TypeInt32, "in.Count = int32(gofakeit.IntRange(1, 10))", false},
			{"Count", model.TypeInt64, "in.Count = int64(gofakeit.IntRange(1, 10))", false},
			{"Login", model.TypeInt32, "in.Login = int32(gofakeit.IntRange(1, 10))", true},
			{"Login", model.TypeInt64, "in.Login = int64(gofakeit.IntRange(1, 10))", true},
		}
		for _, c := range cases {
			var (
				got   string
				found bool
			)
			if c.byName {
				res, ok := ff.ByNameAndType(c.name, c.goType, 0)
				got, found = string(res), ok
			} else {
				res, ok := ff.ByType(c.name, c.goType, model.TypePGInt8, false, 0)
				got, found = string(res), ok
			}

			So(found, ShouldBeTrue)
			So(got, ShouldEqual, c.want)

			_, err := parser.ParseExpr("func() { " + got + " }")
			So(err, ShouldBeNil)
		}
	})
}
