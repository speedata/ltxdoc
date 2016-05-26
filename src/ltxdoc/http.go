package ltxdoc

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/speedata/ltxref"
)

var (
	tpl        *template.Template
	editMode   bool
	edittokens map[string]string
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
	file, err := os.Open("edittokens.txt")
	edittokens = make(map[string]string)
	if err != nil {
		allowEdit = false
	} else {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			kv := strings.Split(scanner.Text(), "=")
			if len(kv) == 2 {
				edittokens[kv[1]] = kv[0]
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
	}

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
				ret = "<tt>{...}</tt>"
			case ltxref.MANDLIST:
				ret = "<tt>{...,...,...}</tt>"
			case ltxref.OPTARG:
				ret = "<tt>[...]</tt>"
			case ltxref.OPTLIST:
				ret = "<tt>[...,...,...]</tt>"
			case ltxref.TODIMENORSPREADDIMEN:
				ret = "<tt>to</tt> <i>‹dimen›</i> or <tt>spread</tt> ‹<i>dimen</i>›"
			case ltxref.KEYVALLIST:
				ret = "<tt>[..=..,..=..,..=..]</tt>"
			case ltxref.PAREN:
				ret = "(...)"
			default:
				ret = "??"
			}
			return template.HTML(ret)
		},
		"buildurl": func(inlist ...string) string {
			u := &url.URL{}
			u.Path = inlist[0]
			for i := 1; i < len(inlist); i++ {
				if inlist[i+1] != "" {
					addKeyValueToUrl(u, inlist[i], inlist[i+1])
				}
				i++
			}
			return u.String()
		},
	}

	maintemplate := string(MustAsset("templates/main.html"))
	edittemplate := string(MustAsset("templates/edit.html"))
	detailtemplate := string(MustAsset("templates/commanddetail.html"))
	layouttemplate := string(MustAsset("templates/layout.html"))

	// var err error
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
	r.HandleFunc("/addenvironment", addEnvironmentHandler).Methods("POST")
	r.HandleFunc("/adddocumentclass", addDocumentClassHandler).Methods("POST")
	r.HandleFunc("/addpackage", addPackageHandler).Methods("POST")
	r.HandleFunc("/editcmd/{command}", editCommandHandler)
	r.HandleFunc("/editenv/{environment}", editEnvironmentHandler)
	r.HandleFunc("/editdc/{documentclass}", editDocumentClassHandler)
	r.HandleFunc("/editpackage/{package}", editPackageHandler)
	r.HandleFunc("/editpkgcmd/{package}/{command}", editCommandHandler)
	r.HandleFunc("/cmd/{command}", commandDetailHandler)
	r.HandleFunc("/class/{documentclass}", documentclassDetailHandler)
	r.HandleFunc("/env/{environment}", environmentDetailHandler)
	r.HandleFunc("/pkg/{package}", packageDetailHandler)
	r.HandleFunc("/pkg/{package}/cmd/{command}", commandDetailHandler)
	r.PathPrefix("/assets/").Handler(http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "httproot"}))
	http.Handle("/", r)
	fmt.Println("Listening on", httpaddress)
	fmt.Println(http.ListenAndServe(httpaddress, nil))
}

var mutex = &sync.Mutex{}

func saveXML() {
	data, err := latexref.ToXML()
	if err != nil {
		return
	}
	mutex.Lock()
	ioutil.WriteFile("savedata.xml", data, os.ModePerm)
	mutex.Unlock()
}

func addEnvironmentHandler(w http.ResponseWriter, r *http.Request) {
	if editToken(r) == "" {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}
	requestedEnvironment := r.FormValue("environment")
	_, err := latexref.AddEnvironment(requestedEnvironment)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		fmt.Println(err)
		return
	}
	backlink := &url.URL{}
	backlink.Path = "/editenv/" + requestedEnvironment
	addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))
	// Post/Redirect/Get doesn't work with temp redirect.
	http.Redirect(w, r, backlink.String(), http.StatusSeeOther)
	return
}

func addDocumentClassHandler(w http.ResponseWriter, r *http.Request) {
	if editToken(r) == "" {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}
	requestedDocumentClass := r.FormValue("documentclass")
	_, err := latexref.AddDocumentClass(requestedDocumentClass)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		fmt.Println(err)
		return
	}
	backlink := &url.URL{}
	backlink.Path = "/editdc/" + requestedDocumentClass
	addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))
	// Post/Redirect/Get doesn't work with temp redirect.
	http.Redirect(w, r, backlink.String(), http.StatusSeeOther)
	return
}

func addCommandHandler(w http.ResponseWriter, r *http.Request) {
	if editToken(r) == "" {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}
	requestedCommand := r.FormValue("command")
	requestedPackage := r.FormValue("package")
	_, err := latexref.AddCommand(requestedCommand, requestedPackage)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		fmt.Println(err)
		return
	}
	backlink := &url.URL{}
	if requestedPackage == "" {
		backlink.Path = "/editcmd/" + requestedCommand
	} else {
		backlink.Path = "/editpkgcmd/" + requestedPackage + "/" + requestedCommand
	}
	addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))
	// Post/Redirect/Get doesn't work with temp redirect.
	http.Redirect(w, r, backlink.String(), http.StatusSeeOther)
	return
}

func addPackageHandler(w http.ResponseWriter, r *http.Request) {
	if editToken(r) == "" {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}
	requestedPackage := r.FormValue("package")
	_, err := latexref.AddPackage(requestedPackage)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		fmt.Println(err)
		return
	}
	backlink := &url.URL{}
	backlink.Path = "/editpackage/" + requestedPackage
	addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))
	// Post/Redirect/Get doesn't work with temp redirect.
	http.Redirect(w, r, backlink.String(), http.StatusSeeOther)
	return
}

func editEnvironmentHandler(w http.ResponseWriter, r *http.Request) {
	requestedPackage := ""
	requestedEnvironment := r.FormValue("environment")
	if requestedEnvironment == "" {
		requestedEnvironment = mux.Vars(r)["environment"]
	}

	if editToken(r) == "" {
		http.Redirect(w, r, "/env/"+escapeurl(requestedEnvironment), http.StatusUnauthorized)
		return
	}

	var env *ltxref.Environment
	env = latexref.GetEnvironmentWithName(requestedEnvironment)
	if env == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)

	}
	switch r.Method {
	case "POST":
		env.ShortDescription["en"] = r.FormValue("shortdesc")
		env.Description["en"] = template.HTML(r.FormValue("description"))
		env.Label = strings.Split(r.FormValue("tags"), ",")
		env.Level = r.FormValue("level")
		env.Variant = nil
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
			env.Variant = append(env.Variant, *v)

		}
		saveXML()
		http.Redirect(w, r, "/env/"+escapeurl(requestedEnvironment)+"?edit="+editToken(r), http.StatusSeeOther)
		return

	case "GET":
		backlink := &url.URL{}
		if requestedPackage == "" {
			backlink.Path = "/env/" + requestedEnvironment
		} else {
			backlink.Path = "/pkg/" + requestedPackage + "/env/" + requestedEnvironment
		}
		addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))

		data := struct {
			Backlink     string
			Environment  *ltxref.Environment
			XMLUrl       string
			PlainTextUrl string
			Edit         string
			Tags         []string
		}{
			Backlink:     backlink.String(),
			Environment:  env,
			Edit:         editToken(r),
			XMLUrl:       addXMLFormatString(r.URL),
			PlainTextUrl: addTXTFormatString(r.URL),
			Tags:         latexref.Tags(),
		}
		err := tpl.ExecuteTemplate(w, "editenvironment", data)
		if err != nil {
			fmt.Println(err)
		}

	}
}

func editPackageHandler(w http.ResponseWriter, r *http.Request) {
	requestedPackage := r.FormValue("package")
	if requestedPackage == "" {
		requestedPackage = mux.Vars(r)["package"]
	}
	if editToken(r) == "" {
		http.Redirect(w, r, "/pkg/"+escapeurl(requestedPackage), http.StatusUnauthorized)
		return
	}
	var pkg *ltxref.Package
	pkg = latexref.GetPackageWithName(requestedPackage)
	switch r.Method {
	case "GET":
		backlink := &url.URL{}
		backlink.Path = "/pkg/" + requestedPackage
		addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))

		data := struct {
			Backlink     string
			Package      *ltxref.Package
			XMLUrl       string
			PlainTextUrl string
			Edit         string
			Tags         []string
		}{
			Backlink:     backlink.String(),
			Package:      pkg,
			Edit:         editToken(r),
			XMLUrl:       addXMLFormatString(r.URL),
			PlainTextUrl: addTXTFormatString(r.URL),
			Tags:         latexref.Tags(),
		}
		err := tpl.ExecuteTemplate(w, "editpackage", data)
		if err != nil {
			fmt.Println(err)
		}
	case "POST":
		pkg.Options = make([]*ltxref.Packageoption, 0)
		pkg.ShortDescription["en"] = r.FormValue("shortdesc")
		pkg.Description["en"] = template.HTML(r.FormValue("description"))
		pkg.Label = strings.Split(r.FormValue("tags"), ",")
		pkg.Level = r.FormValue("level")
		maxpackageoptions, err := strconv.Atoi(r.FormValue("panelcount"))
		if err != nil {
			return
		}
		for i := 1; i <= maxpackageoptions; i++ {
			po := ltxref.NewPackageOption()
			po.ShortDescription["en"] = r.FormValue(fmt.Sprintf("packageoption%dshortdescription", i))
			po.Default = r.FormValue(fmt.Sprintf("packageoption%ddefault", i)) == "on"
			po.Name = r.FormValue(fmt.Sprintf("packageoption%dname", i))
			pkg.Options = append(pkg.Options, po)
		}
		saveXML()
		http.Redirect(w, r, "/pkg/"+escapeurl(requestedPackage)+"?edit="+editToken(r), http.StatusSeeOther)
		return

	}
}

func editDocumentClassHandler(w http.ResponseWriter, r *http.Request) {
	requestedDocumentClass := r.FormValue("documentclass")
	if requestedDocumentClass == "" {
		requestedDocumentClass = mux.Vars(r)["documentclass"]
	}

	if editToken(r) == "" {
		http.Redirect(w, r, "/class/"+escapeurl(requestedDocumentClass), http.StatusUnauthorized)
		return
	}

	var dc *ltxref.DocumentClass
	dc = latexref.GetDocumentClass(requestedDocumentClass)
	if dc == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	switch r.Method {
	case "POST":
		dc.Optiongroup = nil
		dc.ShortDescription["en"] = r.FormValue("shortdesc")
		dc.Description["en"] = template.HTML(r.FormValue("description"))
		dc.Label = strings.Split(r.FormValue("tags"), ",")
		dc.Level = r.FormValue("level")

		maxog, err := strconv.Atoi(r.FormValue("panelcount"))
		if err != nil {
			return
		}

		// starting with 1
		for i := 1; i <= maxog; i++ {
			og := ltxref.NewOptionGroup()
			og.ShortDescription["en"] = r.FormValue(fmt.Sprintf("optiongroup%dshortdescription", i))

			argcountname := fmt.Sprintf("panel%dargumentcount", i)
			argc, err := strconv.Atoi(r.FormValue(argcountname))
			if err == nil {
				for j := 1; j <= argc; j++ {
					co := ltxref.NewClassOption()
					co.Default = r.FormValue(fmt.Sprintf("og%doptional%d", i, j)) == "on"
					co.Name = r.FormValue(fmt.Sprintf("og%doptionname%d", i, j))
					co.ShortDescription["en"] = r.FormValue(fmt.Sprintf("og%dshortdesc%d", i, j))
					if co.Name != "" {
						og.Classoption = append(og.Classoption, co)
					}
				}
				dc.Optiongroup = append(dc.Optiongroup, og)
			}
		}

		saveXML()
		http.Redirect(w, r, "/class/"+escapeurl(requestedDocumentClass)+"?edit="+editToken(r), http.StatusSeeOther)
		return

	case "GET":
		backlink := &url.URL{}
		backlink.Path = "/class/" + requestedDocumentClass
		addKeyValueToUrl(backlink, "edit", r.FormValue("edit"))

		data := struct {
			Backlink      string
			DocumentClass *ltxref.DocumentClass
			XMLUrl        string
			PlainTextUrl  string
			Edit          string
			Tags          []string
		}{
			Backlink:      backlink.String(),
			DocumentClass: dc,
			Edit:          editToken(r),
			XMLUrl:        addXMLFormatString(r.URL),
			PlainTextUrl:  addTXTFormatString(r.URL),
			Tags:          latexref.Tags(),
		}
		err := tpl.ExecuteTemplate(w, "editdocumentclass", data)
		if err != nil {
			fmt.Println(err)
		}

	}
}

func editCommandHandler(w http.ResponseWriter, r *http.Request) {
	requestedPackage := r.FormValue("package")
	if requestedPackage == "" {
		requestedPackage = mux.Vars(r)["package"]
	}
	requestedCommand := r.FormValue("command")
	if requestedCommand == "" {
		requestedCommand = mux.Vars(r)["command"]
	}

	if editToken(r) == "" {
		http.Redirect(w, r, "/cmd/"+escapeurl(requestedCommand), http.StatusUnauthorized)
		return
	}

	var cmd *ltxref.Command
	cmd = latexref.GetCommandFromPackage(requestedCommand, requestedPackage)
	if cmd == nil {
		fmt.Println("Command not found")
		return
	}

	switch r.Method {
	case "POST":
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
		var path string
		if requestedPackage != "" {
			path = "/pkg/" + escapeurl(requestedPackage) + "/cmd/" + escapeurl(requestedCommand)
		} else {
			path = "/cmd/" + escapeurl(requestedCommand)
		}
		saveXML()
		http.Redirect(w, r, path+"?edit="+editToken(r), http.StatusSeeOther)
		return
	case "GET":
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
			Package      string
			Edit         string
			Tags         []string
		}{
			Backlink:     backlink.String(),
			Command:      cmd,
			Edit:         editToken(r),
			XMLUrl:       addXMLFormatString(r.URL),
			Package:      requestedPackage,
			PlainTextUrl: addTXTFormatString(r.URL),
			Tags:         latexref.Tags(),
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
		sendXML(w, str)
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
				Package      string
			}{
				Backlink:     backlink.String(),
				Edit:         editToken(r),
				Filter:       filtervalue,
				Command:      cmd,
				XMLUrl:       addXMLFormatString(r.URL),
				PlainTextUrl: addTXTFormatString(r.URL),
				Package:      requestedPackage,
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

	class := latexref.GetDocumentClass(requestedDocumentclass)
	if class == nil {
		// not found -> error // TODO
		return
	}
	switch strings.ToLower(r.FormValue("format")) {
	case "xml":
		l := ltxref.Ltxref{Version: latexref.Version}
		l.DocumentClasses = append(l.DocumentClasses, class)
		str, err := l.ToXML()
		if err != nil {
			fmt.Println(err)
			return
		}
		sendXML(w, str)
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
			DocumentClass ltxref.DocumentClass
			XMLUrl        string
			PlainTextUrl  string
		}{
			Backlink:      backlink.String(),
			Edit:          editToken(r),
			Filter:        filtervalue,
			DocumentClass: *class,
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
		l.Environments = append(l.Environments, env)
		str, err := l.ToXML()
		if err != nil {
			fmt.Println(err)
			return
		}
		sendXML(w, str)
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

// Write the data as application/xml to w
func sendXML(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-type", "application/xml")
	fmt.Fprint(w, string(data))
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
		l.Packages = append(l.Packages, pkg)
		str, err := l.ToXML()
		if err != nil {
			fmt.Println(err)
			return
		}
		sendXML(w, str)
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
			Edit         string
			Filter       string
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
	expert := r.FormValue("expert") == "on"

	tag := strings.ToLower(r.FormValue("tag"))

	l := ltxref.Ltxref{Version: latexref.Version}
	l.Commands = latexref.FilterCommands(filter, tag, expert)
	l.Packages = latexref.FilterPackages(filter, tag)
	l.Environments = latexref.FilterEnvironments(filter, tag, expert)
	l.DocumentClasses = latexref.FilterDocumentClasses(filter, tag, expert)

	switch strings.ToLower(r.FormValue("format")) {
	case "xml":
		str, err := l.ToXML()
		if err != nil {
			fmt.Println(err)
		}
		sendXML(w, str)
		return
	case "txt":
		l.ToString(w, true)
		return
	}

	data := struct {
		Filter       string
		Tag          string
		Edit         string
		Expert       bool
		L            ltxref.Ltxref
		Tags         []string
		XMLUrl       string
		PlainTextUrl string
	}{
		L:            l,
		Filter:       filter,
		Tag:          tag,
		Edit:         editToken(r),
		Expert:       expert,
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
	x := r.FormValue("edit")
	if edittokens[x] == "" {
		return ""
	} else {
		return x
	}
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
