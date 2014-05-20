package unhtml

import (
	"bytes"
	"code.google.com/p/go.net/html"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// DecodeFunc is a function to be used as context functions to be used
// in func= inside struct field tags. It receives the text part of
// the HTML node in question and should return the value suitable for
// the field in question.
type DecodeFunc func(data string) (interface{}, error)

// Context is used for HTML unmarshaling.
type Context struct {
	Funcs map[string]DecodeFunc
}

func NewContext() *Context {
	return &Context{Funcs: make(map[string]DecodeFunc)}
}

func (ctx *Context) AddFunc(n string, f DecodeFunc) {
	ctx.Funcs[n] = f
}

func (ctx *Context) UnmarshalHtml(r io.Reader, i interface{}) error {
	doc, err := html.Parse(r)
	if err != nil {
		return err
	}
	v := reflect.ValueOf(i)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("UnmarshalHtml recover", r)
			if err, _ = r.(error); err == nil {
				panic(r)
			}
		}
	}()
	ctx.unmarshal(doc, v, unhtmlTag{})
	return err
}

func (ctx *Context) unmarshal(node *html.Node, v reflect.Value, t unhtmlTag) {
	switch {
	case t.decodeFunc != "":
		df, ok := ctx.Funcs[t.decodeFunc]
		if !ok {
			panic(NewErr("MissingFunction", t.decodeFunc))
		}
		unmarshalSpec(node, v, t.sel, func(n *html.Node) reflect.Value {
			intf, err := df(nodeAsString(n))
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(intf)
		})
	case t.innerHtml:
		unmarshalSpec(node, v, t.sel, func(n *html.Node) reflect.Value {
			return reflect.ValueOf(innerHtml(n))
		})
	default:
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			unmarshalFunc(node, t.sel, func(s string) {
				if i, err := strconv.ParseInt(s, 10, v.Type().Bits()); err != nil {
					panic(err)
				} else {
					v.SetInt(i)
				}
			})
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			unmarshalFunc(node, t.sel, func(s string) {
				if u, err := strconv.ParseUint(s, 10, v.Type().Bits()); err != nil {
					panic(err)
				} else {
					v.SetUint(u)
				}
			})
		case reflect.String:
			unmarshalFunc(node, t.sel, func(s string) {
				v.SetString(s)
			})
		case reflect.Slice:
			v.SetLen(0)
			chq := make(chan bool)
			ve := reflect.New(v.Type().Elem()).Elem()
			for cnode := range selectNodes(node, t.sel, chq) {
				ctx.unmarshal(cnode, ve, t)
				v.Set(reflect.Append(v, ve))
			}
			close(chq)
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				vi, sf := v.Field(i), v.Type().Field(i)
				sft := decodeTag(sf.Tag)
				ctx.unmarshal(node, vi, sft)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func nodeAsStringRec(node *html.Node) (s string) {
	if node.Type == html.TextNode {
		s = node.Data
	} else {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			s = s + nodeAsStringRec(c)
		}
	}
	return
}

func nodeAsString(node *html.Node) (s string) {
	return strings.TrimSpace(nodeAsStringRec(node))
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

func unmarshalFunc(node *html.Node, sel Selector, f func(s string)) {
	cnode := selectNode(node, sel)
	if cnode == nil {
		return
	}
	f(strings.TrimSpace(nodeAsString(cnode)))
}

func unmarshalSpec(node *html.Node, v reflect.Value, sel Selector,
	f func(node *html.Node) reflect.Value) {
	switch v.Kind() {
	case reflect.Array:
		chq := make(chan bool)
		i := 0
		for cnode := range selectNodes(node, sel, chq) {
			if i < v.Len() {
				v.Index(i).Set(f(cnode))
				i++
			} else {
				close(chq)
				break
			}
		}
	case reflect.Slice:
		v.SetLen(0)
		chq := make(chan bool)
		for cnode := range selectNodes(node, sel, chq) {
			v = reflect.Append(v, f(cnode))
		}
		close(chq)
	default:
		node = selectNode(node, sel)
		if node != nil {
			v.Set(f(node))
		}
	}
}

type unhtmlTag struct {
	sel        Selector
	innerHtml  bool
	decodeFunc string
}

func (t *unhtmlTag) String() string {
	s := t.sel.String()
	if t.innerHtml {
		s += ",innerhtml"
	}
	if t.decodeFunc != "" {
		s += ",func=" + t.decodeFunc
	}
	return s
}

func ridx(s string, sch rune) int {
	for i, ch := range s {
		if ch == sch {
			return i
		}
	}
	return len(s)
}

func decodeTag(st reflect.StructTag) unhtmlTag {
	return decodeTagString(st.Get("unhtml"))
}

func decodeTagString(value string) unhtmlTag {
	i := ridx(value, ',')
	t := unhtmlTag{sel: nodeSelFromString(value[:i])}
	i++
	for i < len(value) {
		j := i + ridx(value[i:], ',')
		var extra string
		extra, i = value[i:j], j+1
		switch {
		case extra == "innerhtml":
			t.innerHtml = true
		case strings.HasPrefix(extra, "func="):
			t.decodeFunc = extra[5:]
		default:
			panic(NewErr("InvalidTag", value))
		}
	}
	return t
}
