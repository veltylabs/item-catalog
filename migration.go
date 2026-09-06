package itemcatalog

import (
	"webtyp.com/fmt"
)

type CanonicalSpecialty struct {
	Prefix string
	Slug   string
	Name   string
}

// CanonicalSpecialties defines the list of standard specialties derived from SKU prefixes.
var CanonicalSpecialties = []CanonicalSpecialty{
	{Prefix: "md", Slug: "medicina-general", Name: "Medicina General"},
	{Prefix: "do", Slug: "dental", Name: "Dental (Odontología + Ortodoncia)"},
	{Prefix: "tr", Slug: "traumatologia", Name: "Traumatología"},
	{Prefix: "po", Slug: "podologia", Name: "Podología"},
	{Prefix: "la", Slug: "laboratorio", Name: "Laboratorio"},
	{Prefix: "ec", Slug: "ecografia", Name: "Ecografía"},
	{Prefix: "gi", Slug: "ginecologia-y-obstetricia", Name: "Ginecología y Obstetricia"},
	{Prefix: "ca", Slug: "cardiologia", Name: "Cardiología"},
	{Prefix: "ga", Slug: "gastroenterologia", Name: "Gastroenterología"},
	{Prefix: "of", Slug: "oftalmologia", Name: "Oftalmología"},
	{Prefix: "ne", Slug: "neurologia", Name: "Neurología"},
	{Prefix: "ps", Slug: "psicologia", Name: "Psicología"},
	{Prefix: "de", Slug: "dermatologia", Name: "Dermatología"},
	{Prefix: "ra", Slug: "radiologia", Name: "Radiología"},
}

type MigrationReport struct {
	MigratedCount int
	UnmatchedSKUs []string
}

// MigrateSpecialtiesAndItems seeds canonical specialties for the tenant if missing,
// and maps existing CatalogItems to the corresponding specialty_id based on SKU prefix (sku[:2]).
// Unmatched items are left unassigned and returned in MigrationReport.UnmatchedSKUs.
func (m *Module) MigrateSpecialtiesAndItems(tenantId string) (MigrationReport, error) {
	report := MigrationReport{}

	prefixToID := make([]fmt.KeyValue, 0, len(CanonicalSpecialties))

	findID := func(prefix string) (string, bool) {
		for _, kv := range prefixToID {
			if kv.Key == prefix {
				return kv.Value, true
			}
		}
		return "", false
	}

	for _, cs := range CanonicalSpecialties {
		existing, err := m.GetSpecialtyByPrefix(tenantId, cs.Prefix)
		if err == nil {
			prefixToID = append(prefixToID, fmt.KeyValue{Key: cs.Prefix, Value: existing.Id})
		} else if err == ErrSpecialtyNotFound {
			spec := Specialty{
				TenantId:    tenantId,
				Prefix:      cs.Prefix,
				Slug:        cs.Slug,
				Name:        cs.Name,
				IsPublished: false,
			}
			created, err := m.UpsertSpecialty(spec)
			if err != nil {
				return report, fmt.Err("failed to seed specialty %s: %w", cs.Prefix, err)
			}
			prefixToID = append(prefixToID, fmt.KeyValue{Key: cs.Prefix, Value: created.Id})
		} else {
			return report, err
		}
	}

	items, err := m.ListItems(tenantId, ItemFilter{})
	if err != nil {
		return report, err
	}

	for _, item := range items {
		if len(item.Sku) < 2 {
			report.UnmatchedSKUs = append(report.UnmatchedSKUs, item.Sku)
			continue
		}
		prefix := item.Sku[:2]
		specID, found := findID(prefix)
		if !found {
			report.UnmatchedSKUs = append(report.UnmatchedSKUs, item.Sku)
			continue
		}

		item.SpecialtyId = specID
		_, err := m.UpdateItem(item)
		if err != nil {
			return report, fmt.Err("failed to update item %s: %w", item.Id, err)
		}
		report.MigratedCount++
	}

	return report, nil
}
