package itemcatalog

import (
	"webtyp.com/model"
	"webtyp.com/router"
	"webtyp.com/view"
)

const titleCatalog = "Catálogo"
const titleSpecialties = "Especialidades"

// Item projects a CatalogItem as a view.Item — the ONLY view-specific code this record
// carries (view.Itemizer). The Presenter's internal index (built from this on Reload)
// replaces the old manual byID/WithFill lookup.
func (m *CatalogItem) Item() view.Item {
	return view.Item{ID: m.Id, Label: m.Name, Description: m.Sku}
}

// Item projects a Specialty as a view.Item (view.Itemizer).
func (s *Specialty) Item() view.Item {
	return view.Item{ID: s.Id, Label: s.Name, Description: s.Prefix}
}

// NewView builds the catalog item Presenter — the tech-agnostic engine a renderer (crudview,
// or any other) wraps. It is THIS module's job to build it (importing only view+model+router);
// the app decides which renderer draws it.
func NewView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{List: OpListItems, Save: OpUpsertItem, Delete: OpDeleteItem},
		func() model.ModelSlice { return &CatalogItemList{} })
	return view.New(b, &CatalogItem{}, view.WithTitle(titleCatalog))
}

// NewSpecialtyView builds the specialty Presenter.
func NewSpecialtyView(caller router.Caller) view.Presenter {
	b := view.NewCallerLister(caller,
		view.Ops{List: OpListSpecialties, Save: OpUpsertSpecialty, Delete: OpDeleteSpecialty},
		func() model.ModelSlice { return &SpecialtyList{} })
	return view.New(b, &Specialty{}, view.WithTitle(titleSpecialties))
}
