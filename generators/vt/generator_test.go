package vt

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vmkteam/mfd-generator/generators/testdata"

	. "github.com/smartystreets/goconvey/convey"
)

// getDataCommentCount return count custom comment in file
func getDataCommentCount(path string) int {
	var count int
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "//data") {
			count++
		}
	}
	return count
}

// returnTestData function prepare test data
func returnTestData() (err error) {
	ff := make(map[string]string)
	ff[testdata.PathUpdated+"/vt/portal_actual.txt"] = testdata.PathUpdated + "/vt/portal.go"
	ff[testdata.PathUpdated+"/vt/portal_model_actual.txt"] = testdata.PathUpdated + "/vt/portal_model.go"
	ff[testdata.PathUpdated+"/vt/portal_converter_actual.txt"] = testdata.PathUpdated + "/vt/portal_converter.go"

	for srcPath, destPath := range ff {
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("cannot opent file: %v", err)
		}
		defer func(srcFile *os.File) {
			errFile := srcFile.Close()
			if errFile != nil {
				err = fmt.Errorf("cannot close file: %v", errFile)
			}
		}(srcFile)

		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("cannot rewrite file: %v", err)
		}
		defer func() {
			errFile := destFile.Close()
			if errFile != nil {
				err = fmt.Errorf("cannot close file: %v", errFile)
			}
		}()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return fmt.Errorf("cannot copy: %v", err)
		}

	}
	return nil
}

func TestGenerator_Generate(t *testing.T) {
	Convey("TestGenerator_Generate", t, func() {
		Convey("Update only news entity", func() {
			// prepare data
			_ = returnTestData()

			// get count comment before used generator
			startCountServiceComment := getDataCommentCount(testdata.PathUpdated + "/vt/portal.go")
			startCountModelComment := getDataCommentCount(testdata.PathUpdated + "/vt/portal_model.go")
			startCountConvertComment := getDataCommentCount(testdata.PathUpdated + "/vt/portal_converter.go")

			generator := New()

			generator.options.Def()
			generator.options.Output = testdata.PathUpdatedVT
			generator.options.MFDPath = testdata.PathExpectedMFD
			generator.options.Package = testdata.PackageVT
			generator.options.Namespaces = []string{"portal"}

			// added entity what need updated
			generator.options.Entities = []string{"news"}
			generator.options.ModelPackage = "github.com/vmkteam/mfd-generator/generators/testdata/expected/db"
			generator.options.EmbedLogPackage = "github.com/vmkteam/mfd-generator/generators/testdata/expected/embedlog"

			t.Log("Generate vt")
			err := generator.Generate()
			So(err, ShouldBeNil)

			// get count comment after used generator
			endCountServiceComment := getDataCommentCount(testdata.PathUpdated + "/vt/portal.go")
			endCountModelComment := getDataCommentCount(testdata.PathUpdated + "/vt/portal_model.go")
			endCountConvertComment := getDataCommentCount(testdata.PathUpdated + "/vt/portal_converter.go")

			// checked that after generate struct or function rewrite but not all
			So(startCountServiceComment, ShouldNotEqual, endCountServiceComment)
			So(startCountServiceComment-endCountServiceComment, ShouldEqual, 1)

			So(startCountModelComment, ShouldNotEqual, endCountModelComment)
			So(startCountModelComment-endCountModelComment, ShouldEqual, 3)

			So(startCountConvertComment, ShouldNotEqual, endCountConvertComment)
			So(startCountConvertComment-endCountConvertComment, ShouldEqual, 2)

			So(err, ShouldBeNil)
		})
		Convey("Check correct generate", func() {
			generator := New()

			generator.options.Def()
			generator.options.Output = testdata.PathActualVT
			generator.options.MFDPath = testdata.PathExpectedMFD
			generator.options.Package = testdata.PackageVT
			generator.options.Namespaces = []string{"portal"}
			generator.options.ModelPackage = "github.com/vmkteam/mfd-generator/generators/testdata/expected/db"
			generator.options.EmbedLogPackage = "github.com/vmkteam/mfd-generator/generators/testdata/expected/embedlog"

			t.Log("Generate vt")
			err := generator.Generate()
			So(err, ShouldBeNil)
		})

		Convey("Check generated files", func() {
			expectedFilenames := map[string]struct{}{
				"portal.go":           {},
				"portal_converter.go": {},
				"portal_model.go":     {},
			}

			for f := range expectedFilenames {
				t.Logf("Check %s file", f)
				content, err := os.ReadFile(filepath.Join(testdata.PathActualVT, f))
				if err != nil {
					t.Fatal(err)
				}
				expectedContent, err := os.ReadFile(filepath.Join(testdata.PathExpectedVT, f))
				if err != nil {
					t.Fatal(err)
				}
				So(content, ShouldResemble, expectedContent)
			}
		})
	})
}
