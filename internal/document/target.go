package document

import "strings"

// Target is where a link leads: an address outside the site, a page of it, or a mailbox, whose address routes
// writes in each face's format.
//
//sumtype:decl
type Target interface {
	target()
}

// Address is an address outside the site, written whole.
type Address struct {
	URL string
}

// Page is a page of the site, by its kind and slug, and the tab of it shown alone; the zero page is none.
type Page struct {
	Kind PageKind
	Slug string
	// Tab is the name of the tab shown alone (Tab.Slug); none where the page is shown whole
	Tab string
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
	// PageIndex is the site's index, the list of its pages an agent reads first; it has no slug
	PageIndex
	// PageHealth is the hidden health report, after Neovim's :checkhealth; it has no slug
	PageHealth
	// PageSkills is the hidden list of every skill, after lazy.nvim's :Lazy; it has no slug
	PageSkills
	// PagePost is a post of the blog, by its slug; it has no tabs
	PagePost
)

// Mail is a mailbox a link writes to, by its address as the content writes it.
type Mail struct {
	Address string
}

// MailAt is the at a mail address parts at, between its mailbox and its host, which a face draws between the two
const MailAt = "@"

// Parts are the mail's mailbox and its host, parted at its last at, since a host holds none; not parted where it
// holds no at.
func (mail Mail) Parts() (box, host string, parted bool) {
	at := strings.LastIndex(mail.Address, MailAt)
	if at < 0 {
		return mail.Address, "", false
	}
	return mail.Address[:at], mail.Address[at+len(MailAt):], true
}

// LinkOf is the link spans are, where they are one link alone, with no words beside it.
func LinkOf(spans []Span) (Link, bool) {
	if len(spans) != 1 {
		return Link{Target: nil, Spans: nil}, false
	}
	link, isLink := spans[0].(Link)
	return link, isLink
}

// MailOf is the mail spans write to, where they are one link to a mailbox.
func MailOf(spans []Span) (Mail, bool) {
	link, isLink := LinkOf(spans)
	mail, isMail := link.Target.(Mail)
	return mail, isLink && isMail
}

func (Address) target() {}
func (Page) target()    {}
func (Mail) target()    {}
