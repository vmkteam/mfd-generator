package xmllang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

// todo: add generator.options.Output for generate to actual directory. Now it generate in expected dir where located .mfd
// todo: not generate if *.vt.xml not exists, but err == nil
func TestGenerator_Generate(t *testing.T) {
	actualDir := t.TempDir()
	err := prepareFiles(actualDir)
	if err != nil {
		t.Fatal(err)
	}
	mfdPathInActual := filepath.Join(actualDir, filepath.Base(testdata.PathExpectedMFD))

	Convey("TestGenerator_Generate", t, func() {
		Convey("Generate with Entity flag", func() {
			generator := New()
			generator.options.MFDPath = mfdPathInActual
			generator.options.Entities = []string{"category"}

			t.Log("Generate only entity news xml-vt")
			err = generator.Generate()
			So(err, ShouldBeNil)

			t.Logf("Check %s file", "en-one-entity.xml")
			content, err := os.ReadFile(filepath.Join(actualDir, "en.xml"))
			So(err, ShouldBeNil)
			expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpected, "en-one-entity.xml"))
			So(err, ShouldBeNil)
			So(content, ShouldResemble, expectedContent)
		})

		Convey("Check correct generate", func() {
			generator := New()
			generator.options.MFDPath = mfdPathInActual

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
				content, err := os.ReadFile(filepath.Join(actualDir, f))
				So(err, ShouldBeNil)
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpected, f))
				So(err, ShouldBeNil)
				So(content, ShouldResemble, expectedContent)
			}
		})
	})
}

func prepareFiles(actualPath string) error {
	err := os.Link(testdata.PathExpectedMFD, filepath.Join(actualPath, filepath.Base(testdata.PathExpectedMFD)))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, testdata.FilenameXML), filepath.Join(actualPath, testdata.FilenameXML))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, testdata.FilenameVTXML), filepath.Join(actualPath, testdata.FilenameVTXML))
	if err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}
