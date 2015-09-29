package main

import (
	"flag"
	"fmt"
	"os"

	"ltxdoc"
)

func main() {
	var filename, http string
	var edit bool
	flag.StringVar(&http, "http", "", "HTTP service address (e.g., ':6090')")
	flag.StringVar(&filename, "xmlfile", "", "Path to the XML file")
	flag.BoolVar(&edit, "edit", false, "Allow edit")
	flag.Parse()

	if edit {
		fmt.Println("Edit mode")
	}
	if http == "" {
		cmdname := flag.Arg(0)
		if cmdname == "" {
			fmt.Println("Please give exactly one command name to get help for.")
			os.Exit(0)
		}
		ltxdoc.ShowHelpFor(cmdname, filename)
	} else {
		ltxdoc.StartHTTPD(http, filename, edit)
	}

}
