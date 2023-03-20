package xmlvt

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/vmkteam/mfd-generator/generators/testdata"
)

// todo: fail after rerun because invalid generate "xmlns:xsi="
// todo: add generator.options.Output for generate to actual directory. Now it generate in expected dir where located .mfd
func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.MFDPath = testdata.PathExpectedMfd

			t.Log("Generate xml-vt")
			err := generator.Generate()
			So(err, ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"portal.vt.xml": {},
			}

			for f := range expectedFilenames {
				t.Logf("Check %s file", f)
				content, err := os.ReadFile(testdata.PathActual + f)
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(testdata.PathExpected + f)
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, expectedContent)
			}
		})
	})
}
