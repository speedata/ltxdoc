package ltxref

import (
	"fmt"
	ht "html/template"
	"io"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/commonsense-org/html2text"
)

var (
	tpl *template.Template
)

func tfunderline(cmd string, level int) string {
	var char string
	switch level {
	case 1:
		char = "="
	case 2:
		char = "-"
	case 3:
		char = "·"
	}
	return cmd + "\n" + strings.Repeat(char, len(cmd)) + "\n"
}
func tfshowargument(in Argumenttype) string {
	var ret string
	switch in {
	case OPTARG:
		ret = ("[...]")
	case OPTLIST:
		ret = ("[...,...,...]")
	case MANDARG:
		ret = ("{...}")
	case MANDLIST:
		ret = ("{...,...,...}")
	case TODIMENORSPREADDIMEN:
		ret = ("to ‹dimen› [or] spread ‹dimen›")
	default:
		ret = "??"
	}
	return ret
}

func tfspace(cmd string) string {
	l := utf8.RuneCountInString(cmd)
	return strings.Repeat(" ", l)
}

func tfenvspace(cmd string) string {
	l := utf8.RuneCountInString(cmd) + 8
	return strings.Repeat(" ", l)
}

func tfplaceholder(in Argumenttype, count int, optional bool) string {
	l := utf8.RuneCountInString(tfshowargument(in))
	count += 1
	var num string
	if optional {
		num = fmt.Sprintf("(%d)", count)
	} else {
		num = fmt.Sprintf("%d", count)
	}
	l -= utf8.RuneCountInString(num)
	var left, right int

	if l%2 == 0 {
		left = l / 2
		right = l / 2
	} else {
		left = (l + 1) / 2
		right = (l - 1) / 2
	}
	r := strings.Repeat(" ", left) + num + strings.Repeat(" ", right)
	return r
}

func tfshowdescription(in ht.HTML) string {
	str, err := html2text.FromString(string(in), "", 0)
	if err != nil {
		fmt.Println(err)
	}
	return str
}

func init() {
	funcMap := map[string]interface{}{
		"underline":       tfunderline,
		"showargument":    tfshowargument,
		"space":           tfspace,
		"envspace":        tfenvspace,
		"placehoder":      tfplaceholder,
		"showdescription": tfshowdescription,
	}

	maintemplate := string(MustAsset("templates/main.txt"))
	detailtemplate := string(MustAsset("templates/details.txt"))

	tpl = template.Must(template.New("main.txt").Funcs(funcMap).Parse(maintemplate))
	template.Must(tpl.Parse(detailtemplate))

}

func (c *Command) ToString(w io.Writer) {
	data := struct {
		Command *Command
	}{
		Command: c,
	}
	err := tpl.ExecuteTemplate(w, "cmddetail", data)
	if err != nil {
		fmt.Println(err)
	}
}

func (p *Package) ToString(w io.Writer) {
	data := struct {
		Pkg *Package
	}{
		Pkg: p,
	}
	err := tpl.ExecuteTemplate(w, "pkgdetail", data)
	if err != nil {
		fmt.Println(err)
	}
}
func (e *Environment) ToString(w io.Writer) {
	data := struct {
		Environment *Environment
	}{
		Environment: e,
	}
	err := tpl.ExecuteTemplate(w, "envdetail", data)
	if err != nil {
		fmt.Println(err)
	}
}
func (c *DocumentClass) ToString(w io.Writer) {
	data := struct {
		Class *DocumentClass
	}{
		Class: c,
	}
	err := tpl.ExecuteTemplate(w, "classdetail", data)
	if err != nil {
		fmt.Println(err)
	}
}

func (l *Ltxref) ToString(w io.Writer, short bool) {
	data := struct {
		L *Ltxref
	}{
		L: l,
	}
	err := tpl.ExecuteTemplate(w, "main", data)
	if err != nil {
		fmt.Println(err)
	}

}
