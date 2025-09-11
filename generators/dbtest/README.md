# dbtest Generator

A Go package for generating database test helper functions from MFD definitions.
This tool creates or updates functions for inserting test data by namespaces and entities.

## Features

- **Automatic Test Function Generation**: Creates test helper functions for database operations
- **Namespace-based Organization**: Generates functions organized by namespaces
- **Selective Generation**: Supports generating functions for specific namespaces and entities
- **Force Regeneration**: Option to force regenerate existing functions

## Usage

Firstly you need to generate basis of project using [xml generator](../xml/README.md).

After that `dbtest` generator will read annotation of namespaces and entities from the xml file and generate helpers for db tests.

### Command Line Interface

```bash
# Basic usage
dbtest -o ./testdata -m project.mfd -x dbpackage

# Generate for specific namespaces
dbtest -o ./testdata -m project.mfd -x dbpackage -n portal,geo

# Generate for specific entities
dbtest -o ./testdata -m project.mfd -x dbpackage -e News,Categories

# Force regeneration
dbtest -o ./testdata -m project.mfd -x dbpackage -f

# Force regeneration only News entity
dbtest -o ./testdata -m project.mfd -x dbpackage -e News -f
```

### Required Flags

- `-o, --output`: Output directory path for generated files
- `-m, --mfd`: Path to the MFD file containing project definitions
- `-x, --db-pkg`: Package containing database files generated with model generator

### Optional Flags

- `-p, --package`: Package name for generated Go files (defaults to output directory name)
- `-n, --namespaces`: Comma-separated list of namespaces to generate
- `-e, --entities`: Comma-separated list of entities to generate
- `-f, --force`: Force regenerate existing functions

## Generated Output

The generator creates:

1. **Setup File** (`test.go`): Base test utilities and setup functions
2. **Namespace Files**: Separate Go files for each namespace containing:
   - Main test data insertion functions
   - Functions with relations support
   - Functions with fake data generation

## Example Generated Function

```go
func InsertUser(db *pg.DB, user *dbpackage.User) (*dbpackage.User, error) {
    // Generated insertion logic
    return user, db.Insert(user)
}

func InsertUserWithRelations(db *pg.DB, user *dbpackage.User) (*dbpackage.User, error) {
    // Generated insertion with related entities
}

func InsertUserWithFake(db *pg.DB) (*dbpackage.User, error) {
    // Generated insertion with fake data
}
```
