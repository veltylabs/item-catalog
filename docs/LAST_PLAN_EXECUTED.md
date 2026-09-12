---
PLAN: "fix: move schema creation out of New() into a separate Migrate(), matching webtyp.com/auth and webtyp.com/rbac"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — `Migrate()` replaces the `CreateTable` calls inside `New()`

## Part of a multi-repo wave

This is module 1 of `DDL_MIGRATE_ISOLATION_MASTER_PLAN.md` (orchestrator:
`webtyp.com/app-releases`, `docs/DDL_MIGRATE_ISOLATION_MASTER_PLAN.md`). No
dependencies — dispatch any time. `veltylabs/mjosefa-cms`'s `cmd/migrate`
stage depends on the tag this plan produces.

## Why

`New()` (`mcp.go:72-88`) calls `ddl.New(db.RawConn(), ddlCompiler).
CreateTable(...)` for `Specialty`, `CatalogItem`, and `Agreement`
unconditionally, every time a server process constructs this module against
a real SQL backend. Two siblings in this same ecosystem,
`webtyp.com/auth/authority` and `webtyp.com/rbac`, already solve this
differently: they expose a `Migrate(conn ddl.Execer, ddlCompiler
ddl.Compiler) error` function, deliberately **not** called by `New`,
documented as deploy-time work run once from a migration binary. This plan
brings `item_catalog` in line with that pattern — it is the reference module
other `veltylabs/modules/*` repos will copy next.

This is not a safety fix (`CreateTable` compiles to `CREATE TABLE IF NOT
EXISTS` — additive, never alters or drops, verified in
`webtyp.com/postgres`'s DDL translator). It is a **control and build-graph**
fix: today there is no single place to run schema reconciliation without
booting a full server process, and — separately — `mjosefa-cms/modules/
item_catalog/view.go` (compiled into the WASM client) imports this repo's
root package for `itemcatalog.NewView(caller)`. If `Migrate`/`ddl` stayed in
that same root package, `webtyp.com/ddl` would still link into the WASM
binary through the view import, no matter what build tag the composition
root puts on its own `backend.go`. **`Migrate` therefore goes in its own
subpackage, `migrate/`, not a new file in the root package** — nothing on
the WASM build path ever imports `.../item_catalog/migrate`, so
`webtyp.com/ddl` never enters that build graph at all.

## What to change

### 1. `mcp.go` — remove the DDL block from `New()`

Before:

```go
func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("item_catalog: Deps.IDs is required")
	}
	if ddlCompiler, ok := db.RawConn().(ddl.Compiler); ok {
		if err := ddl.New(db.RawConn(), ddlCompiler).CreateTable(&Specialty{}); err != nil {
			return nil, err
		}
		if err := ddl.New(db.RawConn(), ddlCompiler).CreateTable(&CatalogItem{}); err != nil {
			return nil, err
		}
		if err := ddl.New(db.RawConn(), ddlCompiler).CreateTable(&Agreement{}); err != nil {
			return nil, err
		}
	}
	return &Module{db: db, ids: deps.IDs, pub: deps.Publisher}, nil
}
```

After:

```go
func New(db *orm.DB, deps Deps) (*Module, error) {
	if deps.IDs == nil {
		return nil, fmt.Err("item_catalog: Deps.IDs is required")
	}
	return &Module{db: db, ids: deps.IDs, pub: deps.Publisher}, nil
}
```

Remove the now-unused `"webtyp.com/ddl"` import from `mcp.go` — grep the
rest of the file for `ddl.` first; if nothing else in `mcp.go` uses it
(expected: nothing does), delete the import line.

### 2. New package `migrate/` (a subdirectory, not a file in the root package)

Create `migrate/migrate.go`:

```go
package migrate

import (
	"webtyp.com/ddl"

	itemcatalog "github.com/veltylabs/item_catalog"
)

// Migrate reconciles the database schema item_catalog owns: Specialty,
// CatalogItem and Agreement, in FK dependency order (Agreement references
// CatalogItem, CatalogItem references Specialty).
//
// It is deliberately NOT called by New, and deliberately lives in its own
// package rather than a new file in the root package: nothing on a
// consuming app's WASM build path (its view.go, which imports the root
// itemcatalog package for itemcatalog.NewView) ever imports
// "github.com/veltylabs/item_catalog/migrate" — so webtyp.com/ddl never
// enters that build graph, regardless of build tags on the consumer's side.
//
// conn is a ddl.Execer, not an *orm.DB, so a deploy-time transport that can
// only execute DDL satisfies it. An *orm.DB's RawConn() also satisfies it,
// for local/test callers:
//
//	conn, _ := postgres.Open(dsn)
//	compiler, _ := conn.(ddl.Compiler)
//	err := migrate.Migrate(conn, compiler)
func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
	d := ddl.New(conn, ddlCompiler)
	if err := d.CreateTable(&itemcatalog.Specialty{}); err != nil {
		return err
	}
	if err := d.CreateTable(&itemcatalog.CatalogItem{}); err != nil {
		return err
	}
	if err := d.CreateTable(&itemcatalog.Agreement{}); err != nil {
		return err
	}
	return nil
}
```

Import path for consumers: `github.com/veltylabs/item_catalog/migrate`.
The root package keeps its existing name (`itemcatalog`, see `mcp.go:1`);
only the new subdirectory's package is called `migrate`. Do not put this
code in a new file inside the existing root package directory — it must be
a genuinely separate package, in its own `migrate/` subdirectory, for the
WASM-isolation reasoning above to hold.

### 3. New test file `migrate/migrate_test.go`

This test lives inside the new `migrate/` package directory (not in the
root `tests/` directory — `migrate` is its own package, and `webtyp.com/
auth/authority` tests its `Migrate` the same way, next to the code, not
under `tests/`). Mirror `webtyp.com/auth/authority`'s own `migrate_test.go`
exactly — same two fakes, same shape:

```go
package migrate_test

import (
	"testing"

	"github.com/veltylabs/item_catalog/migrate"
	"webtyp.com/ddl"
	"webtyp.com/model"
)

type dummyExecer struct{ calls []string }

func (d *dummyExecer) Exec(query string, args ...any) error {
	d.calls = append(d.calls, query)
	return nil
}

type dummyCompiler struct{}

func (d *dummyCompiler) CompileDDL(stmt ddl.Stmt, m model.Model) (string, []any, error) {
	return stmt.Table, nil, nil
}

func TestMigrate_CreatesAllThreeTablesInOrder(t *testing.T) {
	execer := &dummyExecer{}
	err := migrate.Migrate(execer, &dummyCompiler{})
	if err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	want := []string{"specialty", "catalog_item", "catalog_agreement"}
	if len(execer.calls) != len(want) {
		t.Fatalf("got %d Exec calls, want %d: %v", len(execer.calls), len(want), execer.calls)
	}
	for i, w := range want {
		if execer.calls[i] != w {
			t.Errorf("call %d: got table %q, want %q (order matters — FK dependencies)", i, execer.calls[i], w)
		}
	}
}
```

Check `Specialty{}.ModelName()`/`CatalogItem{}.ModelName()`/
`Agreement{}.ModelName()` for the exact strings the fake `CompileDDL` above
returns via `stmt.Table` before trusting the `want` slice literally — adjust
the three literals in `want` to whatever `ModelName()` actually returns if
it differs from `specialty`/`catalog_item`/`catalog_agreement`.

### 4. This repo's own `AGENTS.md` — replace the "Persistence" bullet

Replace the current bullet (search for `**Persistence**:`) with the text
block from `DDL_MIGRATE_ISOLATION_MASTER_PLAN.md` §2
(`https://github.com/webtyp/app-releases/blob/main/docs/DDL_MIGRATE_ISOLATION_MASTER_PLAN.md`),
quoted here verbatim so this plan stays self-contained:

```markdown
- **Persistence**: `New(db *orm.DB, deps Deps)` receives an already-connected
  `*orm.DB` and assumes its schema already exists — it never creates or
  alters tables, and never imports `webtyp.com/ddl`. Schema reconciliation
  lives in its own **subpackage**, `<module>/migrate` (`package migrate`),
  exporting `Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error`.
  Deliberately not called by `New`, and deliberately not in the module's
  root package: schema work is deploy-time work, run once from a migration
  binary (`cmd/migrate` in the composition-root app) — and keeping it in a
  separate package means nothing on a WASM build's import path (`view.go`,
  `init.go`, the root package itself) ever pulls `webtyp.com/ddl` into that
  binary, regardless of build tags.
  ```go
  // migrate/migrate.go
  package migrate

  import (
      "webtyp.com/ddl"

      thismodule "github.com/veltylabs/<this-module>"
  )

  func Migrate(conn ddl.Execer, ddlCompiler ddl.Compiler) error {
      d := ddl.New(conn, ddlCompiler)
      if err := d.CreateTable(&thismodule.CatalogItem{}); err != nil {
          return err
      }
      return nil
  }
  ```
  A module's own tests build `*orm.DB` over `storage/mem`
  (`orm.New(mem.New())`), which creates tables lazily on first `Exec` — they
  never call `Migrate`, and `New` never needs to type-assert for
  `ddl.Compiler` at all anymore.
```

## What this does NOT change

- No change to `Deps`, `Module`, or any public method signature other than
  the new `Migrate` function — every existing caller of `New` keeps
  compiling unchanged.
- No change to any op handler, model, or the `Specialty`/`CatalogItem`/
  `Agreement` table shapes themselves.
- `storage/mem`-backed tests (this module's existing suite) are unaffected —
  `mem` creates tables lazily and was never reached by the old
  `ddl.Compiler` type-assertion (it doesn't implement that interface), so
  removing the assertion from `New` changes nothing for them.

## Acceptance

- `grep -n "ddl\." mcp.go` → empty (the import and every call moved to
  `migrate/migrate.go`).
- `grep -rn "webtyp.com/ddl" *.go` (root package only, not `migrate/`) →
  empty.
- `gotest` passes, including the new `TestMigrate_CreatesAllThreeTablesInOrder`.
- `grep -n "Persistence" AGENTS.md` shows the new block, not the old
  type-assert-in-`New` one.

## Stages

| Stage | Files | Done when |
|---|---|---|
| 1 | `mcp.go` | DDL block removed from `New`, unused `ddl` import dropped |
| 2 | `migrate/migrate.go` (new package) | `Migrate` implemented exactly as specified, in its own subdirectory |
| 3 | `migrate/migrate_test.go` (new) | test passes, table order verified |
| 4 | `AGENTS.md` | "Persistence" bullet replaced verbatim |
