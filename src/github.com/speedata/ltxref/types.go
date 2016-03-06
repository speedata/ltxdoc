package ltxref

import (
	"html/template"
	"strings"
)

type Argumenttype int

const (
	_ Argumenttype = iota
	MANDARG
	MANDLIST
	OPTARG
	OPTLIST
	TODIMENORSPREADDIMEN
)

var argumenttypemap map[string]Argumenttype
var argumentTypeReveseMap map[Argumenttype]string

func init() {
	argumenttypemap = map[string]Argumenttype{
		"mandarg":              MANDARG,
		"mandlist":             MANDLIST,
		"optarg":               OPTARG,
		"optlist":              OPTLIST,
		"todimenorspreaddimen": TODIMENORSPREADDIMEN,
	}
	argumentTypeReveseMap = make(map[Argumenttype]string, len(argumenttypemap))
	for key, value := range argumenttypemap {
		argumentTypeReveseMap[value] = key
	}
}

type DocumentClasses []*DocumentClass

func (slice DocumentClasses) Len() int {
	return len(slice)
}

func (slice DocumentClasses) Less(i, j int) bool {
	return strings.ToLower(slice[i].Name) < strings.ToLower(slice[j].Name)
}

func (slice DocumentClasses) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type Environments []*Environment

func (slice Environments) Len() int {
	return len(slice)
}

func (slice Environments) Less(i, j int) bool {
	return strings.ToLower(slice[i].Name) < strings.ToLower(slice[j].Name)
}

func (slice Environments) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type Commands []*Command

func (slice Commands) Len() int {
	return len(slice)
}

func (slice Commands) Less(i, j int) bool {
	return strings.ToLower(slice[i].Name) < strings.ToLower(slice[j].Name)
}

func (slice Commands) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type Packages []*Package

func (slice Packages) Len() int {
	return len(slice)
}

func (slice Packages) Less(i, j int) bool {
	return strings.ToLower(slice[i].Name) < strings.ToLower(slice[j].Name)
}

func (slice Packages) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// The LaTeX reference knows about commands, environments, documentclasses and packages
type Ltxref struct {
	Commands        Commands
	Environments    Environments
	DocumentClasses DocumentClasses
	Packages        Packages
	Version         string
}

type DocumentClass struct {
	Name             string
	Label            []string
	Level            string
	ShortDescription map[string]string
	Description      map[string]template.HTML
	Optiongroup      []Optiongroup
}

func NewDocumentClass() *DocumentClass {
	dc := &DocumentClass{}
	dc.ShortDescription = make(map[string]string)
	dc.Description = make(map[string]template.HTML)
	return dc
}

type Optiongroup struct {
	Name             string
	ShortDescription map[string]string
	Classoption      []Classoption
}

type Classoption struct {
	Name             string
	Default          bool
	ShortDescription map[string]string
	Description      map[string]template.HTML
}

func NewCommand() *Command {
	c := &Command{}
	c.ShortDescription = make(map[string]string)
	c.Description = make(map[string]template.HTML)
	return c
}

type Command struct {
	Name             string
	Level            string
	Label            []string
	ShortDescription map[string]string
	Description      map[string]template.HTML
	Variant          []Variant
}

type Packageoption struct {
	Name             string
	Default          bool
	ShortDescription map[string]string
	Description      map[string]template.HTML
}

func NewPackage() *Package {
	p := &Package{}
	p.Label = make([]string, 0)
	p.LoadsPackages = make([]string, 0)
	p.ShortDescription = make(map[string]string)
	p.Description = make(map[string]template.HTML)
	return p
}

type Package struct {
	Name             string
	Level            string
	Label            []string
	LoadsPackages    []string
	ShortDescription map[string]string
	Description      map[string]template.HTML
	Commands         Commands
	Options          []Packageoption
}

type Environment struct {
	Name             string
	Level            string
	Label            []string
	ShortDescription map[string]string
	Description      map[string]template.HTML
	Variant          []Variant
}

func NewEnvironment() *Environment {
	e := &Environment{}
	e.Label = make([]string, 0)
	e.ShortDescription = make(map[string]string)
	e.Description = make(map[string]template.HTML)
	return e
}

func NewVariant() *Variant {
	v := &Variant{}
	v.Arguments = make([]*Argument, 0)
	v.Description = make(map[string]template.HTML)
	return v
}

// Some commands can have variants, such as \section or \section*.
// These commands are similar, so they should be documented together.
type Variant struct {
	Name        string
	Arguments   []*Argument
	Description map[string]template.HTML
}

func NewArgument() *Argument {
	return &Argument{}
}

// Argument of a command or an environment
type Argument struct {
	Optional bool
	Name     string
	Type     Argumenttype
}
