package xmllang

import (
	"os"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

// todo: add generator.options.Output for generate to actual directory. Now it generate in expected dir where located .mfd
// todo: not generate if *.vt.xml not exists, but err == nil
func TestGenerator_Generate(t *testing.T) {
	err := prepareFiles()
	if err != nil {
		t.Fatal(err)
	}

	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.MFDPath = testdata.PathActualMfd

			t.Log("Generate xml-vt")
			err = generator.Generate()
			So(err, ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"en.xml": {},
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

func prepareFiles() error {
	err := os.MkdirAll(testdata.PathActual, 0775)
	if err != nil {
		return err
	}

	err = os.Link(testdata.PathExpectedMfd, testdata.PathActualMfd)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(testdata.PathExpected+testdata.FilenameXml, testdata.PathActual+testdata.FilenameXml)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(testdata.PathExpected+testdata.FilenameVtXml, testdata.PathActual+testdata.FilenameVtXml)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}
