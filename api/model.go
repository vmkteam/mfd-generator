package api

type Project struct {
	Name        string       `json:"name"`
	Languages   []string     `json:"languages"`
	GoPGVer     int          `json:"goPgVer"`
	CustomTypes []CustomType `json:"customTypes"`
	Namespaces  []Namespace  `json:"namespaces"`
	//VTNamespaces []VTNamespace `json:"vtNamespaces"`
}

type CustomType struct {
	DBType   string `json:"dbType"`
	GoImport string `json:"goImport"`
	GoType   string `json:"goType"`
}

type Namespace struct {
	Name string `json:"name"`

	Entities []Entity `json:"entities"`
}

type Entity struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"` // wtf is it here??
	Table     string `json:"table"`

	Attributes []Attribute `json:"attributes"`
	Searches   []Search    `json:"searches"`
}

// Attribute is xml element
type Attribute struct {
	// names
	Name   string `json:"name"`
	DBName string `json:"dbName"`

	// types
	IsArray bool   `json:"isArray"`
	DBType  string `json:"dbType"`
	GoType  string `json:"goType"`

	// Keys
	PrimaryKey bool   `json:"primaryKey"`
	ForeignKey string `json:"foreignKey"`

	// data params
	Nullable  bool   `json:"nullable"`
	Addable   bool   `json:"addable"`
	Updatable bool   `json:"updatable"`
	Min       *int   `json:"min,omitempty"`
	Max       *int   `json:"max,omitempty"`
	Default   string `json:"defaultValue,omitempty"`
}

type Search struct {
	Name       string `json:"name"`
	AttrName   string `json:"attrName"`
	SearchType string `json:"searchType"`
}

type VTNamespace struct {
	Name string `json:"name"`

	Entities []VTEntity `json:"entities"`
}

type VTEntity struct {
	Name string `json:"name"`

	TerminalPath string `json:"terminalPath"`
}
