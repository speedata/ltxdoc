package ltxdoc

import (
	"fmt"

	"github.com/speedata/ltxref"
)

var (
	latexref ltxref.Ltxref
)

func ShowHelpFor(cmdname, filename string) {
	var err error
	latexref, err = ltxref.ReadXMLFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	commands := latexref.FilterCommands(cmdname, "")
	for _, cmd := range commands {
		fmt.Println("Documentation for command", cmd.Name)
	}

	return
}
