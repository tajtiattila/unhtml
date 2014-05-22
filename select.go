package unhtml

import (
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	"strconv"
	"strings"
)

// Selector is something like XPath for HTML
type Selector []selectorElem

// SelectorFromString creates a Selector from s. The string
// is made up from path parts separated by '/'.
//
// Possible path elements:
// * (asterisk)       select all direct children
// // (double slash)  select children recursively
// tag                matches HTML nodes "tag"
// .class             matches HTML nodes with specified class
// #id                matches HTML nodes with specified id
// [index]            matches HTML nodes with given index
//
// tag, #id, .class and [index] may be combined.
//
// Examples:
// */table.comment/tbody/tr/td[0]
//    matches the first table cell of all rows in tables with class "comment"
func SelectorFromString(s string) (Selector, error) {
	i := 0
	var sel Selector
	rec := false
	for i < len(s) {
		j := i + ridx(s[i:], '/')
		var part string
		part, i = s[i:j], j+1

		if part == "" {
			rec = true
			continue
		}
		elem, err := selectorElemFromString(part)
		if err != nil {
			return nil, err
		}
		if rec {
			elem = &recursiveSelectorElem{base: elem}
			rec = false
		}
		sel = append(sel, elem)
	}
	if rec {
		return nil, NewErr("TrailingSlash", s)
	}
	return sel, nil
}

// Get canonical representation of sel.
func (sel Selector) String() string {
	ret, sep := "", ""
	for _, selem := range sel {
		ret, sep = ret+sep+selem.String(), "/"
	}
	return ret
}

////////////////////////////////////////////////////////////////////////////////

type matchNodeContext struct {
	tagMatchIdx int
}

type selectorElem interface {
	matchNode(node *html.Node, ctx *matchNodeContext) (exact bool, rec bool)
	String() string
}

func selectorElemFromString(s string) (selectorElem, error) {
	if s == "*" {
		return new(matchAllSelectorElem), nil
	}
	return newTagSelectorElemFromString(s)
}

// matchAllSelectorElem matches base and searches all children
type recursiveSelectorElem struct {
	base selectorElem
}

func (selem *recursiveSelectorElem) String() string {
	return "/" + selem.base.String()
}

func (selem *recursiveSelectorElem) matchNode(node *html.Node, ctx *matchNodeContext) (bool, bool) {
	exact, _ := selem.base.matchNode(node, ctx)
	return exact, true
}

// matchAllSelectorElem selects all direct children
type matchAllSelectorElem struct{}

func (selem *matchAllSelectorElem) matchNode(node *html.Node, ctx *matchNodeContext) (bool, bool) {
	return true, false
}

func (selem *matchAllSelectorElem) String() string {
	return "*"
}

// tagSelectorElem selects direct children according to tag/id/cls/index spec
type tagSelectorElem struct {
	Atom    atom.Atom
	Id, Cls string
	Index   int
}

func newTagSelectorElemFromString(s string) (selectorElem, error) {
	elem := &tagSelectorElem{Index: -1}
	for {
		i := strings.LastIndexAny(s, ".#[")
		if i == -1 {
			break
		}
		switch s[i] {
		case '[':
			r := s[i+1:]
			if len(r) < 2 || r[len(r)-1] != ']' {
				return nil, NewErr("InvalidSelIndex", s)
			}
			var err error
			if elem.Index, err = strconv.Atoi(r[:len(r)-1]); err != nil {
				return nil, NewErr("InvalidSelIndex", s)
			}
		case '#':
			elem.Id = s[i+1:]
		case '.':
			elem.Cls = s[i+1:]
		}
		s = s[:i]
	}
	elem.Atom = atom.Lookup([]byte(strings.ToLower(s)))
	if s != "" && elem.Atom == 0 {
		return nil, NewErr("InvalidSelTag", s)
	}
	return elem, nil
}

func (selem *tagSelectorElem) matchNode(node *html.Node, ctx *matchNodeContext) (exact bool, rec bool) {
	switch node.Type {
	case html.DocumentNode, html.ElementNode:
		if selem.Atom == 0 || node.DataAtom == selem.Atom {
			var id, cls string
			for _, a := range node.Attr {
				switch a.Key {
				case "id":
					id = a.Val
				case "class":
					cls = a.Val
				}
			}
			match := (selem.Id == "" || selem.Id == id) &&
				(selem.Cls == "" || selem.Cls == cls)
			if match {
				exact = selem.Index == -1 || selem.Index == ctx.tagMatchIdx
				ctx.tagMatchIdx++
			}
		}
	}
	return
}

func (selem *tagSelectorElem) String() string {
	sid, scls, sindex := selem.Id, selem.Cls, ""
	if sid != "" {
		sid = "#" + sid
	}
	if scls != "" {
		scls = "." + scls
	}
	if selem.Index != -1 {
		sindex = "[" + strconv.Itoa(selem.Index) + "]"
	}
	return selem.Atom.String() + scls + sid + sindex
}
