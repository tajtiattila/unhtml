package unhtml

import (
	"code.google.com/p/go.net/html"
)

type Matcher struct {
	chquit  chan bool
	chmatch chan *html.Node
	seen    map[*html.Node]bool
}

type stopMatcher struct{}

// Select creates a Matcher yielding children of node matching sel.
func Select(node *html.Node, sel Selector) *Matcher {
	m := &Matcher{
		chmatch: make(chan *html.Node),
		chquit:  make(chan bool),
		seen:    make(map[*html.Node]bool),
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(stopMatcher); !ok {
					panic(r)
				}
			}
		}()
		m.selectImpl(node, sel)
		close(m.chmatch)
	}()
	return m
}

func (m *Matcher) Match() *html.Node {
	return <-m.chmatch
}

func (m *Matcher) Matches() <-chan *html.Node {
	return m.chmatch
}

func (m *Matcher) Close() error {
	if m.chquit != nil {
		close(m.chquit)
		m.chquit = nil
	}
	return nil
}

func (m *Matcher) selectImpl(node *html.Node, sel Selector) {
	if len(sel) == 0 {
		if !m.seen[node] {
			m.seen[node] = true
			select {
			case <-m.chquit:
				panic(stopMatcher{})
			case m.chmatch <- node:
			}
		}
		return
	}
	selem, rest := sel[0], sel[1:]
	mnc := &matchNodeContext{}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		enter, rec := selem.matchNode(c, mnc)
		if enter {
			m.selectImpl(c, rest)
		}
		if rec {
			m.selectImpl(c, sel)
		}
	}
}

// SelectFirst selects the first child of node matching sel.
func SelectFirst(node *html.Node, sel Selector) *html.Node {
	if len(sel) == 0 {
		return node
	}
	s := Select(node, sel)
	defer s.Close()
	return s.Match()
}
