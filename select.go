package unhtml

import (
	"code.google.com/p/go.net/html"
	"strconv"
	"strings"
)

type SelectorElem struct {
	Tag, Id, Cls string
	index        int
}

func (selem SelectorElem) String() string {
	sid, scls, sindex := selem.Id, selem.Cls, ""
	if sid != "" {
		sid = "#" + sid
	}
	if scls != "" {
		scls = "." + scls
	}
	if selem.index != -1 {
		sindex = "[" + strconv.Itoa(selem.index) + "]"
	}
	return selem.Tag + scls + sid + sindex
}

type Selector []SelectorElem

func (sel Selector) String() string {
	ret, sep := "", ""
	for _, selem := range sel {
		ret, sep = ret+sep+selem.String(), "/"
	}
	return ret
}

func nodeSelFromString(s string) (sel Selector) {
	i := 0
	for i < len(s) {
		j := i + ridx(s[i:], '/')
		var tag string
		tag, i = s[i:j], j+1

		elem := SelectorElem{index: -1}
		for {
			ei := strings.LastIndexAny(tag, ".#[")
			if ei == -1 {
				break
			}
			switch tag[ei] {
			case '[':
				r := tag[ei+1:]
				if len(r) < 2 || r[len(r)-1] != ']' {
					panic(NewErr("InvalidSelIndex", s))
				}
				var err error
				if elem.index, err = strconv.Atoi(r[:len(r)-1]); err != nil {
					panic(NewErr("InvalidSelIndex", s))
				}
			case '#':
				elem.Id = tag[ei+1:]
			case '.':
				elem.Cls = tag[ei+1:]
			}
			tag = tag[:ei]
		}
		elem.Tag = strings.ToLower(tag)
		sel = append(sel, elem)
	}
	return
}

func matchNode(node *html.Node, selem SelectorElem) bool {
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
			return (selem.Id == "" || selem.Id == id) &&
				(selem.Cls == "" || selem.Cls == cls)
		}
	}
	return false
}

func selectNodesImpl(node *html.Node, sel Selector, chquit <-chan bool, chmatch chan<- *html.Node) {
	if len(sel) == 0 {
		select {
		case <-chquit:
		case chmatch <- node:
		}
		return
	}
	selem, rest := sel[0], sel[1:]
	idx := 0
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if matchNode(c, selem) {
			if selem.index == -1 || selem.index == idx {
				selectNodesImpl(c, rest, chquit, chmatch)
			}
			idx++
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
