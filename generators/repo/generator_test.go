package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"
	"github.com/vmkteam/mfd-generator/mfd"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerator_Generate(t *testing.T) {
	actualDir := t.TempDir()
	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.Output = actualDir
			generator.options.MFDPath = testdata.PathExpectedMFD
			generator.options.Package = testdata.PackageDB
			generator.options.Namespaces = []string{"portal", "geo", "card", "common"}

			t.Log("Generate repo")
			So(generator.Generate(), ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"portal.go": {},
				"geo.go":    {},
				"card.go":   {},
				"common.go": {},
			}

			for f := range expectedFilenames {
				t.Logf("Check %s file", f)
				content, err := os.ReadFile(filepath.Join(actualDir, f))
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedDB, f))
				if err != nil {
					t.Fatal(err)
				}
				So(string(content), ShouldResemble, string(expectedContent))
			}
		})
	})
}

func TestPackEntityPreservesCompositePrimaryKeyOrder(t *testing.T) {
	entity := mfd.Entity{
		Name:  "Membership",
		Table: "memberships",
		Attributes: mfd.Attributes{
			{Name: "UserID", DBName: "userId", GoType: "int", PrimaryKey: true},
			{Name: "GroupID", DBName: "groupId", GoType: "int", PrimaryKey: true},
		},
	}

	data := PackEntity(entity, Options{})
	if len(data.PKs) != 2 {
		t.Fatalf("got %d primary keys, want 2", len(data.PKs))
	}
	if data.PKs[0].Field != "UserID" || data.PKs[1].Field != "GroupID" {
		t.Fatalf("primary key order = %q, %q", data.PKs[0].Field, data.PKs[1].Field)
	}
}
