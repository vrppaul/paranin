package document

import (
	"testing"
)

// A mail's address parts at its last at into its mailbox and its host, and an address with no at stays whole.
func TestAMailPartsAtItsLastAt(test *testing.T) {
	cases := []struct {
		name    string
		address string
		box     string
		host    string
		parted  bool
	}{
		{name: "an address", address: "mira@example.test", box: "mira", host: "example.test", parted: true},
		{name: "a quoted mailbox holding an at", address: `"mira@home"@example.test`, box: `"mira@home"`, host: "example.test", parted: true},
		{name: "no at", address: "mira", box: "mira", host: "", parted: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			box, host, parted := Mail{Address: testCase.address}.Parts()
			if box != testCase.box || host != testCase.host || parted != testCase.parted {
				test.Errorf("Parts of %q = %q, %q, %v, want %q, %q, %v", testCase.address, box, host, parted, testCase.box, testCase.host, testCase.parted)
			}
		})
	}
}

// Spans that are one link to a mailbox are a mail; a link elsewhere, though its words hold an at, and words that are
// no link are none.
func TestOneLinkToAMailboxIsAMail(test *testing.T) {
	mail := Mail{Address: "mira@example.test"}
	cases := []struct {
		name   string
		spans  []Span
		isMail bool
	}{
		{name: "a link to a mailbox", spans: []Span{Link{Target: mail, Spans: []Span{Text{Text: "mira@example.test"}}}}, isMail: true},
		{
			name:   "a link elsewhere whose words hold an at",
			spans:  []Span{Link{Target: Address{URL: "https://social.test/@mira"}, Spans: []Span{Text{Text: "social.test/@mira"}}}},
			isMail: false,
		},
		{name: "words", spans: []Span{Text{Text: "mira@example.test"}}, isMail: false},
		{name: "a link to a mailbox and more words", spans: []Span{Link{Target: mail, Spans: nil}, Text{Text: " or ssh"}}, isMail: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if _, isMail := MailOf(testCase.spans); isMail != testCase.isMail {
				test.Errorf("MailOf(%+v) is a mail: %v, want %v", testCase.spans, isMail, testCase.isMail)
			}
		})
	}
}
