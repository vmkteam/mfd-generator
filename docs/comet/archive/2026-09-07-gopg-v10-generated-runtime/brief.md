# Outcome

MFD Generator produces working Go code for PostgreSQL through go-pg/v10 only. The repository generator keeps the existing legacy repository behavior and adds a distinct generic repository mode whose runtime is generated in the target package. Generated code is executable against the test schema, not a compile-only facade.

# Scope

- Migrate the complete generator stack and generated output to go-pg/v10.
- Remove go-pg 8/9 branches, imports, flags, API values, and documentation claims.
- Keep `legacy` as the default repository mode with its observable API and behavior.
- Add `generic` repository generation backed by an in-package `GenRepo[T]` runtime.
- Use a small go-pg-specific `Applier` function type for query composition.
- Preserve filters, search, paging, joins, default sorting, CRUD, transactions, CTEs, delete behavior, error behavior, primary-key order, and column policies.
- Add generated-output compile tests and real PostgreSQL integration tests.
- Update Go, CI, fixtures, and checks to Go 1.26.

# Non-goals

- Bun, pgx, or any second database backend.
- An external `gendb` module or runtime dependency.
- A third `bridge` mode or a backend-neutral ORM facade.
- Compatibility with go-pg 8 or go-pg 9.
- A new iteration API without evidence of a current consumer.
- Push, PR, MR creation, merge, release, or other remote delivery actions.

# Acceptance examples

- A1: A project with `GoPGVer=10` generates model, search, repository, and dbtest code that imports only `github.com/go-pg/pg/v10`.
- A2: A project with `GoPGVer=8` or `GoPGVer=9` is rejected with an actionable migration error before generation.
- A3: Legacy output preserves generated names, method signatures, filter-search-pager-operation ordering, `Full` and default-sort operations, soft and physical delete, column policies, primary-key order, and `pg.ErrNoRows`/`pg.ErrMultiRows` behavior.
- A4: Generic output is distinct from legacy output, compiles without an external runtime module, and executes list, one, count, insert, update, and delete operations against PostgreSQL.
- A5: Query appliers propagate returned queries and errors; CTE and transaction operations execute with native go-pg semantics.
- A6: Generated output tests do not rely on the no-op runtime, mutate tracked fixtures, or silently skip required database checks.
- A7: Go 1.26 build, test, race, vet, formatting, and lint checks pass using pinned dependencies and CI-equivalent PostgreSQL setup.
- A8: Comet Native verification records a result for every acceptance item, with no failed or blocked item before Archive readiness.

# Constraints and invariants

- The implementation is built in the dedicated Native worktree `mfd-generator-native` from `master`.
- The existing dirty worktree is read-only input and is not overwritten or treated as a verified implementation.
- `legacy` remains the default repository mode.
- The generic runtime is generated once per target package and is not duplicated per namespace.
- `Applier` is a go-pg-specific function type: `func(*orm.Query) (*orm.Query, error)`.
- Runtime query order is base filters, search, pager, then explicit operations.
- Integration tests use an isolated disposable PostgreSQL service initialized from `docs/testdb/schema.sql`.
- Required integration failures are test failures, never skips.
- No code is declared complete from local Go 1.27 results alone; Go 1.26-equivalent checks are required.

# Decisions

- The Native project is bound to the dedicated worktree rather than `/Users/rage/arch/us-411`.
- The runtime is generated inside the target package so one change can prove implementation and generated-code behavior without an external release dependency.
- The public repository modes are `legacy` and `generic`; `bridge` is removed because it adds shallow compatibility boilerplate.
- The entire generator stack moves to go-pg/v10, including legacy generation.
- Generics are used for `GenRepo[T]` where they remove repeated repository implementation; ORM-neutral interfaces and reflection are not introduced.
- Native Build, fresh read-only Verify, and Runtime-owned state transitions are the completion workflow.

# Open questions

- [blocking] CONFIRM: Confirm the outcome, scope, non-goals, decisions, acceptance items A1-A8, and the constraint that no PR/MR or remote delivery action is performed.

# Verification expectations

- Run generator unit and golden tests in temporary directories.
- Compile both legacy and generic generated fixtures against real go-pg/v10 code.
- Run PostgreSQL integration tests with the test schema and explicit connection environment.
- Run `go test ./...`, `go test -race ./...`, `go test -tags=integration ./...`, `go vet ./...`, `go build ./...`, and the pinned golangci-lint command.
- Run the Go 1.26-equivalent checks in a pinned Go 1.26 container when the host toolchain is not Go 1.26.
- Run `go generate ./api` and fail if generated API files differ from the checked-in result.
- Submit a Builder handoff, let Runtime execute checks, and require a fresh read-only Verifier result for A1-A8.
