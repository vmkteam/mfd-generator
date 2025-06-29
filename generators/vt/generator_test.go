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
	folderPath := testdata.PathActual + "/vt-updated/"
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.Mkdir(folderPath, 0755)
		if err != nil {
			return err
		}
	}

	ff := make(map[string]string)
	ff[testdata.PathExpected+"/vt-updated/portal_actual.txt"] = testdata.PathExpected + "/vt-updated/portal.go"
	ff[testdata.PathExpected+"/vt-updated/portal_model_actual.txt"] = testdata.PathExpected + "/vt-updated/portal_model.go"
	ff[testdata.PathExpected+"/vt-updated/portal_converter_actual.txt"] = testdata.PathExpected + "/vt-updated/portal_converter.go"

	for srcPath, destPath := range ff {
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("cannot opent file: %w", err)
		}
		defer func(srcFile *os.File) {
			errFile := srcFile.Close()
			if errFile != nil {
				err = fmt.Errorf("cannot close file: %w", errFile)
			}
		}(srcFile)

		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("cannot rewrite file: %w", err)
		}
		defer func() {
			errFile := destFile.Close()
			if errFile != nil {
				err = fmt.Errorf("cannot close file: %w", errFile)
			}
		}()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return fmt.Errorf("cannot copy: %w", err)
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
			startCountServiceComment := getDataCommentCount(testdata.PathExpected + "/vt-updated/portal.go")
			startCountModelComment := getDataCommentCount(testdata.PathExpected + "/vt-updated/portal_model.go")
			startCountConvertComment := getDataCommentCount(testdata.PathExpected + "/vt-updated/portal_converter.go")

			// generate
			generator := New()

			generator.options.Def()
			generator.options.Output = testdata.PathUpdatedVT
			generator.options.MFDPath = testdata.PathExpectedMFD
			generator.options.Package = testdata.PackageVTUpdated
			generator.options.Namespaces = []string{"portal"}

			// added entity what need updated
			generator.options.Entities = []string{"news"}
			generator.options.ModelPackage = "github.com/vmkteam/mfd-generator/generators/testdata/expected/db"
			generator.options.EmbedLogPackage = "github.com/vmkteam/mfd-generator/generators/testdata/expected/embedlog"

			t.Log("Generate vt")
			err := generator.Generate()
			So(err, ShouldBeNil)

			// get count comment after used generator
			endCountServiceComment := getDataCommentCount(testdata.PathUpdatedVT + "/portal.go")
			endCountModelComment := getDataCommentCount(testdata.PathUpdatedVT + "/portal_model.go")
			endCountConvertComment := getDataCommentCount(testdata.PathUpdatedVT + "/portal_converter.go")

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
			So(generator.Generate(), ShouldBeNil)
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
				So(string(content), ShouldResemble, string(expectedContent))
			}
		})
	})
}
