package api

import (
	"context"
	"sync"

	"github.com/vmkteam/zenrpc"
)

type MockService struct {
	zenrpc.Service

	mx    sync.Mutex
	cache map[string]*Project
}

func NewMockService() *MockService {
	return &MockService{
		cache: map[string]*Project{
			"vtsrv": {
				Name:        "vtsrv",
				Languages:   []string{"ru"},
				GoPGVer:     9,
				CustomTypes: nil,
				Namespaces: []Namespace{
					{
						Name: "catalogue",
						Entities: []Entity{
							{
								Name:      "Integration",
								Namespace: "",
								Table:     "integrations",
								Attributes: []Attribute{
									{Name: "ID", DBName: "integrationId", DBType: "int4", GoType: "int", PrimaryKey: true, ForeignKey: "", Nullable: false, Addable: true, Updatable: false, Min: nil, Max: nil, Default: ""},
									{Name: "Network", DBName: "network", DBType: "varchar", GoType: "string", PrimaryKey: false, ForeignKey: "", Nullable: false, Addable: true, Updatable: true, Min: nil, Max: nil, Default: ""},
									{Name: "OrderType", DBName: "orderType", DBType: "varchar", GoType: "string", PrimaryKey: false, ForeignKey: "", Nullable: false, Addable: true, Updatable: true, Min: nil, Max: nil, Default: ""},
									{Name: "StatusID", DBName: "statusId", DBType: "int4", GoType: "int", PrimaryKey: false, ForeignKey: "", Nullable: false, Addable: true, Updatable: true, Min: nil, Max: nil, Default: ""},
									{Name: "UseIntegrationOrderID", DBName: "useIntegrationOrderId", DBType: "bool", GoType: "bool", PrimaryKey: false, ForeignKey: "", Nullable: false, Addable: true, Updatable: true, Min: nil, Max: nil, Default: ""},
								},
								Searches: []Search{
									{Name: "IDs", AttrName: "ID", SearchType: "SEARCHTYPE_ARRAY"},
									{Name: "NetworkILike", AttrName: "Network", SearchType: "SEARCHTYPE_ILIKE"},
									{Name: "OrderTypeILike", AttrName: "OrderType", SearchType: "SEARCHTYPE_ILIKE"},
								},
							},
							{
								Name:      "Pharmacy",
								Namespace: "",
								Table:     "pharmacies",
								Attributes: []Attribute{
									{Name: "ID", DBName: "pharmacyId", DBType: "int4", GoType: "int", PrimaryKey: true, ForeignKey: "", Nullable: false, Addable: true, Updatable: false, Min: nil, Max: nil, Default: ""},
									{Name: "PharmacyNetworkID", DBName: "pharmacyNetworkId", DBType: "int4", GoType: "int", PrimaryKey: false, ForeignKey: "PharmacyNetwork", Nullable: false, Addable: true, Updatable: true, Min: nil, Max: nil},
								},
								Searches: nil,
							},
						},
					},
				},
			},
		},
	}
}

func (s *MockService) Ping(ctx context.Context) (string, error) {
	return "pong", nil
}

//zenrpc:return Project
func (s *MockService) LoadProject(ctx context.Context, filepath string) (*Project, error) {
	// mfd.LoadProject(filepath)
	name := ""
	p, _ := s.cache[name]
	return p, nil
}

//zenrpc:return Project
func (s *MockService) Project(ctx context.Context, name string) (*Project, error) {
	p, _ := s.cache[name]
	return p, nil
}

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
