package xmlvt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerator_Generate(t *testing.T) {
	err := prepareFiles()
	if err != nil {
		t.Fatal(err)
	}

	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.MFDPath = testdata.PathActualMFD

			t.Log("Generate xml-vt")
			err := generator.Generate()
			So(err, ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"portal.vt.xml": {},
				"geo.vt.xml":    {},
				"vfs.vt.xml":    {},
				"card.vt.xml":   {},
				"common.vt.xml": {},
			}

			for f := range expectedFilenames {
				t.Logf("Check %s file", f)
				content, err := os.ReadFile(filepath.Join(testdata.PathActual, f))
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpected, f))
				if err != nil {
					t.Fatal(err)
				}
				So(string(content), ShouldResemble, string(expectedContent))
			}
		})
	})
}

func prepareFiles() error {
	err := os.MkdirAll(testdata.PathActual, 0775)
	if err != nil {
		return err
	}

	err = os.Link(testdata.PathExpectedMFD, testdata.PathActualMFD)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, testdata.FilenameXML), filepath.Join(testdata.PathActual, testdata.FilenameXML))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, "geo.xml"), filepath.Join(testdata.PathActual, "geo.xml"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, "geo.vt.xml"), filepath.Join(testdata.PathActual, "geo.vt.xml"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, "vfs.xml"), filepath.Join(testdata.PathActual, "vfs.xml"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, "vfs.vt.xml"), filepath.Join(testdata.PathActual, "vfs.vt.xml"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, "card.xml"), filepath.Join(testdata.PathActual, "card.xml"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, "card.vt.xml"), filepath.Join(testdata.PathActual, "card.vt.xml"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, "common.xml"), filepath.Join(testdata.PathActual, "common.xml"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Link(filepath.Join(testdata.PathExpected, "common.vt.xml"), filepath.Join(testdata.PathActual, "common.vt.xml"))
	if err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}
