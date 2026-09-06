# item-catalog
<img src="docs/img/badges.svg">

Unified catalog of services and products. Designed for the Velty ecosystem.

## Quick Start

```go
import (
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
	itemcatalog "github.com/veltylabs/item_catalog"
)

// In a real application, inject the database connection chosen by the composer app
db := orm.New(mem.New())
idGenerator := &MyIDGenerator{} // implements model.IDGenerator

module, _ := itemcatalog.New(db, itemcatalog.Deps{
	IDs: idGenerator,
})

// Create specialty
specialty, _ := module.UpsertSpecialty(itemcatalog.Specialty{
	TenantId: "my-tenant",
	Prefix:   "md",
	Slug:     "medicina-general",
	Name:     "Medicina General",
})

// Use the module
item, _ := module.CreateItem(itemcatalog.CatalogItem{
	TenantId:    "my-tenant",
	SpecialtyId: specialty.Id,
	Sku:         "md0001",
	Name:        "Initial Product",
	Type:        itemcatalog.ItemTypeProduct,
	Price:       100,
	Currency:    "USD",
	IsActive:    true,
})
```

## Ops
| Op Name | Description |
|-----------|-------------|
| `list_specialties` | List specialties for a tenant |
| `get_specialty` | Get specialty by ID |
| `upsert_specialty` | Create or update specialty |
| `delete_specialty` | Delete specialty |
| `list_catalog_items` | List items for a tenant |
| `get_catalog_item` | Get item by ID |
| `find_item_by_sku` | Find item by SKU |
| `create_catalog_item` | Create new item |
| `update_catalog_item` | Update existing item |
| `upsert_catalog_item` | Create or update item |
| `deactivate_catalog_item` | Soft-delete item |
| `delete_catalog_item` | Hard-delete item |
| `list_agreements` | List billing agreements |
| `upsert_agreement` | Create or update agreement |
| `delete_agreement` | Delete agreement |

## Key Files
| File | Purpose |
|------|---------|
| `model.go` | Data structures and schemas |
| `mcp.go` | Core logic and transport operations |
| `migration.go` | SKU prefix to specialty migration |
| `model_orm.go` | Generated ORM helpers |
| `tests/` | Functional and unit tests |
