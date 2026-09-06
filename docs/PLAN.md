---
PLAN: "refactor!: migrate github.com/tinywasm -> webtyp.com + adopt view.NewCallerLister"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — `item_catalog`: WebTyp rename + new `view.New` API

The framework moved from `github.com/tinywasm/*` to the vanity path
`webtyp.com/*`, and **every framework module is now published** under the new
path (a Cloudflare Worker serves the `go-import` meta). `origin/main` of this
repo is still entirely on `github.com/tinywasm/*` and does not build against the
current framework. Two jobs:

- **A. The mechanical rename** `github.com/tinywasm/*` → `webtyp.com/*`.
- **B.** Adopt the new `webtyp.com/view` `view.New` signature.

The module import path **stays** `github.com/veltylabs/item_catalog` (only the
*framework* dependency changes org/host).

---

## A. Rename `github.com/tinywasm` → `webtyp.com`

### A1. Go source (`*.go`, including `tests/*.go`)

Replace the import path prefix **`github.com/tinywasm/`** → **`webtyp.com/`**
in every `.go` file. Known files on `origin/main` that reference it:
`mcp.go`, `model.go`, `model_orm.go`, `migration.go`, `view.go`, and under
`tests/`: `catalog_test.go`, `conformance_test.go`, `mcp_test.go`, `orm_test.go`,
`setup_test.go`, `tenant_test.go`. Grep to catch any others:

```
grep -rln 'github.com/tinywasm' --include='*.go' .
```

The package *selectors* do not change (`fmt.`, `orm.`, `view.`, `model.`,
`router.`, `time.`, `ddl.`, `events.` …) — only the path in the `import` block.

### A2. `go.mod` (module root)

`origin/main` `require` block:

```
github.com/tinywasm/ddl v0.0.7
github.com/tinywasm/events v0.0.2
github.com/tinywasm/fmt v0.25.5
github.com/tinywasm/input v0.0.2
github.com/tinywasm/model v0.1.4
github.com/tinywasm/orm v0.11.4
github.com/tinywasm/router v0.1.19
github.com/tinywasm/time v0.5.2
github.com/tinywasm/view v0.1.12
github.com/tinywasm/storage v0.0.2 // indirect
```

For **each** `github.com/tinywasm/<X>` requirement:

```
go mod edit -droprequire=github.com/tinywasm/<X>
go get webtyp.com/<X>@latest
```

Then `go mod tidy`. `@latest` is authoritative; for reference, the current
published tags are `ddl v0.0.15`, `events v0.0.3`, `fmt v1.0.0`, `input v0.0.6`,
`model v0.1.8`, `orm v0.12.1`, `router v0.1.31`, `time v0.5.5`, `view v0.5.2`,
`storage v0.0.7`. There must be **no `github.com/tinywasm/*`** left in `go.mod`
and **no `replace … => ../…`** pointing outside this module.

### A3. `tests/go.mod` (nested test module)

This repo has a second module at `tests/go.mod`. Its
`replace github.com/veltylabs/item_catalog => ../` line **stays** (in-repo, not
outside). Apply the same `-droprequire` / `go get webtyp.com/<X>@latest` /
`go mod tidy` to every `github.com/tinywasm/<X>` there. On `origin/main` those
are: `events fmt form json model orm router storage view` (direct) and
`ddl dom input time widget` (indirect). `form` → `webtyp.com/form@latest`,
`json` → `webtyp.com/json@latest`, `dom` → `webtyp.com/dom@latest`,
`widget` → `webtyp.com/widget@latest`.

### A4. Docs / config text

In `*.md`, `*.yml`, `*.yaml`: `github.com/tinywasm/` → `github.com/webtyp/`
(GitHub org rename — these are repo URLs, not import paths). Known: `README.md`,
`AGENTS.md`. Prose brand tokens `TinyWasm`/`TinyWASM` → `WebTyp`. Do **not**
touch `LICENSE` copyright lines or any upstream "TinyGo" compiler reference.

---

## B. New `view.New` API — use `view.NewCallerLister`

### The change

`webtyp.com/view`'s `view.New` no longer takes a `router.Caller`, an op-name
string, a slice factory, or `view.WithSaveOp` / `view.WithDeleteOp` (both
**removed**). New surface:

```go
// view.New builds the Presenter over a Lister. Its capabilities MIRROR the
// lister's: it is a Saver iff l implements view.Saver, a Deleter iff l
// implements view.Deleter, etc. No option can add a capability the lister
// lacks.
func New(l Lister, record model.Model, opts ...Option) Presenter

type Lister interface{ List() ([]model.Model, error) }
```

The framework ships the adapter that reproduces the **old op/caller behaviour
exactly** — `view.NewCallerLister`:

```go
// Ops.List is required. The returned Lister implements exactly the write
// capabilities whose op names are non-empty.
func NewCallerLister(c router.Caller, ops Ops, newList func() model.ModelSlice) Lister

type Ops struct{ List, Save, Update, Delete string }
```

### Reference implementations (read first)

- **`webtyp.com/auth` → `auth/view.go`** (canonical):
  ```go
  func NewView(caller router.Caller) view.Presenter {
      b := view.NewCallerLister(caller,
          view.Ops{List: OpListUsers, Save: OpUpsertUser, Delete: OpDeleteUser},
          func() model.ModelSlice { return &UserList{} })
      return view.New(b, &User{}, view.WithTitle("Usuarios"))
  }
  ```
- **`github.com/veltylabs/business_hours`** `view.go` + `lister.go` on `main`
  (a read-only sibling that just completed this same migration).

### Do — rewrite `view.go`

`NewView` and `NewSpecialtyView` keep their **exact current signature**
(`func(caller router.Caller) view.Presenter`) — callers and tests do not change.
Only the body changes. All three imports (`webtyp.com/model`,
`webtyp.com/router`, `webtyp.com/view`) **stay** — `model.ModelSlice` is still
used in the `newList` closures, `router.Caller` is still the parameter.

```go
// NewView builds the catalog-item Presenter — the tech-agnostic engine a
// renderer (crudview, or any other) wraps. This module builds it (view + model
// + router only); the app decides which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{List: OpListItems, Save: OpUpsertItem, Delete: OpDeleteItem},
		func() model.ModelSlice { return &CatalogItemList{} })
	return view.New(b, &CatalogItem{}, view.WithTitle(titleCatalog))
}

// NewSpecialtyView builds the specialty Presenter.
func NewSpecialtyView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{List: OpListSpecialties, Save: OpUpsertSpecialty, Delete: OpDeleteSpecialty},
		func() model.ModelSlice { return &SpecialtyList{} })
	return view.New(b, &Specialty{}, view.WithTitle(titleSpecialties))
}
```

`titleCatalog` / `titleSpecialties` are new **unexported string constants** in
this package (`const titleCatalog = "Catálogo"`,
`const titleSpecialties = "Especialidades"`) — do not inline the literals (they
were inline in the old code; fix that here). The `Op*` constants already exist
in `mcp.go` (`OpListItems`, `OpUpsertItem`, `OpDeleteItem`, `OpListSpecialties`,
`OpUpsertSpecialty`, `OpDeleteSpecialty`) — reuse them, do not redeclare.

`(*CatalogItem).Item()` and `(*Specialty).Item()` (the `view.Itemizer` impls)
stay exactly as they are.

### Tests

Because `NewView`/`NewSpecialtyView` signatures are unchanged, existing tests
that call `NewView(fakeCaller)` keep working. If `tests/conformance_test.go`
references a now-removed symbol (`view.WithSaveOp`, an old `view.New` arity, a
`view/conformance` helper that changed), adapt it minimally to the current
`webtyp.com/view` + `webtyp.com/view/conformance` surface — the assertion intent
(lists rows, is Saver, is Deleter) must be preserved.

---

## Verify

```
grep -rn 'github.com/tinywasm' --include='*.go' --include='go.mod' .        # empty
grep -rn 'github.com/tinywasm' tests/go.mod                                 # empty
grep -rn 'view.WithSaveOp\|view.WithDeleteOp' .                             # empty
grep -rn '=> \.\./\.\.' go.mod tests/go.mod                                 # empty
```

## Acceptance

- `go build ./...` → clean (module root).
- `(cd tests && go build ./... && go test ./...)` → clean + green.
- `gotest ./...` from the module root → all green (vet, race, tests).
- No `github.com/tinywasm/*` anywhere in `*.go` / `go.mod` / `tests/go.mod`.
- `view.go` no longer references `view.WithSaveOp` / `view.WithDeleteOp` / the
  5-arg `view.New`.
- `NewView` and `NewSpecialtyView` still have signature
  `func(caller router.Caller) view.Presenter`.

## Constraints

- **No behaviour change** — this is a path rename + an adapter swap. The op
  names carried in `view.Ops` are the same strings the old `WithSaveOp` /
  `WithDeleteOp` / list arg used.
- **No hardcoded strings** — view titles and every op name are named constants
  in this package (op names already are, in `mcp.go`).
- Keep every `//go:build` tag exactly as-is (this module is backend + WASM
  shared).
- Do not add `Update` to `view.Ops` — the old views had no update op; a
  non-empty `Update` would make the Presenter advertise a capability the backend
  never had.
