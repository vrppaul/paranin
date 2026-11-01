package content

import (
	"fmt"
	"strings"
)

// Fault is one thing wrong with the content: in a file, at a field of it where the fault has one.
type Fault struct {
	File    string
	Field   string
	Problem string
}

// String writes a fault as a person reads it: the file, the field, what is wrong.
func (fault Fault) String() string {
	if fault.Field == "" {
		return fmt.Sprintf("%s: %s", fault.File, fault.Problem)
	}
	return fmt.Sprintf("%s: %s %s", fault.File, fault.Field, fault.Problem)
}

// Faults are every fault a load found. As an error it lists them, one a line.
type Faults []Fault

func (faults Faults) Error() string {
	lines := make([]string, 0, len(faults))
	for _, fault := range faults {
		lines = append(lines, fault.String())
	}
	return strings.Join(lines, "\n")
}
