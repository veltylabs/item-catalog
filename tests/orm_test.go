package tests

import (
	"testing"

	"webtyp.com/form"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router/mock"
	"webtyp.com/storage/mem"
	itemcatalog "github.com/veltylabs/item_catalog"
)

type fakeWriter struct{}

func (f fakeWriter) String(name, val string)                 {}
func (f fakeWriter) Int(name string, val int64)              {}
func (f fakeWriter) Float(name string, val float64)          {}
func (f fakeWriter) Bool(name string, val bool)              {}
func (f fakeWriter) Bytes(name string, val []byte)           {}
func (f fakeWriter) Null(name string)                        {}
func (f fakeWriter) Raw(name, val string)                    {}
func (f fakeWriter) Object(name string, val model.Encodable) {}
func (f fakeWriter) Array(name string, n int) model.ArrayWriter {
	return fakeArrayWriter{}
}

type fakeArrayWriter struct{}

func (f fakeArrayWriter) String(val string)          {}
func (f fakeArrayWriter) Int(val int64)              {}
func (f fakeArrayWriter) Float(val float64)          {}
func (f fakeArrayWriter) Bool(val bool)              {}
func (f fakeArrayWriter) Bytes(val []byte)           {}
func (f fakeArrayWriter) Object(val model.Encodable) {}
func (f fakeArrayWriter) Close()                     {}

type fakeReader struct{}

func (f fakeReader) String(name string) (string, bool) { return "", true }
func (f fakeReader) Int(name string) (int64, bool)     { return 0, true }
func (f fakeReader) Float(name string) (float64, bool) { return 0.0, true }
func (f fakeReader) Bool(name string) (bool, bool)     { return false, true }
func (f fakeReader) Bytes(name string) ([]byte, bool)  { return nil, true }
func (f fakeReader) Object(name string, into model.Decodable) bool {
	return true
}
func (f fakeReader) Array(name string) (model.ArrayReader, bool) { return fakeArrayReader{}, true }
func (f fakeReader) Raw(name string) (string, bool)              { return "", true }

type fakeArrayReader struct{}

func (f fakeArrayReader) Len() int                               { return 0 }
func (f fakeArrayReader) String(i int) string                    { return "" }
func (f fakeArrayReader) Int(i int) int64                        { return 0 }
func (f fakeArrayReader) Float(i int) float64                    { return 0.0 }
func (f fakeArrayReader) Bool(i int) bool                        { return false }
func (f fakeArrayReader) Bytes(i int) []byte                     { return nil }
func (f fakeArrayReader) Object(i int, into model.Decodable) bool { return true }

func TestORMBoilerplate(t *testing.T) {
	fw := fakeWriter{}
	fr := fakeReader{}

	// Specialty
	{
		m := &itemcatalog.Specialty{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)
		_ = m.Validate(0)

		l := &itemcatalog.SpecialtyList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// CatalogItem
	{
		m := &itemcatalog.CatalogItem{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)
		_ = m.Validate(0)
		_ = m.SchemaExt()

		l := &itemcatalog.CatalogItemList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// Agreement
	{
		m := &itemcatalog.Agreement{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)
		_ = m.Validate(0)
		_ = m.SchemaExt()

		l := &itemcatalog.AgreementList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// ItemFilter
	{
		m := &itemcatalog.ItemFilter{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.ItemFilterList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// ListItemsArgs
	{
		m := &itemcatalog.ListItemsArgs{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.ListItemsArgsList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// GetItemArgs
	{
		m := &itemcatalog.GetItemArgs{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.GetItemArgsList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// FindBySKUArgs
	{
		m := &itemcatalog.FindBySKUArgs{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.FindBySKUArgsList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// DeactivateItemArgs
	{
		m := &itemcatalog.DeactivateItemArgs{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.DeactivateItemArgsList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// DeleteItemArgs
	{
		m := &itemcatalog.DeleteItemArgs{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.DeleteItemArgsList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// ListAgreementsArgs
	{
		m := &itemcatalog.ListAgreementsArgs{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.ListAgreementsArgsList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// DeleteAgreementArgs
	{
		m := &itemcatalog.DeleteAgreementArgs{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.DeleteAgreementArgsList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// ListSpecialtiesArgs
	{
		m := &itemcatalog.ListSpecialtiesArgs{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.ListSpecialtiesArgsList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// GetSpecialtyArgs
	{
		m := &itemcatalog.GetSpecialtyArgs{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.GetSpecialtyArgsList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}

	// DeleteSpecialtyArgs
	{
		m := &itemcatalog.DeleteSpecialtyArgs{}
		_ = m.ModelName()
		_ = m.Schema()
		_ = m.Pointers()
		_ = m.IsNil()
		m.EncodeFields(fw)
		m.DecodeFields(fr)

		l := &itemcatalog.DeleteSpecialtyArgsList{}
		_ = l.Schema()
		_ = l.Pointers()
		_ = l.Len()
		_ = l.IsNil()
		_ = l.Append()
		_ = l.At(0)
		l.EncodeFields(fw)
		l.DecodeFields(fr)
	}
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	item := itemcatalog.CatalogItem{
		Id:          "item-123",
		TenantId:    "tenant-1",
		SpecialtyId: "spec-123",
		Sku:         "SKU-1",
		Name:        "Name 1",
		Description: "Desc 1",
		Type:        itemcatalog.ItemTypeService,
		Price:       150.0,
		Currency:    "USD",
		IsActive:    true,
		UpdatedAt:   123456,
	}

	ctx := &mock.Context{}
	if err := ctx.Encode(&item); err != nil {
		t.Fatal(err)
	}

	ctx.InBody = ctx.ResponseBody()

	var decoded itemcatalog.CatalogItem
	if err := ctx.Decode(&decoded); err != nil {
		t.Fatal(err)
	}

	if decoded != item {
		t.Errorf("expected %+v, got %+v", item, decoded)
	}
}

func TestSchemaPointersLength(t *testing.T) {
	spec := &itemcatalog.Specialty{}
	if len(spec.Schema()) != len(spec.Pointers()) {
		t.Errorf("expected len(Schema) %d == len(Pointers) %d", len(spec.Schema()), len(spec.Pointers()))
	}

	item := &itemcatalog.CatalogItem{}
	if len(item.Schema()) != len(item.Pointers()) {
		t.Errorf("expected len(Schema) %d == len(Pointers) %d", len(item.Schema()), len(item.Pointers()))
	}

	ag := &itemcatalog.Agreement{}
	if len(ag.Schema()) != len(ag.Pointers()) {
		t.Errorf("expected len(Schema) %d == len(Pointers) %d", len(ag.Schema()), len(ag.Pointers()))
	}
}

func TestValidateConstraints(t *testing.T) {
	goodItem := itemcatalog.CatalogItem{
		Id:          "item-1",
		TenantId:    "tenant-1",
		SpecialtyId: "spec-1",
		Sku:         "SKU-1",
		Name:        "Good Item",
		Type:        itemcatalog.ItemTypeService,
		Price:       100.0,
		Currency:    "USD",
		IsActive:    true,
		UpdatedAt:   12345,
	}

	if err := goodItem.Validate(model.ActionCreate); err != nil {
		t.Errorf("expected good item to be valid, got %v", err)
	}

	badItemEmptyName := goodItem
	badItemEmptyName.Name = ""
	if err := badItemEmptyName.Validate(model.ActionCreate); err == nil {
		t.Error("expected empty name to be rejected")
	}

	// For invalid type "X", we check via the service layer validation!
	db := orm.New(mem.New())
	module, _ := itemcatalog.New(db, itemcatalog.Deps{IDs: &MockIDGen{}})
	badItemInvalidType := goodItem
	badItemInvalidType.Type = "X"
	if _, err := module.CreateItem(badItemInvalidType); err == nil {
		t.Error("expected invalid type X to be rejected by CreateItem")
	}
}

func TestFormInputsGeneration(t *testing.T) {
	item := &itemcatalog.CatalogItem{}
	f, err := form.New("parent", item, &MockIDGen{})
	if err != nil {
		t.Fatalf("failed to create form: %v", err)
	}

	// We expect the form to contain editable input widgets for user-editable fields,
	// and NOT for machine-managed fields (id, tenant_id, updated_at).
	// specialty_id is a Ref field (model.Text()) not input.Input, so editable form input fields are:
	editableFields := []string{"sku", "name", "description", "type", "price", "currency", "is_active"}
	for _, field := range f.Inputs {
		name := field.FieldName()
		isEditable := false
		for _, ef := range editableFields {
			if ef == name {
				isEditable = true
				break
			}
		}
		if !isEditable {
			t.Errorf("field %s is machine-managed and should not be a form input widget", name)
		}
	}
}

func TestErrorAndModelName(t *testing.T) {
	// Cover ValidationError.Error
	ve := itemcatalog.ValidationError{Err: itemcatalog.ErrNotFound}
	_ = ve.Error()

	// Cover Module.ModelName
	db := orm.New(mem.New())
	module, _ := itemcatalog.New(db, itemcatalog.Deps{IDs: &MockIDGen{}})
	_ = module.ModelName()
}
