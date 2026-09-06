package itemcatalog

import (
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"
)

// Exported, greppable, the ONLY place these literals exist:
const (
	ItemTypeService = "S"
	ItemTypeProduct = "P"
)

// Helpers to satisfy ormc v0.9.24 AST static type parser:
var (
	Checkbox_FieldBool = input.Checkbox()
	Decimal_FieldFloat = input.Decimal()

	BaseBool_FieldBool = model.Bool()
	BaseInt_FieldInt   = model.Int()
)

type catalogItemType struct {
	input.Base
}

func (c *catalogItemType) Clone(parentID, name string) input.Input {
	clone := *c
	clone.InitBase(parentID, name, "radio")
	return &clone
}

func itemType() input.Input {
	t := &catalogItemType{}
	t.Letters = true
	t.Numbers = true
	t.Minimum = 1
	t.InitBase("", "", "radio")
	t.SetOptions(
		fmt.KeyValue{Key: ItemTypeService, Value: "Service"},
		fmt.KeyValue{Key: ItemTypeProduct, Value: "Product"},
	)
	return t
}

var SpecialtyModel = model.Definition{
	Name: "specialty",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "prefix", Type: input.Text(), NotNull: true,
			Permitted: model.Permitted{Letters: true, Minimum: 2, Maximum: 2}},
		{Name: "slug", Type: input.Text(), NotNull: true,
			Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-'}, Minimum: 1, Maximum: 60}},
		// input.Text()'s own baseline Extra ('.', ',', '(', ')') no incluye
		// '+' — "Dental (Odontología + Ortodoncia)" lo necesita, así que
		// este campo declara su propio whitelist para reemplazar el piso
		// del Kind (Field.Validate: un whitelist positivo propio reemplaza
		// el floor de Text, no lo suma). Tilde cubre las vocales acentuadas
		// sin listarlas una a una.
		{Name: "name", Type: input.Text(), NotNull: true,
			Permitted: model.Permitted{Letters: true, Tilde: true, Spaces: true, Extra: []rune{'+', '(', ')'}, Minimum: 1, Maximum: 100}},
		{Name: "description", Type: input.Textarea(), OmitEmpty: true},
		{Name: "position", Type: BaseInt_FieldInt, OmitEmpty: true},
		// NotNull deliberadamente ausente: para un bool, NotNull rechaza el
		// zero value (ValidateFields: field.NotNull && IsZeroPtr, y
		// IsZeroPtr para FieldBool es !*p) — con NotNull:true sería
		// imposible crear una especialidad SIN publicar, justo lo opuesto
		// de "cerrado por defecto". El false-por-defecto ya lo garantiza
		// el zero value de Go, no una restricción de esquema.
		{Name: "is_published", Type: Checkbox_FieldBool},
		{Name: "updated_at", Type: BaseInt_FieldInt, OmitEmpty: true},
	},
}

var CatalogItemModel = model.Definition{
	Name: "catalog_item",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "specialty_id", Type: model.Text(), Ref: &SpecialtyModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "sku", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-'}, Minimum: 1, Maximum: 50}},
		{Name: "name", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 1, Maximum: 255}},
		{Name: "description", Type: input.Textarea(), OmitEmpty: true},
		{Name: "type", Type: itemType(), NotNull: true},
		{Name: "price", Type: Decimal_FieldFloat, NotNull: true},
		{Name: "currency", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Letters: true, Minimum: 3, Maximum: 3}},
		{Name: "is_active", Type: Checkbox_FieldBool, NotNull: true},
		{Name: "updated_at", Type: BaseInt_FieldInt, OmitEmpty: true},
	},
}

var AgreementModel = model.Definition{
	Name: "catalog_agreement",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "catalog_item_id", Type: model.Text(), Ref: &CatalogItemModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "insurer", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Letters: true, Spaces: true, Minimum: 1, Maximum: 100}},
		{Name: "code", Type: input.Text(), OmitEmpty: true, Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-'}, Minimum: 1, Maximum: 50}},
		{Name: "price", Type: Decimal_FieldFloat, OmitEmpty: true},
		{Name: "is_active", Type: Checkbox_FieldBool, NotNull: true},
		{Name: "updated_at", Type: BaseInt_FieldInt, OmitEmpty: true},
	},
}

var ItemFilterModel = model.Definition{
	Name: "item_filter",
	Fields: model.Fields{
		{Name: "specialty_id", Type: model.Text()},
		{Name: "type", Type: model.Text()},
		{Name: "active_only", Type: BaseBool_FieldBool},
		{Name: "limit", Type: BaseInt_FieldInt},
		{Name: "offset", Type: BaseInt_FieldInt},
	},
}

var ListItemsArgsModel = model.Definition{
	Name: "list_items_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "specialty_id", Type: model.Text()},
		{Name: "type", Type: model.Text()},
		{Name: "active_only", Type: BaseBool_FieldBool},
		{Name: "limit", Type: BaseInt_FieldInt},
		{Name: "offset", Type: BaseInt_FieldInt},
	},
}

var GetItemArgsModel = model.Definition{
	Name: "get_item_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var FindBySKUArgsModel = model.Definition{
	Name: "find_by_sku_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "sku", Type: model.Text()},
	},
}

var DeactivateItemArgsModel = model.Definition{
	Name: "deactivate_item_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var DeleteItemArgsModel = model.Definition{
	Name: "delete_item_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var ListAgreementsArgsModel = model.Definition{
	Name: "list_agreements_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "catalog_item_id", Type: model.Text()},
	},
}

var DeleteAgreementArgsModel = model.Definition{
	Name: "delete_agreement_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var ListSpecialtiesArgsModel = model.Definition{
	Name: "list_specialties_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
	},
}

var GetSpecialtyArgsModel = model.Definition{
	Name: "get_specialty_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}

var DeleteSpecialtyArgsModel = model.Definition{
	Name: "delete_specialty_args",
	Fields: model.Fields{
		{Name: "tenant_id", Type: model.Text()},
		{Name: "id", Type: model.Text()},
	},
}
