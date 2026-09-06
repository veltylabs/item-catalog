package itemcatalog

import (
	"webtyp.com/ddl"
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router"
	"webtyp.com/time"
)

var ErrNotFound = fmt.Err("item not found")
var ErrAlreadyExists = fmt.Err("item already exists")

var ErrSpecialtyNotFound = fmt.Err("specialty not found")
var ErrSpecialtyInUse = fmt.Err("specialty in use")
var ErrSpecialtyPrefixExists = fmt.Err("specialty prefix already exists")
var ErrSpecialtySlugExists = fmt.Err("specialty slug already exists")

const (
	OpListSpecialties = "list_specialties"
	OpGetSpecialty    = "get_specialty"
	OpUpsertSpecialty = "upsert_specialty"
	OpDeleteSpecialty = "delete_specialty"

	OpListItems      = "list_catalog_items"
	OpGetItem        = "get_catalog_item"
	OpFindItemBySKU  = "find_item_by_sku"
	OpCreateItem     = "create_catalog_item"
	OpUpdateItem     = "update_catalog_item"
	OpUpsertItem     = "upsert_catalog_item"
	OpDeactivateItem = "deactivate_catalog_item"
	OpDeleteItem     = "delete_catalog_item"

	OpListAgreements  = "list_agreements"
	OpUpsertAgreement = "upsert_agreement"
	OpDeleteAgreement = "delete_agreement"

	TopicSpecialtyCreated = "catalog.specialty.created"
	TopicSpecialtyUpdated = "catalog.specialty.updated"
	TopicSpecialtyDeleted = "catalog.specialty.deleted"

	TopicItemCreated      = "catalog.item.created"
	TopicItemUpdated      = "catalog.item.updated"
	TopicItemDeactivated  = "catalog.item.deactivated"
	TopicItemDeleted      = "catalog.item.deleted"
	TopicAgreementCreated = "catalog.agreement.created"
	TopicAgreementUpdated = "catalog.agreement.updated"
	TopicAgreementDeleted = "catalog.agreement.deleted"
)

type ValidationError struct {
	Err error
}

func (v ValidationError) Error() string {
	return v.Err.Error()
}

type Deps struct {
	IDs       model.IDGenerator // requerido — el módulo NUNCA construye un generador
	Publisher events.Publisher  // opcional — nil desactiva la publicación de eventos
}

type Module struct {
	db  *orm.DB
	ids model.IDGenerator
	pub events.Publisher
}

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

// Specialty Service methods

func (m *Module) GetSpecialty(tenantId, id string) (Specialty, error) {
	var spec Specialty
	qb := m.db.Query(&spec).Where(Specialty_.Id).Eq(id).Where(Specialty_.TenantId).Eq(tenantId)
	_, err := ReadOneSpecialty(qb, &spec)
	if err != nil {
		if err == orm.ErrNotFound {
			return Specialty{}, ErrSpecialtyNotFound
		}
		return Specialty{}, err
	}
	return spec, nil
}

func (m *Module) GetSpecialtyByPrefix(tenantId, prefix string) (Specialty, error) {
	var spec Specialty
	qb := m.db.Query(&spec).Where(Specialty_.Prefix).Eq(prefix).Where(Specialty_.TenantId).Eq(tenantId)
	_, err := ReadOneSpecialty(qb, &spec)
	if err != nil {
		if err == orm.ErrNotFound {
			return Specialty{}, ErrSpecialtyNotFound
		}
		return Specialty{}, err
	}
	return spec, nil
}

func (m *Module) GetSpecialtyBySlug(tenantId, slug string) (Specialty, error) {
	var spec Specialty
	qb := m.db.Query(&spec).Where(Specialty_.Slug).Eq(slug).Where(Specialty_.TenantId).Eq(tenantId)
	_, err := ReadOneSpecialty(qb, &spec)
	if err != nil {
		if err == orm.ErrNotFound {
			return Specialty{}, ErrSpecialtyNotFound
		}
		return Specialty{}, err
	}
	return spec, nil
}

func (m *Module) ListSpecialties(tenantId string) ([]Specialty, error) {
	var spec Specialty
	qb := m.db.Query(&spec).Where(Specialty_.TenantId).Eq(tenantId)
	results, err := ReadAllSpecialty(qb)
	if err != nil {
		return nil, err
	}
	list := make([]Specialty, len(results))
	for i, r := range results {
		list[i] = *r
	}
	return list, nil
}

func (m *Module) UpsertSpecialty(spec Specialty) (Specialty, error) {
	action := model.ActionCreate
	if spec.Id != "" {
		action = model.ActionUpdate
	}
	if spec.Id == "" {
		spec.Id = m.ids.NewID()
	}
	spec.UpdatedAt = time.Now()

	if err := spec.Validate(action); err != nil {
		return Specialty{}, ValidationError{Err: err}
	}

	// Validate prefix uniqueness per tenant
	existingPrefix, err := m.GetSpecialtyByPrefix(spec.TenantId, spec.Prefix)
	if err == nil {
		if existingPrefix.Id != spec.Id {
			return Specialty{}, ErrSpecialtyPrefixExists
		}
	} else if err != ErrSpecialtyNotFound {
		return Specialty{}, err
	}

	// Validate slug uniqueness per tenant
	existingSlug, err := m.GetSpecialtyBySlug(spec.TenantId, spec.Slug)
	if err == nil {
		if existingSlug.Id != spec.Id {
			return Specialty{}, ErrSpecialtySlugExists
		}
	} else if err != ErrSpecialtyNotFound {
		return Specialty{}, err
	}

	if action == model.ActionCreate {
		if err := m.db.Create(&spec); err != nil {
			return Specialty{}, err
		}
		if m.pub != nil {
			m.pub.Publish(events.Event{Topic: TopicSpecialtyCreated, Payload: &spec})
		}
		return spec, nil
	}

	// Verify specialty exists and belongs to tenant
	_, err = m.GetSpecialty(spec.TenantId, spec.Id)
	if err != nil {
		return Specialty{}, err
	}

	if err := m.db.Update(&spec, orm.Eq(Specialty_.Id, spec.Id), orm.Eq(Specialty_.TenantId, spec.TenantId)); err != nil {
		return Specialty{}, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicSpecialtyUpdated, Payload: &spec})
	}
	return spec, nil
}

func (m *Module) DeleteSpecialty(tenantId, id string) error {
	spec, err := m.GetSpecialty(tenantId, id)
	if err != nil {
		return err
	}

	// Check if any catalog item is using this specialty
	var item CatalogItem
	qb := m.db.Query(&item).Where(CatalogItem_.TenantId).Eq(tenantId).Where(CatalogItem_.SpecialtyId).Eq(id)
	items, err := ReadAllCatalogItem(qb)
	if err != nil {
		return err
	}
	if len(items) > 0 {
		return ErrSpecialtyInUse
	}

	if err := m.db.Delete(&spec, orm.Eq(Specialty_.Id, id), orm.Eq(Specialty_.TenantId, tenantId)); err != nil {
		return err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicSpecialtyDeleted, Payload: &spec})
	}
	return nil
}

// Service methods

func (m *Module) GetItem(tenantId, id string) (CatalogItem, error) {
	var item CatalogItem
	qb := m.db.Query(&item).Where(CatalogItem_.Id).Eq(id).Where(CatalogItem_.TenantId).Eq(tenantId)
	_, err := ReadOneCatalogItem(qb, &item)
	if err != nil {
		if err == orm.ErrNotFound {
			return CatalogItem{}, ErrNotFound
		}
		return CatalogItem{}, err
	}
	return item, nil
}

func (m *Module) FindBySKU(tenantId, sku string) (CatalogItem, error) {
	var item CatalogItem
	qb := m.db.Query(&item).Where(CatalogItem_.Sku).Eq(sku).Where(CatalogItem_.TenantId).Eq(tenantId)
	_, err := ReadOneCatalogItem(qb, &item)
	if err != nil {
		if err == orm.ErrNotFound {
			return CatalogItem{}, ErrNotFound
		}
		return CatalogItem{}, err
	}
	return item, nil
}

func (m *Module) ListItems(tenantId string, filter ItemFilter) ([]CatalogItem, error) {
	var item CatalogItem
	qb := m.db.Query(&item).Where(CatalogItem_.TenantId).Eq(tenantId)
	if filter.SpecialtyId != "" {
		qb = qb.Where(CatalogItem_.SpecialtyId).Eq(filter.SpecialtyId)
	}
	if filter.Type != "" {
		qb = qb.Where(CatalogItem_.Type).Eq(filter.Type)
	}
	if filter.ActiveOnly {
		qb = qb.Where(CatalogItem_.IsActive).Eq(true)
	}
	if filter.Limit > 0 {
		qb = qb.Limit(int(filter.Limit))
	}
	if filter.Offset > 0 {
		qb = qb.Offset(int(filter.Offset))
	}
	results, err := ReadAllCatalogItem(qb)
	if err != nil {
		return nil, err
	}
	items := make([]CatalogItem, len(results))
	for i, r := range results {
		items[i] = *r
	}
	return items, nil
}

func (m *Module) CreateItem(item CatalogItem) (CatalogItem, error) {
	item.Id = m.ids.NewID()
	item.UpdatedAt = time.Now()

	if err := item.Validate(model.ActionCreate); err != nil {
		return CatalogItem{}, ValidationError{Err: err}
	}
	if item.Type != ItemTypeService && item.Type != ItemTypeProduct {
		return CatalogItem{}, ValidationError{Err: fmt.Err("invalid item type: %s", item.Type)}
	}

	// Validate SKU uniqueness per tenant
	existing, err := m.FindBySKU(item.TenantId, item.Sku)
	if err == nil {
		if existing.Id != "" {
			return CatalogItem{}, ErrAlreadyExists
		}
	} else if err != ErrNotFound {
		return CatalogItem{}, err
	}

	if err := m.db.Create(&item); err != nil {
		return CatalogItem{}, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicItemCreated, Payload: &item})
	}
	return item, nil
}

func (m *Module) UpdateItem(item CatalogItem) (CatalogItem, error) {
	if err := item.Validate(model.ActionUpdate); err != nil {
		return CatalogItem{}, ValidationError{Err: err}
	}
	if item.Type != ItemTypeService && item.Type != ItemTypeProduct {
		return CatalogItem{}, ValidationError{Err: fmt.Err("invalid item type: %s", item.Type)}
	}

	// Verify item exists and belongs to tenant
	_, err := m.GetItem(item.TenantId, item.Id)
	if err != nil {
		return CatalogItem{}, err
	}

	item.UpdatedAt = time.Now()
	if err := m.db.Update(&item, orm.Eq(CatalogItem_.Id, item.Id), orm.Eq(CatalogItem_.TenantId, item.TenantId)); err != nil {
		return CatalogItem{}, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicItemUpdated, Payload: &item})
	}
	return item, nil
}

func (m *Module) DeactivateItem(tenantId, id string) error {
	item, err := m.GetItem(tenantId, id)
	if err != nil {
		return err
	}
	item.IsActive = false
	item.UpdatedAt = time.Now()
	if err := m.db.Update(&item, orm.Eq(CatalogItem_.Id, item.Id), orm.Eq(CatalogItem_.TenantId, item.TenantId)); err != nil {
		return err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicItemDeactivated, Payload: &item})
	}
	return nil
}

func (m *Module) DeleteItem(tenantId, id string) error {
	item, err := m.GetItem(tenantId, id)
	if err != nil {
		return err
	}
	if err := m.db.Delete(&item, orm.Eq(CatalogItem_.Id, item.Id), orm.Eq(CatalogItem_.TenantId, item.TenantId)); err != nil {
		return err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicItemDeleted, Payload: &item})
	}
	return nil
}

func (m *Module) ServiceExists(tenantId, serviceId string) (bool, error) {
	item, err := m.GetItem(tenantId, serviceId)
	if err != nil {
		if err == ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return item.Type == ItemTypeService && item.IsActive, nil
}

func (m *Module) ListAgreements(tenantId, catalogItemId string) ([]Agreement, error) {
	var a Agreement
	qb := m.db.Query(&a).Where(Agreement_.TenantId).Eq(tenantId)
	if catalogItemId != "" {
		qb = qb.Where(Agreement_.CatalogItemId).Eq(catalogItemId)
	}
	results, err := ReadAllAgreement(qb)
	if err != nil {
		return nil, err
	}
	items := make([]Agreement, len(results))
	for i, r := range results {
		items[i] = *r
	}
	return items, nil
}

func (m *Module) GetAgreement(tenantId, id string) (Agreement, error) {
	var a Agreement
	qb := m.db.Query(&a).Where(Agreement_.Id).Eq(id).Where(Agreement_.TenantId).Eq(tenantId)
	_, err := ReadOneAgreement(qb, &a)
	if err != nil {
		if err == orm.ErrNotFound {
			return Agreement{}, ErrNotFound
		}
		return Agreement{}, err
	}
	return a, nil
}

func (m *Module) UpsertAgreement(a Agreement) (Agreement, error) {
	action := model.ActionCreate
	if a.Id != "" {
		action = model.ActionUpdate
	}
	if a.Id == "" {
		a.Id = m.ids.NewID()
	}
	a.UpdatedAt = time.Now()

	if err := a.Validate(action); err != nil {
		return Agreement{}, ValidationError{Err: err}
	}

	if action == model.ActionCreate {
		if err := m.db.Create(&a); err != nil {
			return Agreement{}, err
		}
		if m.pub != nil {
			m.pub.Publish(events.Event{Topic: TopicAgreementCreated, Payload: &a})
		}
		return a, nil
	}

	// Verify agreement exists and belongs to tenant
	_, err := m.GetAgreement(a.TenantId, a.Id)
	if err != nil {
		return Agreement{}, err
	}

	if err := m.db.Update(&a, orm.Eq(Agreement_.Id, a.Id), orm.Eq(Agreement_.TenantId, a.TenantId)); err != nil {
		return Agreement{}, err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicAgreementUpdated, Payload: &a})
	}
	return a, nil
}

func (m *Module) DeleteAgreement(tenantId, id string) error {
	a, err := m.GetAgreement(tenantId, id)
	if err != nil {
		return err
	}

	if err := m.db.Delete(&a, orm.Eq(Agreement_.Id, id), orm.Eq(Agreement_.TenantId, tenantId)); err != nil {
		return err
	}
	if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicAgreementDeleted, Payload: &a})
	}
	return nil
}

func (m *Module) ModelName() string { return "item_catalog" }

func (m *Module) MountOps(reg router.OpRegistry) {
	reg.Op(OpListSpecialties, m.opListSpecialties).Requires("specialty", model.Read).Accepts(&ListSpecialtiesArgs{})
	reg.Op(OpGetSpecialty, m.opGetSpecialty).Requires("specialty", model.Read).Accepts(&GetSpecialtyArgs{})
	reg.Op(OpUpsertSpecialty, m.opUpsertSpecialty).Requires("specialty", model.Create|model.Update).Accepts(&Specialty{})
	reg.Op(OpDeleteSpecialty, m.opDeleteSpecialty).Requires("specialty", model.Delete).Accepts(&DeleteSpecialtyArgs{})

	reg.Op(OpListItems, m.opListItems).Requires("catalog_item", model.Read).Accepts(&ListItemsArgs{})
	reg.Op(OpGetItem, m.opGetItem).Requires("catalog_item", model.Read).Accepts(&GetItemArgs{})
	reg.Op(OpFindItemBySKU, m.opFindItemBySKU).Requires("catalog_item", model.Read).Accepts(&FindBySKUArgs{})
	reg.Op(OpCreateItem, m.opCreateItem).Requires("catalog_item", model.Create).Accepts(&CatalogItem{})
	reg.Op(OpUpdateItem, m.opUpdateItem).Requires("catalog_item", model.Update).Accepts(&CatalogItem{})
	reg.Op(OpUpsertItem, m.opUpsertItem).Requires("catalog_item", model.Create|model.Update).Accepts(&CatalogItem{})
	reg.Op(OpDeactivateItem, m.opDeactivateItem).Requires("catalog_item", model.Update).Accepts(&DeactivateItemArgs{})
	reg.Op(OpDeleteItem, m.opDeleteItem).Requires("catalog_item", model.Delete).Accepts(&DeleteItemArgs{})

	reg.Op(OpListAgreements, m.opListAgreements).Requires("catalog_agreement", model.Read).Accepts(&ListAgreementsArgs{})
	reg.Op(OpUpsertAgreement, m.opUpsertAgreement).Requires("catalog_agreement", model.Create|model.Update).Accepts(&Agreement{})
	reg.Op(OpDeleteAgreement, m.opDeleteAgreement).Requires("catalog_agreement", model.Delete).Accepts(&DeleteAgreementArgs{})
}

var _ router.OpModule = (*Module)(nil)

func (m *Module) opListSpecialties(ctx router.Context) {
	var args ListSpecialtiesArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	list, err := m.ListSpecialties(args.TenantId)
	if err != nil {
		ctx.WriteStatus(500)
		return
	}
	resList := make(SpecialtyList, len(list))
	for i := range list {
		resList[i] = &list[i]
	}
	if err := ctx.Encode(&resList); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opGetSpecialty(ctx router.Context) {
	var args GetSpecialtyArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	spec, err := m.GetSpecialty(args.TenantId, args.Id)
	if err != nil {
		if err == ErrSpecialtyNotFound || err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&spec); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opUpsertSpecialty(ctx router.Context) {
	var spec Specialty
	if err := ctx.Decode(&spec); err != nil {
		ctx.WriteStatus(400)
		return
	}
	out, err := m.UpsertSpecialty(spec)
	if err != nil {
		if _, ok := err.(ValidationError); ok {
			ctx.WriteStatus(400)
		} else if err == ErrSpecialtyPrefixExists || err == ErrSpecialtySlugExists {
			ctx.WriteStatus(409)
		} else if err == ErrSpecialtyNotFound || err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&out); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opDeleteSpecialty(ctx router.Context) {
	var args DeleteSpecialtyArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	if err := m.DeleteSpecialty(args.TenantId, args.Id); err != nil {
		if err == ErrSpecialtyNotFound || err == ErrNotFound {
			ctx.WriteStatus(404)
		} else if err == ErrSpecialtyInUse {
			ctx.WriteStatus(400)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	ctx.WriteStatus(200)
}

func (m *Module) opListItems(ctx router.Context) {
	var args ListItemsArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	filter := ItemFilter{SpecialtyId: args.SpecialtyId, Type: args.Type, ActiveOnly: args.ActiveOnly, Limit: args.Limit, Offset: args.Offset}
	items, err := m.ListItems(args.TenantId, filter)
	if err != nil {
		ctx.WriteStatus(500)
		return
	}
	list := make(CatalogItemList, len(items))
	for i := range items {
		list[i] = &items[i]
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opGetItem(ctx router.Context) {
	var args GetItemArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	item, err := m.GetItem(args.TenantId, args.Id)
	if err != nil {
		if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&item); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opFindItemBySKU(ctx router.Context) {
	var args FindBySKUArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	item, err := m.FindBySKU(args.TenantId, args.Sku)
	if err != nil {
		if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&item); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opCreateItem(ctx router.Context) {
	var item CatalogItem
	if err := ctx.Decode(&item); err != nil {
		ctx.WriteStatus(400)
		return
	}
	created, err := m.CreateItem(item)
	if err != nil {
		if _, ok := err.(ValidationError); ok {
			ctx.WriteStatus(400)
		} else if err == ErrAlreadyExists {
			ctx.WriteStatus(409)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&created); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opUpdateItem(ctx router.Context) {
	var item CatalogItem
	if err := ctx.Decode(&item); err != nil {
		ctx.WriteStatus(400)
		return
	}
	updated, err := m.UpdateItem(item)
	if err != nil {
		if _, ok := err.(ValidationError); ok {
			ctx.WriteStatus(400)
		} else if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&updated); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opUpsertItem(ctx router.Context) {
	var item CatalogItem
	if err := ctx.Decode(&item); err != nil {
		ctx.WriteStatus(400)
		return
	}
	var out CatalogItem
	var err error
	if item.Id == "" {
		out, err = m.CreateItem(item)
	} else {
		out, err = m.UpdateItem(item)
	}
	if err != nil {
		if _, ok := err.(ValidationError); ok {
			ctx.WriteStatus(400)
		} else if err == ErrAlreadyExists {
			ctx.WriteStatus(409)
		} else if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&out); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opDeactivateItem(ctx router.Context) {
	var args DeactivateItemArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	if err := m.DeactivateItem(args.TenantId, args.Id); err != nil {
		if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	ctx.WriteStatus(200)
}

func (m *Module) opDeleteItem(ctx router.Context) {
	var args DeleteItemArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	if err := m.DeleteItem(args.TenantId, args.Id); err != nil {
		if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	ctx.WriteStatus(200)
}

func (m *Module) opListAgreements(ctx router.Context) {
	var args ListAgreementsArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	items, err := m.ListAgreements(args.TenantId, args.CatalogItemId)
	if err != nil {
		ctx.WriteStatus(500)
		return
	}
	list := make(AgreementList, len(items))
	for i := range items {
		list[i] = &items[i]
	}
	if err := ctx.Encode(&list); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opUpsertAgreement(ctx router.Context) {
	var a Agreement
	if err := ctx.Decode(&a); err != nil {
		ctx.WriteStatus(400)
		return
	}
	out, err := m.UpsertAgreement(a)
	if err != nil {
		if _, ok := err.(ValidationError); ok {
			ctx.WriteStatus(400)
		} else if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	if err := ctx.Encode(&out); err != nil {
		ctx.WriteStatus(500)
	}
}

func (m *Module) opDeleteAgreement(ctx router.Context) {
	var args DeleteAgreementArgs
	if err := ctx.Decode(&args); err != nil {
		ctx.WriteStatus(400)
		return
	}
	if err := m.DeleteAgreement(args.TenantId, args.Id); err != nil {
		if err == ErrNotFound {
			ctx.WriteStatus(404)
		} else {
			ctx.WriteStatus(500)
		}
		return
	}
	ctx.WriteStatus(200)
}
