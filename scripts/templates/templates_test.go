package main

import (
	"os"
	"path/filepath"
	"testing"
)

// A template as templ formats it, the same template out of its format, and another as templ formats it
const (
	formattedTemplate   = "package views\n\ntempl greeting() {\n\t<p>hello</p>\n}\n"
	unformattedTemplate = "package views\n\ntempl greeting() {\n<p>hello</p>\n}\n"
	changedTemplate     = "package views\n\ntempl greeting() {\n\t<p>goodbye</p>\n}\n"
)

// The template the tests write, under the root's templates folder
const templatePath = "internal/views/greeting.templ"

// A template written passes the check with its Go; the same template put out of its format fails it, and so does
// one changed after its Go was written.
func TestTheCheckHoldsEachTemplateToItsFormatAndItsGo(test *testing.T) {
	cases := []struct {
		name     string
		template string
		want     int
	}{
		{name: "as it was written", template: formattedTemplate, want: succeeded},
		{name: "out of its format", template: unformattedTemplate, want: failed},
		{name: "changed after its Go", template: changedTemplate, want: failed},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			root := test.TempDir()
			writeTemplate(test, root, formattedTemplate)
			if code := run([]string{writeFlag}, root); code != succeeded {
				test.Fatalf("run(%q) = %d, want %d", writeFlag, code, succeeded)
			}
			writeTemplate(test, root, testCase.template)
			if got := run(nil, root); got != testCase.want {
				test.Errorf("run() over %q = %d, want %d", testCase.template, got, testCase.want)
			}
		})
	}
}

func writeTemplate(test *testing.T, root, template string) {
	path := filepath.Join(root, filepath.FromSlash(templatePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		test.Fatalf("make the folder of %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(template), 0o600); err != nil {
		test.Fatalf("write %s: %v", path, err)
	}
}
