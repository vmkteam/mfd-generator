package xmlvt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerator_Generate(t *testing.T) {
	actualDir := t.TempDir()
	err := prepareFiles(actualDir)
	if err != nil {
		t.Fatal(err)
	}

	Convey("TestGenerator_Generate", t, func() {
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.MFDPath = filepath.Join(actualDir, testdata.FilenameMFD)

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
				content, err := os.ReadFile(filepath.Join(actualDir, f))
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

func prepareFiles(actualPath string) error {
	err := os.MkdirAll(actualPath, 0775)
	if err != nil {
		return err
	}

	err = copyFile(testdata.PathExpectedMFD, filepath.Join(actualPath, testdata.FilenameMFD))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(testdata.PathExpected, testdata.FilenameXML), filepath.Join(actualPath, testdata.FilenameXML))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(testdata.PathExpected, "geo.xml"), filepath.Join(actualPath, "geo.xml"))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(testdata.PathExpected, "geo.vt.xml"), filepath.Join(actualPath, "geo.vt.xml"))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(testdata.PathExpected, "vfs.xml"), filepath.Join(actualPath, "vfs.xml"))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(testdata.PathExpected, "vfs.vt.xml"), filepath.Join(actualPath, "vfs.vt.xml"))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(testdata.PathExpected, "card.xml"), filepath.Join(actualPath, "card.xml"))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(testdata.PathExpected, "card.vt.xml"), filepath.Join(actualPath, "card.vt.xml"))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(testdata.PathExpected, "common.xml"), filepath.Join(actualPath, "common.xml"))
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(testdata.PathExpected, "common.vt.xml"), filepath.Join(actualPath, "common.vt.xml"))
	if err != nil {
		return err
	}

	return nil
}

func copyFile(source, destination string) error {
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, content, 0o644)
}
