package ltxdoc

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/speedata/ltxref"
)

var (
	tpl *template.Template
)

func dummy() {
	fmt.Println()
}

func StartHTTPD(httpaddress, filename string) {

	funcMap := template.FuncMap{
		"urlescape": func(in string) string {
			var Url *url.URL
			Url, err := url.Parse(in)
			if err != nil {
				fmt.Println(err)
			}
			return Url.String()
		},
		"addone": func(in int) int {
			return in + 1
		},
		"showargument": func(in ltxref.Argumenttype) template.HTML {
			var ret string
			switch in {
			case ltxref.OPTARG:
				ret = ("<tt>[...]</tt>")
			case ltxref.OPTLIST:
				ret = ("<tt>[...,...,...]</tt>")
			case ltxref.MANDARG:
				ret = ("<tt>{...}</tt>")
			case ltxref.MANDLIST:
				ret = ("<tt>{...,...,...}</tt>")
			case ltxref.TODIMENORSPREADDIMEN:
				ret = ("<tt>to</tt> <i>‹dimen›</i> or <tt>spread</tt> ‹<i>dimen</i>›")
			default:
				ret = "??"
			}
			return template.HTML(ret)
		},
	}
	tpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*html"))
	var err error
	latexref, err = ltxref.ReadXMLFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/", mainHandler)
	r.HandleFunc("/cmd/{command}", commandDetailHandler)
	r.HandleFunc("/tag/{tagname}", tagHandler)
	r.PathPrefix("/assets/").Handler(http.FileServer(http.Dir("httproot")))
	http.Handle("/", r)
	fmt.Println("Listening on", httpaddress)
	http.ListenAndServe(httpaddress, nil)
}

func commandDetailHandler(w http.ResponseWriter, r *http.Request) {
	requestedCommand := mux.Vars(r)["command"]
	filtervalue := r.FormValue("filter")

	for _, command := range latexref.Commands {
		if command.Name == requestedCommand {
			data := struct {
				Filter  string
				Command ltxref.Command
			}{
				Filter:  filtervalue,
				Command: command,
			}
			err := tpl.ExecuteTemplate(w, "commanddetail.html", data)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}

}

type mainstruct struct {
	Filter   string
	Commands []ltxref.Command
	Tags     []string
}

// Show commands with the given tag only
func tagHandler(w http.ResponseWriter, r *http.Request) {
	tagname := mux.Vars(r)["tagname"]

	data := mainstruct{
		Commands: latexref.GetCommandsWithTag(tagname),
		Tags:     latexref.Tags(),
	}

	err := tpl.ExecuteTemplate(w, "main.html", data)
	if err != nil {
		fmt.Println(err)
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	// empty string if no filter is given
	filterFormValue := r.FormValue("filter")

	data := mainstruct{
		Filter:   filterFormValue,
		Commands: latexref.FilterCommands(filterFormValue),
		Tags:     latexref.Tags(),
	}

	err := tpl.ExecuteTemplate(w, "main.html", data)
	if err != nil {
		fmt.Println(err)
	}
}
