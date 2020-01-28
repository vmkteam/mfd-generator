package xml

import (
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	generator := New()

	generator.options.Def()
	generator.options.URL = `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable`
	generator.options.Output = "./test/test.mfd"

	if err := generator.Generate(); err != nil {
		t.Errorf("generate error = %v", err)
		return
	}
}
