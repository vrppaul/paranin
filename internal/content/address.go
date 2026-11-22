package content

import (
	"net/mail"
	"net/url"
	"strings"

	"go.yaml.in/yaml/v3"
)

// The schemes an address of the profile may have
const (
	secureScheme = "https"
	plainScheme  = "http"
)

// WebAddress is an absolute web address as the content writes it, read once: what is written, and its host where
// it reads as an address of the web, or nothing where it does not, which the check names.
type WebAddress struct {
	written string
	host    string
	// secure is whether its scheme is https, which a check of an address that must be one names
	secure bool
}

// UnmarshalYAML reads a web address, and never fails on its text: what is no address is a fault for the check to
// name.
func (address *WebAddress) UnmarshalYAML(node *yaml.Node) error {
	*address = readWebAddress(node.Value)
	return nil
}

func readWebAddress(written string) WebAddress {
	parsed, err := url.Parse(written)
	if err != nil || (parsed.Scheme != secureScheme && parsed.Scheme != plainScheme) {
		return WebAddress{written: written, host: "", secure: false}
	}
	return WebAddress{written: written, host: parsed.Host, secure: parsed.Scheme == secureScheme}
}

// Written is the address as the content writes it.
func (address WebAddress) Written() string {
	return address.written
}

// Host is the address's host, which names the site it leads to; an address that reads as none has no host.
func (address WebAddress) Host() string {
	return address.host
}

// absolute is whether the address reads as an absolute web address: one of the web's schemes, and a host.
func (address WebAddress) absolute() bool {
	return address.host != ""
}

// The scheme of a post's link that writes to a mailbox
const mailScheme = "mailto:"

// Mailbox is the mail address a post's link writes to, where its address is written under the scheme that writes to
// one and holds a bare mail address.
func Mailbox(address string) (string, bool) {
	box, isMail := strings.CutPrefix(address, mailScheme)
	return box, isMail && bareMail(box)
}

// bareMail is whether a text is a mail address alone, with no name and no brackets round it.
func bareMail(address string) bool {
	parsed, err := mail.ParseAddress(address)
	return err == nil && parsed.Address == address
}
