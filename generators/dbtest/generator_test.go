package dbtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.Output = testdata.PathActualDBTest
			generator.options.MFDPath = testdata.PathExpectedMFD
			generator.options.Package = testdata.PackageDBTest
			generator.options.DBPackage = "github.com/vmkteam/mfd-generator/generators/testdata/expected/db"

			t.Log("Generate model")
			So(generator.Generate(), ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"test.go": {},
			}

			for f := range expectedFilenames {
				t.Logf("Check %s file", f)
				content, err := os.ReadFile(filepath.Join(testdata.PathActualDBTest, f))
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedDBTest, f))
				if err != nil {
					t.Fatal(err)
				}

				So(string(content), ShouldResemble, string(expectedContent))
			}
		})
	})
}
