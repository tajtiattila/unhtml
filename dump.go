package unhtml

import (
	"code.google.com/p/go.net/html"
	"fmt"
	"strings"
)

func dumpTree(doc *html.Node) {
	var f func(l int, n *html.Node)
	f = func(l int, n *html.Node) {
		prefix := strings.Repeat("  ", l)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			switch c.Type {
			case html.TextNode:
				t := strings.TrimSpace(c.Data)
				if t != "" {
					fmt.Printf("%s%#v\n", prefix, t)
				}
			case html.DocumentNode, html.ElementNode:
				id, cls := "", ""
				for _, a := range c.Attr {
					switch a.Key {
					case "id":
						id = "#" + a.Val
					case "class":
						cls = "." + a.Val
					}
				}
				fmt.Print(prefix + c.Data + cls + id)
				if c.FirstChild == nil {
					fmt.Println(" { }")
				} else {
					fmt.Println(" {")
					f(l+1, c)
					fmt.Println(prefix + "}")
				}
			case html.CommentNode:
				fmt.Println(prefix + "// " + c.Data)
			}
		}
	}
	f(0, doc)
}

/*
type printer struct {
	w io.Reader
	err error
}
func (p *printer) p(data ...interface{}) {
	_, err := fmt.Fprint(p.w, data...)
	if err != nil && p.err == nil {
		p.err = nil
	}
}

func (p *printer) toHtml(node *html.Node) {
	switch node.Type {
	case TextNode:
	case DocumentNode:
	case ElementNode:
	case CommentNode:
	case DoctypeNode:
	)
	if node.Namespace == "" {
		p.p("<", node.Data)
	} else {
		p.p("<", node.NameSpace, ":", node.Data)
	}
	for _, a := range node.Attr {
		n, k, v := a
		if n == "" {
			p.p(" ", k, "=", v)
		} else {
			p.p(" ", n, ":", k, "=", v)
		}
		if p.err != nil {
			return
		}
	}
	if node.FirstChild != nil {
		p.p(">")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			p.toHtml(c)
		}
		if node.Namespace == "" {
			p.p("</", node.Data, ">")
		} else {
			p.p("</", node.NameSpace, ":", node.Data, ">")
		}
	} else {
		p.p("/>")
	}
}
*/
