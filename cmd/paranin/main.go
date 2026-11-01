// Command paranin is the composition root of the site: it builds every part and hands each what it needs.
// For now it has one command, check, which loads a content folder as the site would and lists every fault in
// it, so the content's hook and a person editing it refuse what the server would refuse.
//
//	paranin check content
//
// It prints the faults, one a line, and exits 0 when the content loads, 1 when it has faults, and 2 when the
// command itself is written wrong.
package main

import (
	"fmt"
	"os"

	"github.com/vrppaul/paranin/internal/content"
)

// What the program exits with
const (
	contentLoads  = 0
	contentFaulty = 1
	usageMistaken = 2
)

// How the command is written: the program, the command, the folder
const (
	checkCommand    = "check"
	usage           = "usage: paranin check <content folder>"
	commandArgument = 1
	folderArgument  = 2
	argumentCount   = 3
)

func main() {
	os.Exit(run(os.Args))
}

func run(arguments []string) int {
	if len(arguments) != argumentCount || arguments[commandArgument] != checkCommand {
		fmt.Fprintln(os.Stderr, usage)
		return usageMistaken
	}
	folder := arguments[folderArgument]
	if _, err := content.Load(os.DirFS(folder)); err != nil {
		fmt.Fprintf(os.Stderr, "%s:\n%v\n", folder, err)
		return contentFaulty
	}
	return contentLoads
}
