# Generated go-pg/v10 Runtime

## Version Policy

The complete generator stack supports only go-pg/v10. Project loading, XML generation, model generation, search generation, repository generation, dbtest generation, VT generation, public API reporting, fixtures, and documentation use version 10.

Projects declaring go-pg 8 or go-pg 9 are rejected before generated files are written. The error identifies the unsupported version and instructs the user to migrate the project to version 10. There is no compatibility branch or generated import for either older version.

## Repository Modes

The repository generator exposes two modes:

- `legacy`, which is the default and emits the current named repository API using go-pg/v10 imports.
- `generic`, which emits the generated generic runtime and the typed repository surface for the target package.

The `bridge` mode, `GenDBImport` option, external runtime import, and bridge adapter templates are absent.

Legacy and generic output share model metadata and reusable generated support, but generic mode does not silently change legacy query construction or method signatures.

## Runtime Interface

Generic output contains one runtime support file per target package. The runtime is implemented in that package and uses native go-pg types.

The query composition seam is a function type:

```go
type Applier func(*orm.Query) (*orm.Query, error)
```

The runtime provides typed `GenRepo[T]` construction and operations for one, list, count, insert, update, and delete. Transaction rebinding accepts `*pg.Tx` through go-pg's `orm.DB` contract. Runtime helpers remain unexported unless generated code requires a name at package scope.

The runtime applies every returned query and propagates every error. It does not use reflection to discover a backend and does not expose a second ORM abstraction.

## Query Behavior

For read operations, runtime order is:

1. Repository base filters.
2. Entity search.
3. Pager.
4. Explicit operations supplied by the caller.

`Full<Entity>` and `Default<Entity>Sort` remain explicit operations in legacy output. Status filtering, joins, and default sorting are not added implicitly by generic construction unless that is already an observable legacy default.

Native go-pg CTE operations are accepted through `Applier`, including `With`, `WrapWith`, and supported insert, update, and delete CTE operations. No new CTE language or SQL builder is introduced.

## CRUD and Errors

Generated repositories preserve the existing observable behavior:

- `ErrNoRows` returns a nil object and nil error for one-row lookup.
- `ErrMultiRows` is returned to the caller.
- Update and delete report whether `RowsAffected()` is greater than zero.
- Default insert and update column exclusions are applied when no explicit column operation is supplied.
- Soft-delete updates the status column where the model has a status field; otherwise delete is physical.
- Composite primary-key methods preserve metadata order.
- `WithTransaction` returns a repository using the supplied transaction without mutating the original value repository.

## Generated Output

Generated output is formatted Go source. The runtime is generated at most once for a target package, regardless of namespace count. Legacy output remains independently usable. Generic output is a separate, compile-tested template and is not an alias of the legacy template.

Search generation preserves nil-search behavior and uses the returned query and error from every applier. `model_params.go` remains append-only according to current generator behavior.

## Verification

The capability is accepted only when all of the following have direct evidence:

- v10-only validation and imports across the complete generator stack;
- legacy golden and API behavior checks;
- generic generated compile checks against the actual in-package runtime;
- SQL and behavior checks for filters, search, paging, joins, sorting, CTEs, CRUD, transactions, deletes, errors, primary keys, and column policies;
- isolated PostgreSQL integration checks using `docs/testdb/schema.sql`;
- Go 1.26 build, test, race, vet, formatting, and lint checks;
- fresh read-only Comet Native verification for every acceptance item.
