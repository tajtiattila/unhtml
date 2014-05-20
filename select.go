package unhtml

import (
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	"strconv"
	"strings"
)

type Selector []selectorElem

func (sel Selector) String() string {
	ret, sep := "", ""
	for _, selem := range sel {
		ret, sep = ret+sep+selem.String(), "/"
	}
	return ret
}

func SelectorFromString(s string) (Selector, error) {
	i := 0
	var sel Selector
	for i < len(s) {
		j := i + ridx(s[i:], '/')
		var part string
		part, i = s[i:j], j+1

		elem, err := selectorElemFromString(s)
		if err != nil {
			return nil, err
		}
		sel = append(sel, elem)
	}
	return sel, nil
}

type matchNodeContext struct {
	tagMatchIdx int
}

type selectorElem interface {
	matchNode(node *html.Node, ctx *matchNodeContext) bool
	String() string
}

func selectorElemFromString(s string) (selectorElem, error) {
	if s == '*' {
		return new(matchAllSelectorElem), nil
	}
	return newTagSelectorElemFromString(se)
}

type matchAllSelectorElem struct{}

func (selem *matchAllSelectorElem) matchNode(node *html.Node, ctx *matchNodeContext) bool {
	return true
}

func (selem *matchAllSelectorElem) String() string {
	return "*"
}

type TagSelectorElem struct {
	Tag, Id, Cls string
	Index        int
}

func newTagSelectorElemFromString(se string) (selectorElem, error) {
	elem := &TagSelectorElem{index: -1}
	for {
		i := strings.LastIndexAny(se, ".#[")
		if i == -1 {
			break
		}
		switch se[i] {
		case '[':
			r := se[i+1:]
			if len(r) < 2 || r[len(r)-1] != ']' {
				return nil, NewErr("InvalidSelIndex", s)
			}
			var err error
			if elem.Index, err = strconv.Atoi(r[:len(r)-1]); err != nil {
				return nil, NewErr("InvalidSelIndex", s)
			}
		case '#':
			elem.Id = se[i+1:]
		case '.':
			elem.Cls = se[i+1:]
		}
		se = se[:i]
	}
	elem.Tag = strings.ToLower(se)
	if atom.Lookup(elem.Tag) == nil {
		return nil, NewErr("InvalidSelTag", se)
	}
	return elem
}

func (selem *TagSelectorElem) matchNode(node *html.Node, ctx *matchNodeContext) bool {
	switch node.Type {
	case html.DocumentNode, html.ElementNode:
		if strings.ToLower(node.Data) == selem.Tag {
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
				ret := selem.Index == -1 || selem.Index == ctx.tagMatchIdx
				ctx.tagMatchIdx++
				return ret
			}
		}
	}
	return false
}

func (selem *TagSelectorElem) String() string {
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
	return selem.Tag + scls + sid + sindex
}

//func selectNodesImpl2(node *html.Node, sel Selector, chquit <-chan bool, chmatch chan<- *html.Node) {
//}

func selectNodesImpl(node *html.Node, sel Selector, chquit <-chan bool, chmatch chan<- *html.Node) {
	if len(sel) == 0 {
		select {
		case <-chquit:
		case chmatch <- node:
		}
		return
	}
	selem, rest := sel[0], sel[1:]
	mnc := &matchNodeContext{}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if selem.matchNode(mnc, c) {
			selectNodesImpl(c, rest, chquit, chmatch)
		}
	}
}

func selectNodes(node *html.Node, sel Selector, chquit <-chan bool) <-chan *html.Node {
	ch := make(chan *html.Node)
	go func() {
		selectNodesImpl(node, sel, chquit, ch)
		close(ch)
	}()
	return ch
}

func selectNode(node *html.Node, sel Selector) *html.Node {
	if len(sel) == 0 {
		return node
	}
	chq := make(chan bool)
	found, ok := <-selectNodes(node, sel, chq)
	close(chq)
	if ok {
		return found
	}
	return nil
}
