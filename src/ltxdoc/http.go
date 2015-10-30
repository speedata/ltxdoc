package ltxdoc

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/speedata/ltxref"
)

var (
	tpl      *template.Template
	editMode bool
)

func escapeurl(part string) string {
	var Url *url.URL
	Url, err := url.Parse(part)
	if err != nil {
		fmt.Println(err)
	}
	return Url.String()
}

func StartHTTPD(httpaddress, filename string, allowEdit bool) {
	editMode = allowEdit

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
			case ltxref.MANDARG:
				ret = ("<tt>{...}</tt>")
			case ltxref.MANDLIST:
				ret = ("<tt>{...,...,...}</tt>")
			case ltxref.OPTARG:
				ret = ("<tt>[...]</tt>")
			case ltxref.OPTLIST:
				ret = ("<tt>[...,...,...]</tt>")
			case ltxref.TODIMENORSPREADDIMEN:
				ret = ("<tt>to</tt> <i>‹dimen›</i> or <tt>spread</tt> ‹<i>dimen</i>›")
			default:
				ret = "??"
			}
			return template.HTML(ret)
		},
	}

	maintemplate := string(MustAsset("templates/main.html"))
	edittemplate := string(MustAsset("templates/edit.html"))
	detailtemplate := string(MustAsset("templates/commanddetail.html"))
	layouttemplate := string(MustAsset("templates/layout.html"))

	var err error
	tpl = template.Must(template.New("main.html").Funcs(funcMap).Parse(maintemplate))
	template.Must(tpl.Parse(detailtemplate))
	template.Must(tpl.Parse(layouttemplate))
	template.Must(tpl.Parse(edittemplate))

	if filename != "" {
		latexref, err = ltxref.ReadXMLFile(filename)
	} else {
		latexref, err = ltxref.ReadXMLData(MustAsset("ltxref.xml"))
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/", mainHandler)
	r.HandleFunc("/editcmd", editCommandHandler)
	r.HandleFunc("/addcommand", addCommandHandler).Methods("POST")
	r.HandleFunc("/editcmd/{command}", editCommandHandler)
	r.HandleFunc("/cmd/{command}", commandDetailHandler)
	r.HandleFunc("/class/{documentclass}", documentclassDetailHandler)
	r.HandleFunc("/env/{environment}", environmentDetailHandler)
	r.HandleFunc("/pkg/{package}", packageDetailHandler)
	r.HandleFunc("/pkg/{package}/cmd/{command}", commandDetailHandler)
	r.PathPrefix("/assets/").Handler(http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "httproot"}))
	http.Handle("/", r)
	fmt.Println("Listening on", httpaddress)
	http.ListenAndServe(httpaddress, nil)
}

func addCommandHandler(w http.ResponseWriter, r *http.Request) {
	if editToken(r) == "" {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}
	requestedCommand := r.FormValue("command")

	_, err := latexref.AddCommand(requestedCommand)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		fmt.Println(err)
		return
	}
	backlink := &url.URL{}
	backlink.Path = "/editcmd/" + requestedCommand
	addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))
	// Post/Redirect/Get doesn't work with temp redirect.
	http.Redirect(w, r, backlink.String(), http.StatusSeeOther)
	return
}

func editCommandHandler(w http.ResponseWriter, r *http.Request) {
	requestedPackage := ""
	requestedCommand := r.FormValue("command")
	if requestedCommand == "" {
		requestedCommand = mux.Vars(r)["command"]
	}

	var cmd *ltxref.Command

	if editToken(r) == "" {
		http.Redirect(w, r, "/cmd/"+escapeurl(requestedCommand), http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "POST":
		cmd = latexref.GetCommandFromPackage(requestedCommand, "")
		cmd.ShortDescription["en"] = r.FormValue("shortdesc")
		cmd.Description["en"] = template.HTML(r.FormValue("description"))
		cmd.Label = strings.Split(r.FormValue("tags"), ",")
		cmd.Level = r.FormValue("level")
		cmd.Variant = nil
		variantcount, err := strconv.Atoi(r.FormValue("maxvarpanelcount"))
		if err != nil {
			fmt.Println(err)
			return
		}
		for i := 1; i <= variantcount; i++ {
			v := ltxref.NewVariant()
			v.Name = r.FormValue(fmt.Sprintf("name%d", i))
			v.Description["en"] = template.HTML(r.FormValue(fmt.Sprintf("variant%d", i)))
			numarguments, err := strconv.Atoi(r.FormValue(fmt.Sprintf("argcount%d", i)))
			if err != nil {
				fmt.Println(err)
				return
			}
			for arg := 1; arg <= numarguments; arg++ {
				a := ltxref.NewArgument()
				a.Name = r.FormValue(fmt.Sprintf("v%da%dname", i, arg))
				a.Optional = r.FormValue(fmt.Sprintf("v%da%doptional", i, arg)) == "on"

				tmp, err := strconv.Atoi(r.FormValue(fmt.Sprintf("v%da%dtype", i, arg)))
				if err != nil {
					fmt.Println(err)
					return
				}
				a.Type = ltxref.Argumenttype(tmp)
				v.Arguments = append(v.Arguments, a)
			}
			cmd.Variant = append(cmd.Variant, *v)
		}

		http.Redirect(w, r, "/cmd/"+escapeurl(requestedCommand)+"?edit="+editToken(r), http.StatusSeeOther)
		return

	case "GET":
		cmd = latexref.GetCommandFromPackage(requestedCommand, "")
		if cmd == nil {
			fmt.Println("Command not found")
			return
		}

		backlink := &url.URL{}
		if requestedPackage == "" {
			backlink.Path = "/cmd/" + requestedCommand
		} else {
			backlink.Path = "/pkg/" + requestedPackage + "/cmd/" + requestedCommand
		}
		addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))

		data := struct {
			Backlink     string
			Command      *ltxref.Command
			XMLUrl       string
			PlainTextUrl string
			Edit         string
		}{
			Backlink:     backlink.String(),
			Command:      cmd,
			Edit:         editToken(r),
			XMLUrl:       addXMLFormatString(r.URL),
			PlainTextUrl: addTXTFormatString(r.URL),
		}
		err := tpl.ExecuteTemplate(w, "editcommand", data)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func commandDetailHandler(w http.ResponseWriter, r *http.Request) {
	requestedCommand := mux.Vars(r)["command"]
	requestedPackage := mux.Vars(r)["package"]
	filtervalue := strings.ToLower(r.FormValue("filter"))

	cmd := latexref.GetCommandFromPackage(requestedCommand, requestedPackage)

	switch strings.ToLower(r.FormValue("format")) {
	case "xml":
		l := ltxref.Ltxref{Version: latexref.Version}
		l.Commands = append(l.Commands, cmd)
		str, err := l.ToXML()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprint(w, string(str))
		return
	case "txt":
		cmd.ToString(w)
		return
	default:
		backlink := &url.URL{}
		if requestedPackage == "" {
			backlink.Path = "/"
		} else {
			backlink.Path = "/pkg/" + requestedPackage
		}
		addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))
		addKeyValueToUrl(backlink, "filter", r.FormValue("filter"))

		if cmd != nil {
			data := struct {
				Command      *ltxref.Command
				Backlink     string
				Edit         string
				Filter       string
				XMLUrl       string
				PlainTextUrl string
			}{
				Backlink:     backlink.String(),
				Edit:         editToken(r),
				Filter:       filtervalue,
				Command:      cmd,
				XMLUrl:       addXMLFormatString(r.URL),
				PlainTextUrl: addTXTFormatString(r.URL),
			}
			err := tpl.ExecuteTemplate(w, "commanddetail.html", data)
			if err != nil {
				fmt.Println(err)
			}
			return

		}

	}
	fmt.Println("Command not found")
	return
}

func documentclassDetailHandler(w http.ResponseWriter, r *http.Request) {
	requestedDocumentclass := mux.Vars(r)["documentclass"]
	filtervalue := strings.ToLower(r.FormValue("filter"))

	class := latexref.GetDocumentclass(requestedDocumentclass)
	if class == nil {
		// not found -> error // TODO
		return
	}
	switch strings.ToLower(r.FormValue("format")) {
	case "xml":
		l := ltxref.Ltxref{Version: latexref.Version}
		l.Documentclasses = append(l.Documentclasses, *class)
		str, err := l.ToXML()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprint(w, string(str))
		return
	case "txt":
		class.ToString(w)
		return
	default:
		backlink := &url.URL{}
		backlink.Path = "/"
		addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))
		addKeyValueToUrl(backlink, "filter", r.FormValue("filter"))

		data := struct {
			Backlink      string
			Filter        string
			Edit          string
			Documentclass ltxref.Documentclass
			XMLUrl        string
			PlainTextUrl  string
		}{
			Backlink:      backlink.String(),
			Edit:          editToken(r),
			Filter:        filtervalue,
			Documentclass: *class,
			XMLUrl:        addXMLFormatString(r.URL),
			PlainTextUrl:  addTXTFormatString(r.URL),
		}
		err := tpl.ExecuteTemplate(w, "classdetail", data)
		if err != nil {
			fmt.Println(err)
		}
	}
	return
}

func environmentDetailHandler(w http.ResponseWriter, r *http.Request) {
	requestedEnvironment := mux.Vars(r)["environment"]
	filtervalue := strings.ToLower(r.FormValue("filter"))

	env := latexref.GetEnvironmentWithName(requestedEnvironment)
	if env == nil {
		// not found -> error // TODO
		return
	}

	switch strings.ToLower(r.FormValue("format")) {
	case "xml":
		l := ltxref.Ltxref{Version: latexref.Version}
		l.Environments = append(l.Environments, *env)
		str, err := l.ToXML()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprint(w, string(str))
		return
	case "txt":
		env.ToString(w)
	default:
		backlink := &url.URL{}
		backlink.Path = "/"
		addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))
		addKeyValueToUrl(backlink, "filter", r.FormValue("filter"))

		data := struct {
			Backlink     string
			Filter       string
			Edit         string
			Environment  ltxref.Environment
			XMLUrl       string
			PlainTextUrl string
		}{
			Backlink:     backlink.String(),
			Edit:         editToken(r),
			Filter:       filtervalue,
			Environment:  *env,
			XMLUrl:       addXMLFormatString(r.URL),
			PlainTextUrl: addTXTFormatString(r.URL),
		}
		err := tpl.ExecuteTemplate(w, "envdetail", data)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
}

func packageDetailHandler(w http.ResponseWriter, r *http.Request) {
	requestedPackage := mux.Vars(r)["package"]
	filtervalue := strings.ToLower(r.FormValue("filter"))

	pkg := latexref.GetPackageWithName(requestedPackage)
	if pkg == nil {
		// not found -> error // TODO
		return
	}

	switch strings.ToLower(r.FormValue("format")) {
	case "xml":
		l := ltxref.Ltxref{Version: latexref.Version}
		l.Packages = append(l.Packages, *pkg)
		str, err := l.ToXML()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprint(w, string(str))
		return
	case "txt":
		pkg.ToString(w)
	default:
		backlink := &url.URL{}
		backlink.Path = "/"
		addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))
		addKeyValueToUrl(backlink, "filter", r.FormValue("filter"))

		data := struct {
			Backlink     string
			Filter       string
			Edit         string
			Package      ltxref.Package
			XMLUrl       string
			PlainTextUrl string
		}{
			Backlink:     backlink.String(),
			Edit:         editToken(r),
			Filter:       filtervalue,
			Package:      *pkg,
			XMLUrl:       addXMLFormatString(r.URL),
			PlainTextUrl: addTXTFormatString(r.URL),
		}
		err := tpl.ExecuteTemplate(w, "pkgdetail", data)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {

	filter := strings.ToLower(r.FormValue("filter"))
	tag := strings.ToLower(r.FormValue("tag"))

	l := ltxref.Ltxref{Version: latexref.Version}
	l.Commands = latexref.FilterCommands(filter, tag)
	l.Packages = latexref.FilterPackages(filter, tag)
	l.Environments = latexref.FilterEnvironments(filter, tag)
	l.Documentclasses = latexref.FilterDocumentclasses(filter, tag)

	switch strings.ToLower(r.FormValue("format")) {
	case "xml":
		str, err := l.ToXML()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Fprint(w, string(str))
		return
	case "txt":
		l.ToString(w, true)
		return
	}

	data := struct {
		Filter       string
		Tag          string
		Edit         string
		L            ltxref.Ltxref
		Tags         []string
		XMLUrl       string
		PlainTextUrl string
	}{
		L:            l,
		Filter:       filter,
		Tag:          tag,
		Edit:         editToken(r),
		Tags:         latexref.Tags(),
		XMLUrl:       addXMLFormatString(r.URL),
		PlainTextUrl: addTXTFormatString(r.URL),
	}

	err := tpl.ExecuteTemplate(w, "main.html", data)
	if err != nil {
		fmt.Println(err)
	}
}

func editToken(r *http.Request) string {
	if !editMode {
		return ""
	}
	// Will be replaced by a more sane auth mode
	return r.FormValue("edit")
}

// Add ?format=xml to the given URL
func addXMLFormatString(u *url.URL) string {
	ret, err := u.Parse("")
	if err != nil {
		fmt.Println(err)
		return u.String()
	}
	val := ret.Query()
	val.Add("format", "xml")
	ret.RawQuery = val.Encode()
	return ret.String()
}

// Add ?format=txt to the given URL
func addTXTFormatString(u *url.URL) string {
	ret, err := u.Parse("")
	if err != nil {
		fmt.Println(err)
		return u.String()
	}
	val := ret.Query()
	val.Add("format", "txt")
	ret.RawQuery = val.Encode()
	return ret.String()
}

// Add ?key=value to the given URL
func addKeyValueToUrl(u *url.URL, key, value string) {
	if value == "" {
		return
	}
	val := u.Query()
	val.Del(key)
	val.Add(key, value)
	u.RawQuery = val.Encode()
}
