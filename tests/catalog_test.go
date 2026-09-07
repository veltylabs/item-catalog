package tests

import (
	"testing"

	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router/mock"
	"webtyp.com/storage/mem"
	itemcatalog "github.com/veltylabs/item_catalog"
)

func TestCatalog(t *testing.T) {
	db := orm.New(mem.New())

	pub := &MockPublisher{}
	module, err := itemcatalog.New(db, itemcatalog.Deps{
		IDs:       &MockIDGen{},
		Publisher: pub,
	})
	if err != nil {
		t.Fatal(err)
	}

	tenantID := "tenant-1"

	// Create specialty first
	spec, err := module.UpsertSpecialty(itemcatalog.Specialty{
		TenantId: tenantID,
		Prefix:   "md",
		Slug:     "medicina-general",
		Name:     "Medicina General",
	})
	if err != nil {
		t.Fatalf("failed to create specialty: %v", err)
	}

	// Test CreateItem
	item := itemcatalog.CatalogItem{
		TenantId:    tenantID,
		SpecialtyId: spec.Id,
		Sku:         "md1234",
		Name:        "Test Service",
		Type:        itemcatalog.ItemTypeService,
		Price:       10.5,
		Currency:    "USD",
		IsActive:    true,
	}

	created, err := module.CreateItem(item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}
	if created.Id == "" {
		t.Error("expected non-empty ID")
	}

	// Test duplicate SKU
	_, err = module.CreateItem(item)
	if err == nil {
		t.Error("expected error for duplicate SKU")
	}

	// Test FindBySKU
	found, err := module.FindBySKU(tenantID, "md1234")
	if err != nil {
		t.Errorf("failed to find item by SKU: %v", err)
	}
	if found.Id != created.Id {
		t.Errorf("expected ID %s, got %s", created.Id, found.Id)
	}

	// Test GetItem
	found, err = module.GetItem(tenantID, created.Id)
	if err != nil {
		t.Errorf("failed to get item: %v", err)
	}
	if found.Sku != "md1234" {
		t.Errorf("expected SKU md1234, got %s", found.Sku)
	}

	// Test ServiceExists
	exists, err := module.ServiceExists(tenantID, created.Id)
	if err != nil || !exists {
		t.Errorf("expected ServiceExists to be true, got %v, err: %v", exists, err)
	}

	// Test UpdateItem
	created.Name = "Updated Name"
	updated, err := module.UpdateItem(created)
	if err != nil {
		t.Errorf("failed to update item: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected Name Updated Name, got %s", updated.Name)
	}

	// Test ListItems
	items, err := module.ListItems(tenantID, itemcatalog.ItemFilter{})
	if err != nil {
		t.Errorf("failed to list items: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}

	// Create second specialty and item to test filter isolation between two real specialties
	spec2, err := module.UpsertSpecialty(itemcatalog.Specialty{
		TenantId: tenantID,
		Prefix:   "of",
		Slug:     "oftalmologia",
		Name:     "Oftalmología",
	})
	if err != nil {
		t.Fatalf("failed to create second specialty: %v", err)
	}

	item2 := itemcatalog.CatalogItem{
		TenantId:    tenantID,
		SpecialtyId: spec2.Id,
		Sku:         "of1234",
		Name:        "Ophthalmology Exam",
		Type:        itemcatalog.ItemTypeService,
		Price:       20.0,
		Currency:    "USD",
		IsActive:    true,
	}
	created2, err := module.CreateItem(item2)
	if err != nil {
		t.Fatalf("failed to create item 2: %v", err)
	}

	// Test ListItems with filter matching spec1
	filteredItems, err := module.ListItems(tenantID, itemcatalog.ItemFilter{
		SpecialtyId: spec.Id,
		Type:        itemcatalog.ItemTypeService,
		ActiveOnly:  true,
		Limit:       5,
		Offset:      0,
	})
	if err != nil {
		t.Errorf("failed to list items with filters: %v", err)
	}
	if len(filteredItems) != 1 || filteredItems[0].Id != created.Id {
		t.Errorf("expected 1 item with specialty spec1 ID %s, got %d items", created.Id, len(filteredItems))
	}

	// Test ListItems with filter matching spec2
	filteredItems2, err := module.ListItems(tenantID, itemcatalog.ItemFilter{
		SpecialtyId: spec2.Id,
	})
	if err != nil {
		t.Errorf("failed to list items with spec2 filter: %v", err)
	}
	if len(filteredItems2) != 1 || filteredItems2[0].Id != created2.Id {
		t.Errorf("expected 1 item with specialty spec2 ID %s, got %d items", created2.Id, len(filteredItems2))
	}

	// Clean up second item and specialty for rest of test flow
	if err := module.DeleteItem(tenantID, created2.Id); err != nil {
		t.Fatalf("failed to delete item 2: %v", err)
	}
	if err := module.DeleteSpecialty(tenantID, spec2.Id); err != nil {
		t.Fatalf("failed to delete specialty 2: %v", err)
	}

	// Test DeleteSpecialty while in use -> should fail with ErrSpecialtyInUse
	err = module.DeleteSpecialty(tenantID, spec.Id)
	if err != itemcatalog.ErrSpecialtyInUse {
		t.Errorf("expected ErrSpecialtyInUse, got %v", err)
	}

	// Test DeactivateItem
	err = module.DeactivateItem(tenantID, created.Id)
	if err != nil {
		t.Errorf("failed to deactivate item: %v", err)
	}
	deactivated, _ := module.GetItem(tenantID, created.Id)
	if deactivated.IsActive {
		t.Error("expected item to be inactive")
	}

	// Test ServiceExists for inactive
	exists, err = module.ServiceExists(tenantID, created.Id)
	if err != nil || exists {
		t.Errorf("expected ServiceExists to be false for inactive, got %v", exists)
	}

	// Test DeleteItem
	err = module.DeleteItem(tenantID, created.Id)
	if err != nil {
		t.Errorf("failed to delete item: %v", err)
	}
	_, err = module.GetItem(tenantID, created.Id)
	if err == nil {
		t.Error("expected error getting deleted item")
	}

	// Test DeleteSpecialty now that item is deleted -> should succeed
	err = module.DeleteSpecialty(tenantID, spec.Id)
	if err != nil {
		t.Errorf("failed to delete specialty after item deletion: %v", err)
	}
}

func TestSpecialtyUniquenessAndDefaults(t *testing.T) {
	db := orm.New(mem.New())
	module, err := itemcatalog.New(db, itemcatalog.Deps{IDs: &MockIDGen{}})
	if err != nil {
		t.Fatal(err)
	}

	tenant1 := "tenant-1"
	tenant2 := "tenant-2"

	// 1. Create specialty in tenant 1
	spec1, err := module.UpsertSpecialty(itemcatalog.Specialty{
		TenantId: tenant1,
		Prefix:   "of",
		Slug:     "oftalmologia",
		Name:     "Oftalmología",
	})
	if err != nil {
		t.Fatalf("failed to create specialty: %v", err)
	}

	// Verify is_published defaults to false
	if spec1.IsPublished {
		t.Error("expected is_published to default to false")
	}

	// 2. Duplicate prefix in tenant 1 -> rejected
	_, err = module.UpsertSpecialty(itemcatalog.Specialty{
		TenantId: tenant1,
		Prefix:   "of",
		Slug:     "oftalmologia-2",
		Name:     "Oftalmología Segunda",
	})
	if err != itemcatalog.ErrSpecialtyPrefixExists {
		t.Errorf("expected ErrSpecialtyPrefixExists, got %v", err)
	}

	// 3. Duplicate slug in tenant 1 -> rejected
	_, err = module.UpsertSpecialty(itemcatalog.Specialty{
		TenantId: tenant1,
		Prefix:   "ox",
		Slug:     "oftalmologia",
		Name:     "Oftalmología Segunda",
	})
	if err != itemcatalog.ErrSpecialtySlugExists {
		t.Errorf("expected ErrSpecialtySlugExists, got %v", err)
	}

	// 4. Same prefix in tenant 2 -> allowed
	spec2, err := module.UpsertSpecialty(itemcatalog.Specialty{
		TenantId: tenant2,
		Prefix:   "of",
		Slug:     "oftalmologia",
		Name:     "Oftalmología",
	})
	if err != nil {
		t.Fatalf("expected duplicate prefix across different tenants to be allowed, got: %v", err)
	}
	if spec2.Id == spec1.Id {
		t.Error("expected different IDs for different tenant specialties")
	}
}

func TestMigration(t *testing.T) {
	db := orm.New(mem.New())
	module, err := itemcatalog.New(db, itemcatalog.Deps{IDs: &MockIDGen{}})
	if err != nil {
		t.Fatal(err)
	}

	tenantID := "tenant-1"

	// Create item with canonical prefix "mdcong" directly in DB (simulating legacy data before migration)
	item1 := itemcatalog.CatalogItem{
		Id:       "unmigrated-1",
		TenantId: tenantID,
		Sku:      "mdcong",
		Name:     "Congreso Medico",
		Type:     itemcatalog.ItemTypeService,
		Price:    100.0,
		Currency: "USD",
		IsActive: true,
	}
	if err := db.Create(&item1); err != nil {
		t.Fatalf("failed to create unmigrated item 1: %v", err)
	}

	// Create item with unknown prefix "zz1234" directly in DB
	item2 := itemcatalog.CatalogItem{
		Id:       "unmigrated-2",
		TenantId: tenantID,
		Sku:      "zz1234",
		Name:     "Unknown Item",
		Type:     itemcatalog.ItemTypeService,
		Price:    50.0,
		Currency: "USD",
		IsActive: true,
	}
	if err := db.Create(&item2); err != nil {
		t.Fatalf("failed to create unmigrated item 2: %v", err)
	}

	// Run migration
	report, err := module.MigrateSpecialtiesAndItems(tenantID)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	if report.MigratedCount != 1 {
		t.Errorf("expected 1 migrated item, got %d", report.MigratedCount)
	}
	if len(report.UnmatchedSKUs) != 1 || report.UnmatchedSKUs[0] != "zz1234" {
		t.Errorf("expected [zz1234] in unmatched SKUs, got %v", report.UnmatchedSKUs)
	}

	// Verify item 1 has specialty_id set to "md" specialty ID
	migrated1, err := module.GetItem(tenantID, "unmigrated-1")
	if err != nil {
		t.Fatal(err)
	}
	if migrated1.SpecialtyId == "" {
		t.Error("expected item 1 specialty_id to be populated")
	}

	specMD, err := module.GetSpecialtyByPrefix(tenantID, "md")
	if err != nil {
		t.Fatal(err)
	}
	if migrated1.SpecialtyId != specMD.Id {
		t.Errorf("expected specialty_id %s, got %s", specMD.Id, migrated1.SpecialtyId)
	}

	// Verify item 2 specialty_id is still empty
	migrated2, err := module.GetItem(tenantID, "unmigrated-2")
	if err != nil {
		t.Fatal(err)
	}
	if migrated2.SpecialtyId != "" {
		t.Errorf("expected item 2 specialty_id to be empty, got %s", migrated2.SpecialtyId)
	}
}

func TestAgreements(t *testing.T) {
	db := orm.New(mem.New())

	pub := &MockPublisher{}
	module, err := itemcatalog.New(db, itemcatalog.Deps{
		IDs:       &MockIDGen{},
		Publisher: pub,
	})
	if err != nil {
		t.Fatal(err)
	}

	tenantID := "tenant-1"

	spec, err := module.UpsertSpecialty(itemcatalog.Specialty{
		TenantId: tenantID,
		Prefix:   "md",
		Slug:     "medicina-general",
		Name:     "Medicina General",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create CatalogItem first so foreign key validation is satisfied
	goodItem := itemcatalog.CatalogItem{
		TenantId:    tenantID,
		SpecialtyId: spec.Id,
		Sku:         "md-AG1",
		Name:        "Agreement Service",
		Type:        itemcatalog.ItemTypeService,
		Price:       100.0,
		Currency:    "USD",
		IsActive:    true,
	}
	createdItem, err := module.CreateItem(goodItem)
	if err != nil {
		t.Fatalf("failed to create parent item: %v", err)
	}
	itemID := createdItem.Id

	// Clear events before testing agreements
	pub.Events = nil

	// 1. Test UpsertAgreement - Create (Id == "")
	ag := itemcatalog.Agreement{
		TenantId:      tenantID,
		CatalogItemId: itemID,
		Insurer:       "FONASA",
		Code:          "F-12345",
		Price:         8500.0,
		IsActive:      true,
	}

	created, err := module.UpsertAgreement(ag)
	if err != nil {
		t.Fatalf("failed to create agreement: %v", err)
	}
	if created.Id == "" {
		t.Error("expected non-empty ID for created agreement")
	}
	if created.UpdatedAt == 0 {
		t.Error("expected UpdatedAt to be set")
	}
	if len(pub.Events) != 1 || pub.Events[0].Topic != "catalog.agreement.created" {
		t.Errorf("expected catalog.agreement.created event, got %v", pub.Events)
	}

	// 2. Test ListAgreements - Filter by Tenant and Item ID
	list, err := module.ListAgreements(tenantID, itemID)
	if err != nil {
		t.Fatalf("failed to list agreements: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 agreement, got %d", len(list))
	}
	if list[0].Id != created.Id {
		t.Errorf("expected ID %s, got %s", created.Id, list[0].Id)
	}

	// Filter by non-existent item id
	emptyList, err := module.ListAgreements(tenantID, "non-existent")
	if err != nil {
		t.Fatalf("failed to list agreements: %v", err)
	}
	if len(emptyList) != 0 {
		t.Errorf("expected 0 agreements, got %d", len(emptyList))
	}

	// 3. Test UpsertAgreement - Update (Id != "")
	created.Price = 9000.0
	updated, err := module.UpsertAgreement(created)
	if err != nil {
		t.Fatalf("failed to update agreement: %v", err)
	}
	if updated.Price != 9000.0 {
		t.Errorf("expected updated price to be 9000.0, got %f", updated.Price)
	}
	if len(pub.Events) != 2 || pub.Events[1].Topic != "catalog.agreement.updated" {
		t.Errorf("expected catalog.agreement.updated event, got %v", pub.Events)
	}

	// 4. Test DeleteAgreement
	err = module.DeleteAgreement(tenantID, updated.Id)
	if err != nil {
		t.Fatalf("failed to delete agreement: %v", err)
	}
	listAfterDelete, _ := module.ListAgreements(tenantID, itemID)
	if len(listAfterDelete) != 0 {
		t.Errorf("expected 0 agreements after deletion, got %d", len(listAfterDelete))
	}
	if len(pub.Events) != 3 || pub.Events[2].Topic != "catalog.agreement.deleted" {
		t.Errorf("expected catalog.agreement.deleted event, got %v", pub.Events)
	}
}

func TestModule_MountOperationsAndView(t *testing.T) {
	db := orm.New(mem.New())
	pub := &MockPublisher{}
	module, err := itemcatalog.New(db, itemcatalog.Deps{IDs: &MockIDGen{}, Publisher: pub})
	if err != nil {
		t.Fatal(err)
	}

	r := &mock.Router{}
	module.MountOperations(r)

	infos := r.Routes()
	var found bool
	for _, i := range infos {
		if i.Path == itemcatalog.OpUpsertItem || i.Path == "/"+itemcatalog.OpUpsertItem {
			found = true
			if i.Resource != "catalog_item" || i.Action != (model.Create|model.Update) {
				t.Errorf("RBAC mismatch for %s: %+v", itemcatalog.OpUpsertItem, i)
			}
		}
	}
	if !found {
		t.Fatalf("MountOperations did not register %s", itemcatalog.OpUpsertItem)
	}

	caller := &mock.Caller{}
	pres := itemcatalog.NewView(caller)
	if pres.Title() == "" {
		t.Error("expected a non-empty view title")
	}

	specPres := itemcatalog.NewSpecialtyView(caller)
	if specPres.Title() == "" {
		t.Error("expected a non-empty specialty view title")
	}
}
