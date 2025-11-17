# dbtest Generator

A generator designed to create helper functions for database testing from MFD definitions.
This tool creates or updates functions for inserting test data by namespaces and entities.

## Features

- **Automatic Test Function Generation**: Creates helper functions for inserting data into tables.
- **Automatic Optional Function Generation**: Creates optional functions to fill required fields
  with random data, for pre-creating relations necessary for inserting an entity.
- **Selective Generation**: Supports generating functions for specific namespaces and entities.
- **Force Regeneration**: Option to forcibly regenerate existing functions.

## Usage

First, you need to generate the project foundation using the [xml generator](../xml/README.md),
[model generator](../model/README.md), and [repository generator](../repo/README.md).

After that, the `dbtest` generator will read annotations of namespaces and entities from the xml file
and generate helpers for db tests.

### Command Line Interface

To generate, go to your project directory. A basic command example with required flags:
```bash
# Basic usage
dbtest -x 'newsportal/pkg/db' -o './pkg/db/test' -m './docs/model/newsportal.mfd'
```

### Required Flags

- `-x, --db-pkg`: Package containing database files generated with the model generator.  
  Must be specified exactly as it appears in the `import` block.
- `-o, --output`: Path where the generated files will be placed.
- `-m, --mfd`: Path to the `.mfd` file of the project.

### Optional Flags

- `-n, --namespaces` - generate only for the specified namespaces (similar to the [repository generator](../repo/README.md)).  
  Example: `-n portal,geo`.
- `-e, --entities` - generate only for the specified entities (similar to the [vt generator](../vt/README.md)).  
  Example: `-e news,categories`.
- `-f, --force` - force regeneration. If this flag is set, existing content will be replaced during regeneration.  
  If `-f` is passed without `-n` and `-e`, all existing files will be replaced with new ones.  
  If `-f` is passed along with at least `-n` or `-e`, all passed entities or entities within the specified namespaces
  will be found, and only their content will be replaced.

### Examples:

```bash
# Generate for specific namespaces
dbtest -x 'newsportal/pkg/db' -o './pkg/db/test' -m './docs/model/newsportal.mfd' -n portal,geo

# Generate for specific entities
dbtest -x 'newsportal/pkg/db' -o './pkg/db/test' -m './docs/model/newsportal.mfd' -e news,categories

# Force regeneration
dbtest -x 'newsportal/pkg/db' -o './pkg/db/test' -m './docs/model/newsportal.mfd' -f

# Force regeneration only for the News entity
dbtest -x 'newsportal/pkg/db' -o './pkg/db/test' -m './docs/model/newsportal.mfd' -e news -f
```

## Generated Output

The generator creates:

1. **Setup File** (`test.go`): Base testing utilities and setup functions.
2. **Namespace Files**: Separate Go files for each namespace containing:
   - Main functions for inserting test data,
   - The OpFunc type for additional functions,
   - Additional functions with support for inserting all nested relations,
   - Additional functions with random data generation for required fields.

## Example Generated Function

```go
// NewsOpFunc Type of additional functions
type NewsOpFunc func(t *testing.T, dbo orm.DB, in *db.News) Cleaner

// News Main helper for inserting a news item
func News(t *testing.T, dbo orm.DB, in *db.News, ops ...NewsOpFunc) (*db.News, Cleaner) {
    // Generated insertion logic
}

// WithNewsRelations Additional function to pass into News.
// Helps insert all necessary relations for News before directly inserting the news itself.
// Useful if we don’t care which relations are created, and the caller doesn’t want to think about it.
// If the caller wants to strictly control relation behavior, this can be omitted when calling News.
func WithNewsRelations(t *testing.T, dbo orm.DB, in *db.News) Cleaner {
    // Generated insertion with related entities
}

// WithFakeNews Additional function to pass into News.
// Generates random data for required fields before directly inserting the news item.
// Useful if we don’t care what the data is, but we need something and don’t want to do it manually.
// If the caller wants to strictly control data behavior, this can be omitted when calling News.
func WithFakeNews(t *testing.T, dbo orm.DB, in *db.News) Cleaner {
    // Generated insertion with fake data
}
```

## Example Usage of Generated Functions

```go
   // Simple tag insertion
   func Test_test(t *testing.T) {
       dbo := test.Setup(t)

       res, clean := test.Tag(t, dbo, nil)
       defer clean()

       fmt.Println(res)
   }

   // Simple news insertion with a specified tag and category
   func Test_test(t *testing.T) {
       dbo := test.Setup(t)

       tag, clean := test.Tag(t, dbo, nil)
       defer clean()
       c, clean := test.Category(t, dbo, nil)
       defer clean()
       news, clean := test.News(t, dbo, &db.News{CategoryID: c.ID, TagIDs: []int{tag.ID}})
       defer clean()

       fmt.Println(news)
   }

   // Insert news with random relations and generated required fields
   func Test_test(t *testing.T) {
       dbo := test.Setup(t)

       news, clean := test.News(t, dbo, nil, WithNewsRelations, WithFakeNews)
       defer clean()

       fmt.Println(news)
   }

    // Insert 100 news items in one category
   func Test_test(t *testing.T) {
       dbo := test.Setup(t)

       news, clean := test.News(t, dbo, nil, WithNewsRelations, WithFakeNews)
       defer clean()

       for i := 0; i < 99; i++ {
           _, clean := test.News(t, dbo, &db.News{
              CategoryID: news.CategoryID,
           }, WithNewsRelations, WithFakeNews)
           defer clean()
       }
   }
```

## Notes

If the required entity already exists in the DB, and you need to get it by PKs, just call:

```go
   // Simple tag insertion
   func Test_test(t *testing.T) {
       dbo := test.Setup(t)

       news, clean := test.News(t, dbo, &db.News{ID: 10})
       defer clean()

       fmt.Println(news)
   }
```

When a PK is provided, there will always be an attempt to find the corresponding entity by it.

**Important!** If all PKs in the DB are `not null` and do not have a `default` property
(i.e., no default value is provided for empty input), in this case, if the entity is not found,
an attempt will be made to **insert** a new entity with the PK values that were provided.

This behavior is needed when we want to insert rows into tables that have a simple or composite PK
marked as `not null` in the DB, and no default value is defined for it in the DB.

For example, there is a price table with a composite PK without default values:

```sql
create table prices (
    "partnerId" text not null,
    "productId" text not null,
    title text,
    constraint pk_prices
       primary key ("partnerId", "productId")
)
```

To insert data into such a table, you must specify `partnerId` and `productId`.

Therefore, for such cases, there is support for inserting new rows using the provided PKs,
if they were not set.

In other cases, for example, for news:

```sql
CREATE TABLE "news" (
   "newsId" SERIAL NOT NULL,
   "title" varchar(255) NOT NULL,
   "content" text,
   "categoryId" int4 NOT NULL,
   "tagIds" int4[],
   "createdAt" timestamp with time zone NOT NULL DEFAULT NOW(),
   "publishedAt" timestamp with time zone,
   "statusId" int4 NOT NULL,
   PRIMARY KEY("newsId")
);
```

The PK is `newsId`, which is also `not null`, but it has a default value.

For such cases, if the entity is not found by the provided PK, the test will fail with an error.
