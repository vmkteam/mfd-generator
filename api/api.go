package api

import (
	"context"
	"sync"

	"github.com/vmkteam/zenrpc"

	"github.com/vmkteam/mfd-generator/mfd"
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
func (s *MockService) Project(ctx context.Context, filepath string) (*Project, error) {
	project, err := mfd.LoadProject(filepath, false, 0)
	if err != nil {
		return nil, err
	}

	p := newProject(project)

	return p, nil
}
