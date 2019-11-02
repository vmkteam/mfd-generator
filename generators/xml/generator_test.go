package xml

import (
	"go.uber.org/zap"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.Encoding = "console"
	logger, _ := config.Build()

	generator := New(logger)

	generator.options.Def()
	generator.options.URL = `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable`
	generator.options.Output = "./test/test.mfd"

	if err := generator.Generate(); err != nil {
		t.Errorf("generate error = %v", err)
		return
	}
}
