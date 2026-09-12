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
