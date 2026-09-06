package tests

import (
	"testing"

	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	itemcatalog "github.com/veltylabs/item_catalog"
)

func TestTenantIsolation(t *testing.T) {
	db := orm.New(mem.New())
	module, err := itemcatalog.New(db, itemcatalog.Deps{
		IDs: &MockIDGen{},
	})
	if err != nil {
		t.Fatal(err)
	}

	tenantA := "tenant-A"
	tenantB := "tenant-B"

	// Create specialty for tenant A
	specA, err := module.UpsertSpecialty(itemcatalog.Specialty{
		TenantId: tenantA,
		Prefix:   "md",
		Slug:     "medicina-general",
		Name:     "Medicina General",
	})
	if err != nil {
		t.Fatalf("failed to create specialty A: %v", err)
	}

	// Tenant B cannot get tenant A's specialty
	_, err = module.GetSpecialty(tenantB, specA.Id)
	if err != itemcatalog.ErrSpecialtyNotFound {
		t.Errorf("expected ErrSpecialtyNotFound for tenant B getting tenant A's specialty, got %v", err)
	}

	// Create an item as tenant A
	itemA := itemcatalog.CatalogItem{
		TenantId:    tenantA,
		SpecialtyId: specA.Id,
		Sku:         "md-SKU-A",
		Name:        "Tenant A Product",
		Type:        itemcatalog.ItemTypeProduct,
		Price:       10.0,
		Currency:    "USD",
		IsActive:    true,
	}

	createdA, err := module.CreateItem(itemA)
	if err != nil {
		t.Fatalf("failed to create tenant A item: %v", err)
	}

	// 1. Tenant B must NOT be able to Get tenant A's item
	_, err = module.GetItem(tenantB, createdA.Id)
	if err != itemcatalog.ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant B getting tenant A's item, got %v", err)
	}

	// 2. Tenant B must NOT be able to Update tenant A's item
	createdA.TenantId = tenantB
	createdA.Name = "Hijacked Name"
	_, err = module.UpdateItem(createdA)
	if err != itemcatalog.ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant B updating tenant A's item, got %v", err)
	}

	// Revert tenant id on createdA helper
	createdA.TenantId = tenantA

	// 3. Tenant B must NOT be able to Deactivate tenant A's item
	err = module.DeactivateItem(tenantB, createdA.Id)
	if err != itemcatalog.ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant B deactivating tenant A's item, got %v", err)
	}

	// 4. Tenant B must NOT be able to Delete tenant A's item
	err = module.DeleteItem(tenantB, createdA.Id)
	if err != itemcatalog.ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant B deleting tenant A's item, got %v", err)
	}

	// Create an agreement for tenant A's item
	agA := itemcatalog.Agreement{
		TenantId:      tenantA,
		CatalogItemId: createdA.Id,
		Insurer:       "FONASA",
		Code:          "F-A",
		Price:         50.0,
		IsActive:      true,
	}

	createdAgA, err := module.UpsertAgreement(agA)
	if err != nil {
		t.Fatalf("failed to create agreement: %v", err)
	}

	// 5. Tenant B must NOT be able to Upsert/Update tenant A's agreement
	createdAgA.TenantId = tenantB
	createdAgA.Insurer = "FONASA Hijacked"
	_, err = module.UpsertAgreement(createdAgA)
	if err != itemcatalog.ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant B updating tenant A's agreement, got %v", err)
	}

	// Revert tenant ID
	createdAgA.TenantId = tenantA

	// 6. Tenant B must NOT be able to Delete tenant A's agreement
	err = module.DeleteAgreement(tenantB, createdAgA.Id)
	if err != itemcatalog.ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant B deleting tenant A's agreement, got %v", err)
	}
}
