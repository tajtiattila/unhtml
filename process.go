package unhtml

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	"strings"
)

func cloneNode(node *html.Node) *html.Node {
	x := &html.Node{
		Type:      node.Type,
		DataAtom:  node.DataAtom,
		Data:      node.Data,
		Namespace: node.Namespace,
		Attr:      make([]html.Attribute, 0, len(node.Attr)),
	}
	for _, a := range node.Attr {
		x.Attr = append(x.Attr, a)
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		xc := cloneNode(c)
		xc.Parent = x
		if x.FirstChild == nil {
			x.FirstChild = xc
		} else {
			x.LastChild.NextSibling = xc
			xc.PrevSibling = x.LastChild
		}
		x.LastChild = xc
	}
	return x
}

func absUrl(node *html.Node, baseurl string) {
	avv := []struct {
		t atom.Atom
		a string
	}{
		{atom.A, "href"},
		{atom.Img, "src"},
		{atom.Link, "rel"},
	}
	for _, v := range avv {
		if v.t == node.DataAtom {
			for i, a := range node.Attr {
				if a.Key == v.a && strings.Index(a.Val, "://") == -1 {
					node.Attr[i] = html.Attribute{a.Namespace, a.Key, baseurl + a.Val}
				}
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		absUrl(c, baseurl)
	}
}

func innerHtml(node *html.Node) string {
	buf := &bytes.Buffer{}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(buf, c); err != nil {
			panic(err)
		}
	}
	return strings.TrimSpace(buf.String())
}
