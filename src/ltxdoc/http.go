package ltxdoc

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/speedata/ltxref"
)

var (
	tpl *template.Template
)

func dummy() {
	fmt.Println()
}

func escapeurl(part string) string {
	var Url *url.URL
	Url, err := url.Parse(part)
	if err != nil {
		fmt.Println(err)
	}
	return Url.String()
}

func StartHTTPD(httpaddress, filename string) {

	funcMap := template.FuncMap{
		"urlescape": func(in string) string {
			return escapeurl(in)
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

	maintemplate := string(MustAsset("templates/main.html"))
	detailtemplate := string(MustAsset("templates/commanddetail.html"))
	layouttemplate := string(MustAsset("templates/layout.html"))

	var err error
	tpl = template.Must(template.New("main.html").Funcs(funcMap).Parse(maintemplate))
	template.Must(tpl.Parse(detailtemplate)).Parse(layouttemplate)
	// a bug in go-bindata leads to the duplication of the path
	latexref, err = ltxref.ReadXMLData(MustAsset("ltxref.xml"))
	if err != nil {
		fmt.Println(err)
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/", mainHandler)
	r.HandleFunc("/cmd/{command}", commandDetailHandler)
	r.HandleFunc("/class/{documentclass}", documentclassDetailHandler)
	r.HandleFunc("/env/{environment}", environmentDetailHandler)
	r.HandleFunc("/pkg/{package}", packageDetailHandler)
	r.HandleFunc("/pkg/{package}/cmd/{command}", commandDetailHandler)
	r.HandleFunc("/tag/{tagname}", tagHandler)
	r.PathPrefix("/assets/").Handler(http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "httproot"}))
	http.Handle("/", r)
	fmt.Println("Listening on", httpaddress)
	http.ListenAndServe(httpaddress, nil)
}

func commandDetailHandler(w http.ResponseWriter, r *http.Request) {
	requestedCommand := mux.Vars(r)["command"]
	requestedPackage := mux.Vars(r)["package"]
	filtervalue := strings.ToLower(r.FormValue("filter"))
	var backlink string
	if requestedPackage == "" {
		backlink = "/"
	} else {
		backlink = "/pkg/" + escapeurl(requestedPackage)
	}
	cmd := latexref.GetCommandFromPackage(requestedCommand, requestedPackage)
	if cmd != nil {
		data := struct {
			Backlink string
			Filter   string
			Command  *ltxref.Command
		}{
			Backlink: backlink,
			Filter:   filtervalue,
			Command:  cmd,
		}
		err := tpl.ExecuteTemplate(w, "commanddetail.html", data)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	fmt.Println("Command not found")
	return
}

func documentclassDetailHandler(w http.ResponseWriter, r *http.Request) {
	requestedDocumentclass := mux.Vars(r)["documentclass"]
	filtervalue := strings.ToLower(r.FormValue("filter"))

	for _, env := range latexref.Documentclasses {
		if env.Name == requestedDocumentclass {
			data := struct {
				Filter        string
				Backlink      string
				Documentclass ltxref.Documentclass
			}{
				Filter:        filtervalue,
				Documentclass: env,
			}
			err := tpl.ExecuteTemplate(w, "classdetail", data)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}

}

func environmentDetailHandler(w http.ResponseWriter, r *http.Request) {
	requestedEnvironment := mux.Vars(r)["environment"]
	filtervalue := strings.ToLower(r.FormValue("filter"))

	for _, env := range latexref.Environments {
		if env.Name == requestedEnvironment {
			data := struct {
				Filter      string
				Environment ltxref.Environment
			}{
				Filter:      filtervalue,
				Environment: env,
			}
			err := tpl.ExecuteTemplate(w, "envdetail", data)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}
}

func packageDetailHandler(w http.ResponseWriter, r *http.Request) {
	requestedPackage := mux.Vars(r)["package"]
	filtervalue := strings.ToLower(r.FormValue("filter"))
	for _, pkg := range latexref.Packages {
		if pkg.Name == requestedPackage {
			data := struct {
				Filter  string
				Package ltxref.Package
			}{
				Filter:  filtervalue,
				Package: pkg,
			}
			err := tpl.ExecuteTemplate(w, "pkgdetail", data)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}
}

type mainstruct struct {
	Filter          string
	Commands        []ltxref.Command
	Environments    []ltxref.Environment
	Documentclasses []ltxref.Documentclass
	Packages        []ltxref.Package
	Tags            []string
}

// Show commands with the given tag only
func tagHandler(w http.ResponseWriter, r *http.Request) {
	tagname := mux.Vars(r)["tagname"]

	data := mainstruct{
		Commands:        latexref.GetCommandsWithTag(tagname),
		Environments:    latexref.GetEnvironmentsWithTag(tagname),
		Packages:        latexref.GetPackagesWithTag(tagname),
		Documentclasses: latexref.GetDocumentclassesWithTag(tagname),
		Tags:            latexref.Tags(),
	}

	err := tpl.ExecuteTemplate(w, "main.html", data)
	if err != nil {
		fmt.Println(err)
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	// empty string if no filter is given
	filterFormValue := strings.ToLower(r.FormValue("filter"))

	data := mainstruct{
		Filter:          filterFormValue,
		Commands:        latexref.FilterCommands(filterFormValue),
		Environments:    latexref.FilterEnvironments(filterFormValue),
		Documentclasses: latexref.FilterDocumentclasses(filterFormValue),
		Packages:        latexref.FilterPackages(filterFormValue),
		Tags:            latexref.Tags(),
	}

	err := tpl.ExecuteTemplate(w, "main.html", data)
	if err != nil {
		fmt.Println(err)
	}
}
