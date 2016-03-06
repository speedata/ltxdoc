package html2text

// Fork of https://github.com/jaytaylor/html2text
// This does additional processing to omit non-content areas and rewrite;
//    a: tags to markdown style links
//    HX: headers to markdown style with prefixed 'X' number of "#" characters
//    li: tags to markdown style unordered lists (even for ol li's)
// Omission from extraction is done via a combinations of;
//   - tag selection
//   - class selection
//   - aria roles
import (
	"bytes"
	"golang.org/x/net/html"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	OmitClasses = 1 << iota
	OmitIds
	OmitRoles
)

type state struct {
	node          *html.Node
	buf           *bytes.Buffer
	url           string
	omitClasses   bool
	omitIds       bool
	omitRoles     bool
	indent        int
	gatherHeaders bool
	column        int
	headers       []string
}

// Indent spacing (used for nested li's)
const spacing = 4

var (
	spacingRe      = regexp.MustCompile(`[[:space:]]+`)
	newlineRe      = regexp.MustCompile("\n\n+")
	newlineSpaceRe = regexp.MustCompile("\n[ \r\n\t]+\n")
)

// https://developers.whatwg.org/content-models.html#kinds-of-content
var omitTags = []string{
	// Metadata content
	// https://developers.whatwg.org/content-models.html#metadata-content-0
	"base",
	"link", // @todo: unless itemprop attribute is present
	"meta", // @todo: unless itemprop attribute is present
	"noscript",
	"script",
	"style",
	"template",
	"title",

	// Flow control
	// https://developers.whatwg.org/content-models.html#flow-content-0
	//"a",
	"abbr",
	"address",
	//"area", (if it is a descendant of a map element)
	"article",
	//"aside",
	//"audio",
	//"b",
	//"bdi",
	//"bdo",
	"blockquote",
	//"br",
	"button",
	"canvas",
	"cite",
	"code",
	"data",
	"datalist",
	"del",
	//"details",
	//"dfn",
	"dialog",
	//"div",
	//"dl",
	//"em",
	"embed",
	"fieldset",
	//"figure",
	"footer",
	//"form",
	//"h1",
	//"h2",
	//"h3",
	//"h4",
	//"h5",
	//"h6",
	//"header",
	//"hgroup",
	"hr",
	//"i",
	"iframe",
	//"img",
	"input",
	"ins",
	"kbd",
	"keygen",
	//"label",
	"link", //(if the itemprop attribute is present)
	//"main",
	"map",
	//"mark",
	"math",
	"menu",
	"meta", //(if the itemprop attribute is present)
	"meter",
	"nav",
	"noscript",
	"object",
	//"ol",
	"output",
	//"p",
	//"pre",
	"progress",
	//"q",
	//"ruby",
	"s",
	//"samp",
	"script",
	//"section",
	"select",
	//"small",
	//"span",
	//"strong",
	"style", //(if the scoped attribute is present)
	//"sub",
	//"sup",
	"svg",
	//"table",
	"template",
	"textarea",
	"time",
	//"u",
	//"ul",
	//"var",
	"video",
	"wbr",

	// Interactive content (we exclude "a" and "img")
	// https://developers.whatwg.org/content-models.html#interactive-content
	// "a",
	"audio",
	"button",
	"details",
	"embed",
	"iframe",
	// "img",
	"input",
	"keygen",
	//"label",
	"object",
	"select",
	"textarea",
	"sorting",
	"video",

	// embedded content (not included above)
	// https://developers.whatwg.org/content-models.html#embedded-content-0
	// "audio",
	"canvas",
	// "embed",
	// "iframe",
	// "img",
	"math",
	"object",
	"svg",
	"video",
}

// If any class contains a *substring* in this list it will be filtered out.
var omitAttrValues = []string{
	"example",
	"footer",
	"header",
	"menu",
	"modal",
	"nav",
	"overlay",
	"social",
	"tip",
	"tooltip",
}

// Classes to always omit.
var alwaysOmit = []string{
	"hidden",
	"hide",
}

// http://www.w3.org/TR/wai-aria/roles#role_definitions_header
var omitAriaRoles = []string{
	"alert",
	"alertdialog",
	//"application",
	//"article",
	"banner",
	"button",
	"checkbox",
	//"columnheader",
	"combobox",
	"command",
	//"complementary",
	"composite",
	"contentinfo",
	//"definition",
	"dialog",
	"directory",
	//"document",
	"form",
	"grid",
	"gridcell",
	"group",
	//"heading",
	"img",
	"input",
	"landmark",
	//"link",
	//"list",
	"listbox",
	//"listitem",
	"log",
	//"main",
	"marquee",
	"math",
	"menu",
	"menubar",
	"menuitem",
	"menuitemcheckbox",
	"menuitemradio",
	"navigation",
	//"note",
	"option",
	"presentation",
	"progressbar",
	"radio",
	"radiogroup",
	"range",
	//"region",
	"roletype",
	//"row",
	//"rowgroup",
	//"rowheader",
	"scrollbar",
	"search",
	//"section",
	//"sectionhead",
	"select",
	"separator",
	"slider",
	"spinbutton",
	"status",
	//"structure",
	"tab",
	"tablist",
	"tabpanel",
	"textbox",
	"timer",
	"toolbar",
	"tooltip",
	//"tree",
	//"treegrid",
	//"treeitem",
	"widget",
	"window",
}

var sectionTags = []string{
	// Sectioning content;
	// https://developers.whatwg.org/content-models.html#sectioning-content-0
	"article",
	"aside",
	"nav",
	"section",

	// Heading content;
	// https://developers.whatwg.org/content-models.html#heading-content
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"hgroup",
}

// Tags that should have trailing newline.
var breakingTags = append(sectionTags, []string{
	"p",
	"div",
	"br",
	"ul",
	"ol",
}...)

var debug bool
var logger *log.Logger

func init() {
	// Currently if env DEBUG is any non null string we debug.
	env := os.Getenv("DEBUG")
	if env != "" {
		debug = true
		logger = log.New(os.Stderr, "html2text: ", log.LstdFlags|log.Lshortfile)
	}
}

// Given a curstate struct descend down the HTML hierarchy accumulating and
// plaintexting content along the way.
func textify(curState *state) error {
	var curNode *html.Node
	var err error
	var noRecurse bool
	var i int64
	curNode = curState.node

	if stringInSlice(curState.node.DataAtom.String(), breakingTags) {
		if _, err = curState.buf.WriteString("\n\n"); err != nil {
			return err
		}
	}

	if stringInSlice(curState.node.DataAtom.String(), sectionTags) {
		if _, err = curState.buf.WriteString("\n\n"); err != nil {
			return err
		}

		last := curState.node.DataAtom.String()[len(curState.node.DataAtom.String())-1:]
		if i, err = strconv.ParseInt(last, 10, 8); err == nil {
			if _, err = curState.buf.WriteString(strings.Repeat("#", int(i)) + " "); err != nil {
				return err
			}
		}
	}

	switch curState.node.Type {
	case html.TextNode:
		data := strings.Trim(spacingRe.ReplaceAllString(curState.node.Data, " "), "\r\n\t")
		// We want to write non-whitespace only strings but don't want to strip leading/trailing spaces on an actual string.
		if len(strings.TrimSpace(data)) > 0 {
			if _, err = curState.buf.WriteString(data); err != nil {
				return err
			}
		}
	case html.ElementNode:
		switch curState.node.DataAtom.String() {
		case "a":
			var text string
			text, err = handleATag(curState)
			if _, err = curState.buf.WriteString(text); err != nil {
				return err
			}
		case "ul", "ol":
			curState.indent++
		case "li":
			var text string
			text, err = handleChildren(curState)
			if strings.TrimSpace(text) != "" {
				if _, err = curState.buf.WriteString("\n" + strings.Repeat(" ", spacing*curState.indent) + "* " + strings.TrimSpace(text)); err != nil {
					return err
				}
			}
		case "table":
			var text string
			// indicate we need to capture headers
			curState.gatherHeaders = true
			text, err = handleChildren(curState)
			if strings.TrimSpace(text) != "" {
				if _, err = curState.buf.WriteString("\n\n" + strings.TrimSpace(text) + "\n\n"); err != nil {
					return err
				}
			}
		case "td", "th":
			var text string
			text, err = handleChildren(curState)
			if strings.TrimSpace(text) != "" {
				data := strings.TrimSpace(spacingRe.ReplaceAllString(text, " "))
				if len(data) > 0 {
					if curState.gatherHeaders {
						curState.headers = append(curState.headers, data)
					} else {
						if curState.column < len(curState.headers) {
							if _, err = curState.buf.WriteString(strings.TrimSpace(curState.headers[curState.column]+": "+data) + "\n"); err != nil {
								return err
							}
						} else {
							if _, err = curState.buf.WriteString(strings.TrimSpace(data) + "\n"); err != nil {
								return err
							}
						}
					}
				}
			}
		default:
			// Stop checking at first noRecurse. This is largely here to avoid recursing
			// down sections of the markup that are non-content related.
			if curState.node.DataAtom.String() != "" && stringInSlice(curState.node.DataAtom.String(), omitTags) {
				noRecurse = true
				break
			}

			if AttrHasString(curState.node, "class", alwaysOmit) && curNode.DataAtom.String() != "body" {
				if debug {
					logger.Printf("[%s] skipping recursion for %s bc class\n", curState.url, curState.node.DataAtom.String())
				}
				noRecurse = true
				break
			}

			// Omit classes except when they appear on the body
			if curState.omitClasses {
				if AttrHasString(curState.node, "class", omitAttrValues) && curNode.DataAtom.String() != "body" {
					if debug {
						logger.Printf("[%s] skipping recursion for %s bc class\n", curState.url, curState.node.DataAtom.String())
					}
					noRecurse = true
					break
				}
			}

			// Omit ids except when they appear on the body
			if curState.omitIds {
				if AttrHasString(curState.node, "id", omitAttrValues) && curNode.DataAtom.String() != "body" {
					if debug {
						logger.Printf("[%s] skipping recursion for %s bc id\n", curState.url, curState.node.DataAtom.String())
					}
					noRecurse = true
					break
				}
			}

			if curState.omitRoles {
				if AttrHasString(curState.node, "role", omitAriaRoles) {
					if debug {
						logger.Printf("[%s] skipping recursion for %s bc role\n", curState.url, curState.node.DataAtom.String())
					}
					noRecurse = true
					break
				}
			}

		}
	}
	if !noRecurse {
		for c := curState.node.FirstChild; c != nil; c = c.NextSibling {
			curState.node = c
			if err = textify(curState); err != nil {
				return err
			}
		}
	}

	// After tag work.
	switch curNode.DataAtom.String() {
	case "ul", "ol":
		curState.indent--
	case "td", "th":
		curState.column++
	case "tr":
		// Stop gathering headers
		curState.column = 0
		curState.gatherHeaders = false
		if err = curState.buf.WriteByte('\n'); err != nil {
			return err
		}
	case "table":
		curState.headers = []string{}
	}

	if stringInSlice(curState.node.DataAtom.String(), breakingTags) {
		if _, err = curState.buf.WriteString("\n\n"); err != nil {
			return err
		}
	}

	return nil
}

// Recurse down an "a" node tree and collect all node content and merge to
// one line of text. This is so that nested divs and other elements don't
// create excessive whitespace.
func handleATag(curState *state) (string, error) {
	var text string
	var err error
	for _, attr := range curState.node.Attr {
		if attr.Key == "href" {
			// Relative -> Absolute URL
			var link string
			if strings.HasPrefix(attr.Val, "#") || strings.HasPrefix(attr.Val, curState.url+"#") {
				// We don't want to capture internal links
			} else if strings.HasPrefix(attr.Val, "/") {
				link = curState.url + attr.Val
			} else if strings.HasPrefix(attr.Val, ".") {
				link = curState.url + "/" + attr.Val
			} else {
				link = attr.Val
			}
			buf := curState.buf
			curState.buf = &bytes.Buffer{}
			for c := curState.node.FirstChild; c != nil; c = c.NextSibling {
				curState.node = c
				if err = textify(curState); err != nil {
					return text, err
				}
			}
			if link != "" {
				linkText := strings.Trim(spacingRe.ReplaceAllString(curState.buf.String(), " "), "\r\n \t")
				text = "[" + linkText + "](" + spacingRe.ReplaceAllString(link, "") + ")"
			}
			curState.buf = buf
			break
		}
	}
	return text, err
}

// Accumulates child content. We use this rather than a normal recursion via
// textify when we may want to accumulate all child content before handling a
// particular node. eg. tables etc.
func handleChildren(curState *state) (string, error) {
	var text string
	var err error

	buf := curState.buf
	curState.buf = &bytes.Buffer{}
	for c := curState.node.FirstChild; c != nil; c = c.NextSibling {
		curState.node = c
		if err = textify(curState); err != nil {
			return text, err
		}
	}
	text = curState.buf.String()
	curState.buf = buf

	return text, err
}

// Returns true if the string 's' appears in any string in list, false otherwise.
func stringInSlice(s string, list []string) bool {
	for _, b := range list {
		if s == b {
			return true
		}
	}
	return false
}

// Returns true if the string 's' appears in any string in list, false otherwise.
func stringContainsSlice(s string, list []string) bool {
	for _, b := range list {
		if strings.Contains(s, b) {
			return true
		}
	}
	return false
}

// Returns true if the node has an attribute 'key' containing a value that is a
// a substring in any string in 'values'.
func AttrHasString(n *html.Node, key string, values []string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key {
			if stringContainsSlice(attr.Val, values) {
				if debug {
					logger.Printf("node(%s) has '%s' with '%s' matching string in (%s)\n", n.DataAtom.String(), key, attr.Val, strings.Join(values, ", "))
				}
				return true
			}
			break
		}
	}
	return false
}

func FromReader(reader io.Reader, url string, flags int) (string, error) {
	var err error
	curState := &state{}
	curState.buf = &bytes.Buffer{}
	curState.url = url
	if (flags & OmitClasses) != 0 {
		curState.omitClasses = true
	}
	if (flags & OmitIds) != 0 {
		curState.omitIds = true
	}
	if (flags & OmitRoles) != 0 {
		curState.omitRoles = true
	}
	curState.node, err = html.Parse(reader)
	if err != nil {
		return "", err
	}
	if err = textify(curState); err != nil {
		return "", err
	}
	text := strings.TrimSpace(newlineSpaceRe.ReplaceAllString(curState.buf.String(), "\n\n"))
	return text, nil
}

func FromString(input string, url string, flags int) (string, error) {
	text, err := FromReader(strings.NewReader(input), url, flags)
	if err != nil {
		return "", err
	}
	return text, nil
}
