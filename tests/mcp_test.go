package tests

import (
	"testing"

	"webtyp.com/json"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router"
	"webtyp.com/router/mock"
	"webtyp.com/storage/mem"
	itemcatalog "github.com/veltylabs/item_catalog"
)

func TestMCPOperations(t *testing.T) {
	db := orm.New(mem.New())
	pub := &MockPublisher{}
	module, err := itemcatalog.New(db, itemcatalog.Deps{
		IDs:       &MockIDGen{},
		Publisher: pub,
	})
	if err != nil {
		t.Fatal(err)
	}

	r := &mock.Router{}
	r.Configure(mock.Config{
		Authn: func(next router.HandlerFunc) router.HandlerFunc {
			return func(ctx router.Context) {
				ctx.SetUserID("user-123")
				next(ctx)
			}
		},
		Authorize: func(userID string, res model.Resource, act model.Action) bool {
			return true
		},
	})
	module.MountOperations(r)

	tenantID := "tenant-1"

	// 0. Upsert Specialty via OP
	spec := itemcatalog.Specialty{
		TenantId: tenantID,
		Prefix:   "md",
		Slug:     "medicina-general",
		Name:     "Medicina General",
	}
	var specBytes []byte
	_ = json.Encode(&spec, &specBytes)

	ctxSpecUpsert := &mock.Context{
		InBody: specBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpUpsertSpecialty, ctxSpecUpsert)
	if ctxSpecUpsert.Status != 0 && ctxSpecUpsert.Status != 200 {
		t.Fatalf("expected specialty upsert status success, got %d", ctxSpecUpsert.Status)
	}

	var createdSpec itemcatalog.Specialty
	if err := json.Decode(ctxSpecUpsert.ResponseBody(), &createdSpec); err != nil {
		t.Fatal(err)
	}
	if createdSpec.Id == "" {
		t.Fatal("expected non-empty Specialty ID")
	}

	// List Specialties via OP
	listSpecArgs := itemcatalog.ListSpecialtiesArgs{
		TenantId: tenantID,
	}
	var listSpecBytes []byte
	_ = json.Encode(&listSpecArgs, &listSpecBytes)

	ctxSpecList := &mock.Context{
		InBody: listSpecBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpListSpecialties, ctxSpecList)
	if ctxSpecList.Status != 0 && ctxSpecList.Status != 200 {
		t.Fatalf("expected specialty list status success, got %d", ctxSpecList.Status)
	}

	var specListResult itemcatalog.SpecialtyList
	if err := json.Decode(ctxSpecList.ResponseBody(), &specListResult); err != nil {
		t.Fatal(err)
	}
	if specListResult.Len() != 1 {
		t.Fatalf("expected 1 specialty, got %d", specListResult.Len())
	}

	// 1. Create item via OP
	item := itemcatalog.CatalogItem{
		TenantId:    tenantID,
		SpecialtyId: createdSpec.Id,
		Sku:         "md-M1",
		Name:        "Mcp Product",
		Type:        itemcatalog.ItemTypeProduct,
		Price:       100.0,
		Currency:    "USD",
		IsActive:    true,
	}
	var itemBytes []byte
	_ = json.Encode(&item, &itemBytes)

	ctxCreate := &mock.Context{
		InBody: itemBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpCreateItem, ctxCreate)
	if ctxCreate.Status != 0 && ctxCreate.Status != 200 {
		t.Fatalf("expected create status success, got %d", ctxCreate.Status)
	}

	var createdItem itemcatalog.CatalogItem
	if err := json.Decode(ctxCreate.ResponseBody(), &createdItem); err != nil {
		t.Fatal(err)
	}
	if createdItem.Id == "" {
		t.Fatal("expected non-empty ID")
	}

	// 2. Get item via OP
	getArgs := itemcatalog.GetItemArgs{
		TenantId: tenantID,
		Id:       createdItem.Id,
	}
	var getArgsBytes []byte
	_ = json.Encode(&getArgs, &getArgsBytes)

	ctxGet := &mock.Context{
		InBody: getArgsBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpGetItem, ctxGet)
	if ctxGet.Status != 0 && ctxGet.Status != 200 {
		t.Fatalf("expected get status success, got %d, body: %s", ctxGet.Status, string(ctxGet.ResponseBody()))
	}

	var gotItem itemcatalog.CatalogItem
	if err := json.Decode(ctxGet.ResponseBody(), &gotItem); err != nil {
		t.Fatal(err)
	}
	if gotItem.Sku != "md-M1" {
		t.Fatalf("expected SKU md-M1, got %s", gotItem.Sku)
	}

	// 3. Find by SKU via OP
	skuArgs := itemcatalog.FindBySKUArgs{
		TenantId: tenantID,
		Sku:      "md-M1",
	}
	var skuArgsBytes []byte
	_ = json.Encode(&skuArgs, &skuArgsBytes)

	ctxSku := &mock.Context{
		InBody: skuArgsBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpFindItemBySKU, ctxSku)
	if ctxSku.Status != 0 && ctxSku.Status != 200 {
		t.Fatalf("expected find sku status success, got %d", ctxSku.Status)
	}

	// 4. Update item via OP
	createdItem.Name = "Mcp Product Updated"
	var updateBytes []byte
	_ = json.Encode(&createdItem, &updateBytes)

	ctxUpdate := &mock.Context{
		InBody: updateBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpUpdateItem, ctxUpdate)
	if ctxUpdate.Status != 0 && ctxUpdate.Status != 200 {
		t.Fatalf("expected update status success, got %d", ctxUpdate.Status)
	}

	// 5. Upsert item via OP (Create with empty ID)
	upsertItem := itemcatalog.CatalogItem{
		TenantId:    tenantID,
		SpecialtyId: createdSpec.Id,
		Sku:         "md-M2",
		Name:        "Mcp Product 2",
		Type:        itemcatalog.ItemTypeService,
		Price:       200.0,
		Currency:    "USD",
		IsActive:    true,
	}
	var upsertBytes []byte
	_ = json.Encode(&upsertItem, &upsertBytes)

	ctxUpsert := &mock.Context{
		InBody: upsertBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpUpsertItem, ctxUpsert)
	if ctxUpsert.Status != 0 && ctxUpsert.Status != 200 {
		t.Fatalf("expected upsert status success, got %d", ctxUpsert.Status)
	}

	var upsertedItem itemcatalog.CatalogItem
	if err := json.Decode(ctxUpsert.ResponseBody(), &upsertedItem); err != nil {
		t.Fatal(err)
	}

	// 6. List items via OP
	listArgs := itemcatalog.ListItemsArgs{
		TenantId:    tenantID,
		SpecialtyId: createdSpec.Id,
		Type:        itemcatalog.ItemTypeService,
		ActiveOnly:  true,
		Limit:       10,
		Offset:      0,
	}
	var listBytes []byte
	_ = json.Encode(&listArgs, &listBytes)

	ctxList := &mock.Context{
		InBody: listBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpListItems, ctxList)
	if ctxList.Status != 0 && ctxList.Status != 200 {
		t.Fatalf("expected list status success, got %d", ctxList.Status)
	}

	var listResult itemcatalog.CatalogItemList
	if err := json.Decode(ctxList.ResponseBody(), &listResult); err != nil {
		t.Fatal(err)
	}
	if listResult.Len() != 1 {
		t.Fatalf("expected 1 item of type S, got %d", listResult.Len())
	}

	// 7. Deactivate item via OP
	deactivateArgs := itemcatalog.DeactivateItemArgs{
		TenantId: tenantID,
		Id:       createdItem.Id,
	}
	var deactivateBytes []byte
	_ = json.Encode(&deactivateArgs, &deactivateBytes)

	ctxDeactivate := &mock.Context{
		InBody: deactivateBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpDeactivateItem, ctxDeactivate)
	if ctxDeactivate.Status != 0 && ctxDeactivate.Status != 200 {
		t.Fatalf("expected deactivate status success, got %d", ctxDeactivate.Status)
	}

	// 8. Delete item via OP
	deleteArgs := itemcatalog.DeleteItemArgs{
		TenantId: tenantID,
		Id:       createdItem.Id,
	}
	var deleteBytes []byte
	_ = json.Encode(&deleteArgs, &deleteBytes)

	ctxDelete := &mock.Context{
		InBody: deleteBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpDeleteItem, ctxDelete)
	if ctxDelete.Status != 0 && ctxDelete.Status != 200 {
		t.Fatalf("expected delete status success, got %d", ctxDelete.Status)
	}

	// 9. Upsert agreement via OP (Create with empty ID)
	ag := itemcatalog.Agreement{
		TenantId:      tenantID,
		CatalogItemId: upsertedItem.Id,
		Insurer:       "Mcp Insurer",
		Code:          "M-123",
		Price:         150.0,
		IsActive:      true,
	}
	var agBytes []byte
	_ = json.Encode(&ag, &agBytes)

	ctxUpsertAg := &mock.Context{
		InBody: agBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpUpsertAgreement, ctxUpsertAg)
	if ctxUpsertAg.Status != 0 && ctxUpsertAg.Status != 200 {
		t.Fatalf("expected upsert agreement status success, got %d", ctxUpsertAg.Status)
	}

	var upsertedAg itemcatalog.Agreement
	if err := json.Decode(ctxUpsertAg.ResponseBody(), &upsertedAg); err != nil {
		t.Fatal(err)
	}

	// 10. List agreements via OP
	listAgArgs := itemcatalog.ListAgreementsArgs{
		TenantId:      tenantID,
		CatalogItemId: upsertedItem.Id,
	}
	var listAgBytes []byte
	_ = json.Encode(&listAgArgs, &listAgBytes)

	ctxListAg := &mock.Context{
		InBody: listAgBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpListAgreements, ctxListAg)
	if ctxListAg.Status != 0 && ctxListAg.Status != 200 {
		t.Fatalf("expected list agreements status success, got %d", ctxListAg.Status)
	}

	var listAgResult itemcatalog.AgreementList
	if err := json.Decode(ctxListAg.ResponseBody(), &listAgResult); err != nil {
		t.Fatal(err)
	}
	if listAgResult.Len() != 1 {
		t.Fatalf("expected 1 agreement, got %d", listAgResult.Len())
	}

	// 11. Delete agreement via OP
	deleteAgArgs := itemcatalog.DeleteAgreementArgs{
		TenantId: tenantID,
		Id:       upsertedAg.Id,
	}
	var deleteAgBytes []byte
	_ = json.Encode(&deleteAgArgs, &deleteAgBytes)

	ctxDeleteAg := &mock.Context{
		InBody: deleteAgBytes,
	}
	r.Invoke("OP", "/"+itemcatalog.OpDeleteAgreement, ctxDeleteAg)
	if ctxDeleteAg.Status != 0 && ctxDeleteAg.Status != 200 {
		t.Fatalf("expected delete agreement status success, got %d", ctxDeleteAg.Status)
	}
}

func TestMCPOperationsErrorPaths(t *testing.T) {
	db := orm.New(mem.New())
	module, err := itemcatalog.New(db, itemcatalog.Deps{
		IDs: &MockIDGen{},
	})
	if err != nil {
		t.Fatal(err)
	}

	r := &mock.Router{}
	r.Configure(mock.Config{
		Authn: func(next router.HandlerFunc) router.HandlerFunc {
			return func(ctx router.Context) {
				ctx.SetUserID("user-123")
				next(ctx)
			}
		},
		Authorize: func(userID string, res model.Resource, act model.Action) bool {
			return true
		},
	})
	module.MountOperations(r)

	// Decode error path
	ctxErr := &mock.Context{
		InBody: []byte(`invalid-json`),
	}
	r.Invoke("OP", "/"+itemcatalog.OpCreateItem, ctxErr)
	if ctxErr.Status != 400 {
		t.Fatalf("expected status 400 on decode error, got %d", ctxErr.Status)
	}
}
