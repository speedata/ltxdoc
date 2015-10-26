package ltxdoc

import (
	"fmt"
	"os"

	"github.com/speedata/ltxref"
)

var (
	latexref ltxref.Ltxref
)

func ShowHelpFor(cmdname, filename string) {
	var err error
	if filename != "" {
		latexref, err = ltxref.ReadXMLFile(filename)
	} else {
		latexref, err = ltxref.ReadXMLData(MustAsset("ltxref.xml"))
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	commands := latexref.FilterCommands(cmdname, "")
	for _, cmd := range commands {
		cmd.ToString(os.Stdout)
	}

	return
}
