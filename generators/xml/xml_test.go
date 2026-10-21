package xml

import (
	"testing"

	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/dizzyfool/genna/model"
	"github.com/dizzyfool/genna/util"
)

func newIntColumn(pgName string, pk, fk, nullable bool) model.Column {
	return model.NewColumn(pgName, model.TypePGInt4, "", pk, nullable, false, false, 0, pk, fk, 0, nil, mfd.GoPG10, nil)
}

func newTitleColumn() model.Column {
	return model.NewColumn("title", model.TypePGVarchar, "", false, false, false, false, 0, false, false, 255, nil, mfd.GoPG10, nil)
}

func newRelation(column, table string) model.Relation {
	return model.NewRelation([]string{column}, util.PublicSchema, table)
}

// buildProject repeats Generator.Generate steps that run after reading tables from db.
func buildProject(namespace string, entities ...model.Entity) *mfd.Project {
	project := mfd.NewProject("test", mfd.GoPG10)
	for _, entity := range entities {
		project.AddEntity(namespace, PackEntity(namespace, entity, nil, nil))
	}

	project.SuggestArrayLinks()
	project.UpdateLinks()

	return project
}

// statuses table is never added to project (see PromptNS), so any fk to it must not break generation.
func TestPackEntity_StatusesFK(t *testing.T) {
	tests := []struct {
		name      string
		columns   []model.Column
		relations []model.Relation
	}{
		{
			name: "statusId only",
			columns: []model.Column{
				newIntColumn("productId", true, false, false),
				newTitleColumn(),
				newIntColumn("statusId", false, true, false),
			},
			relations: []model.Relation{
				newRelation("statusId", "statuses"),
			},
		},
		{
			name: "statusId and reviewStatusId",
			columns: []model.Column{
				newIntColumn("productId", true, false, false),
				newTitleColumn(),
				newIntColumn("statusId", false, true, false),
				newIntColumn("reviewStatusId", false, true, false),
			},
			relations: []model.Relation{
				newRelation("statusId", "statuses"),
				newRelation("reviewStatusId", "statuses"),
			},
		},
		{
			name: "reviewStatusId only",
			columns: []model.Column{
				newIntColumn("productId", true, false, false),
				newTitleColumn(),
				newIntColumn("reviewStatusId", false, true, false),
			},
			relations: []model.Relation{
				newRelation("reviewStatusId", "statuses"),
			},
		},
		{
			name: "nullable reviewStatusId and oldStatusId",
			columns: []model.Column{
				newIntColumn("productId", true, false, false),
				newIntColumn("statusId", false, true, false),
				newIntColumn("reviewStatusId", false, true, true),
				newIntColumn("oldStatusId", false, true, true),
			},
			relations: []model.Relation{
				newRelation("statusId", "statuses"),
				newRelation("reviewStatusId", "statuses"),
				newRelation("oldStatusId", "statuses"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := buildProject("product", model.NewEntity(util.PublicSchema, "products", tt.columns, tt.relations))

			if err := project.IsConsistent(); err != nil {
				t.Errorf("IsConsistent() error = %v", err)
			}

			entity := project.Entity("Product")
			for _, column := range tt.columns {
				attr := entity.AttributeByDBName(column.PGName, column.PGType)
				if attr == nil {
					t.Errorf("attribute for column %s not found", column.PGName)
					continue
				}

				if attr.ForeignKey != "" {
					t.Errorf("attribute %s: ForeignKey = %q, want empty", attr.Name, attr.ForeignKey)
				}
				// nullability comes from db for all columns except statusId
				if wantNullable := column.IsPK || column.Nullable; !mfd.IsStatus(column.PGName) && attr.Nullable() != wantNullable {
					t.Errorf("attribute %s: Nullable = %v, want %v", attr.Name, attr.Nullable(), wantNullable)
				}
			}
		})
	}
}

// sellersrv case: reviewStatusId is already in xml without fk, regeneration keeps it as is.
func TestPackEntity_StatusesFKExisting(t *testing.T) {
	existing := &mfd.Entity{Name: "Product", Namespace: "product", Table: "products", Attributes: mfd.Attributes{
		{Name: "ID", DBName: "productId", DBType: model.TypePGInt4, GoType: model.TypeInt, PrimaryKey: true, Null: mfd.NullableYes},
		{Name: "ReviewStatusID", DBName: "reviewStatusId", DBType: model.TypePGInt4, GoType: model.TypeInt, Null: mfd.NullableNo},
	}}

	entity := model.NewEntity(util.PublicSchema, "products", []model.Column{
		newIntColumn("productId", true, false, false),
		newIntColumn("reviewStatusId", false, true, false),
	}, []model.Relation{
		newRelation("reviewStatusId", "statuses"),
	})

	packed := PackEntity("product", entity, existing, nil)
	if len(packed.Attributes) != 2 {
		t.Fatalf("attributes = %d, want 2", len(packed.Attributes))
	}

	if attr := packed.AttributeByName("ReviewStatusID"); attr.ForeignKey != "" || attr.Nullable() {
		t.Errorf("ReviewStatusID: ForeignKey = %q, Nullable = %v, want empty and false", attr.ForeignKey, attr.Nullable())
	}
}

func TestPackEntity_MultipleFKToSameTable(t *testing.T) {
	categories := model.NewEntity(util.PublicSchema, "categories", []model.Column{
		newIntColumn("categoryId", true, false, false),
		newTitleColumn(),
	}, nil)

	offers := model.NewEntity(util.PublicSchema, "offers", []model.Column{
		newIntColumn("offerId", true, false, false),
		newIntColumn("categoryId", false, true, false),
		newIntColumn("extraCategoryId", false, true, true),
	}, []model.Relation{
		newRelation("categoryId", "categories"),
		newRelation("extraCategoryId", "categories"),
	})

	nodes := model.NewEntity(util.PublicSchema, "nodes", []model.Column{
		newIntColumn("nodeId", true, false, false),
		newIntColumn("parentNodeId", false, true, true),
		newIntColumn("rootNodeId", false, true, true),
	}, []model.Relation{
		newRelation("parentNodeId", "nodes"),
		newRelation("rootNodeId", "nodes"),
	})

	tests := []struct {
		name     string
		entities []model.Entity
		entity   string
		target   string
		attrs    []string
	}{
		{
			name:     "two fks to other table",
			entities: []model.Entity{categories, offers},
			entity:   "Offer",
			target:   "Category",
			attrs:    []string{"CategoryID", "ExtraCategoryID"},
		},
		{
			name:     "two fks to self",
			entities: []model.Entity{nodes},
			entity:   "Node",
			target:   "Node",
			attrs:    []string{"ParentNodeID", "RootNodeID"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := buildProject("catalog", tt.entities...)

			if err := project.IsConsistent(); err != nil {
				t.Fatalf("IsConsistent() error = %v", err)
			}

			target := project.Entity(tt.target)
			for _, name := range tt.attrs {
				attr := project.Entity(tt.entity).AttributeByName(name)
				if attr == nil {
					t.Fatalf("attribute %s not found", name)
				}
				if attr.ForeignKey != tt.target || attr.ForeignEntity != target {
					t.Errorf("attribute %s: ForeignKey = %q, ForeignEntity = %v, want %s", name, attr.ForeignKey, attr.ForeignEntity, tt.target)
				}
			}
		})
	}
}
