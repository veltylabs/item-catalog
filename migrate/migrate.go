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
