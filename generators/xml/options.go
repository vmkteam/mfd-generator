package xml

import (
	"github.com/dizzyfool/genna/generators/base"
)

type Options struct {
	base.Options

	// quiet mode
	Quiet string

	// Namespaces stores table=namespace preset
	Packages map[string]string
}
