package document

// Target is where a link leads: an address outside the site, or a page of it, whose address routes writes in
// each face's format.
//
//sumtype:decl
type Target interface {
	target()
}

// Address is an address outside the site, written whole.
type Address struct {
	URL string
}

// Page is a page of the site, by its kind and slug; the zero page is none.
type Page struct {
	Kind PageKind
	Slug string
}

// PageKind is what a page of the site is.
type PageKind int

const (
	// PageNone is no page: a page left unset is none
	PageNone PageKind = iota
	// PageResume is the résumé, the site's first page; it has no slug
	PageResume
	// PageProject is a project's own page, by the project's slug
	PageProject
)

func (Address) target() {}
func (Page) target()    {}
