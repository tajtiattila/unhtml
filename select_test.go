package unhtml

import (
	"code.google.com/p/go.net/html"
	"strings"
	"testing"
)

func TestSelectorFromString(t *testing.T) {
	SelectorFromString("html/body/div.container/div.contentbox")
	SelectorFromString("table.commentheader/tbody/tr.commentheader/td[0]/a.commentmem")
	SelectorFromString("table.commentheader/tbody/tr.commentheader/td[3]/a.contlinks")
}

func TestDecodeTag(t *testing.T) {
	decodeTag("html/body/div.container/div.contentbox")
	decodeTag("table.commentheader/tbody/tr.commentheader/td[0]/a.commentmem")
	decodeTag("table.commentheader/tbody/tr.commentheader/td[3]/a.contlinks,func=idx")
}

var s1html = `
  <table width="100%" class=commentheader>
        <tr class=commentheader>
	        <td width=138><a href="main.pl?amp;menupage=member&amp;mid=16265" class=commentmem>Szlarti</a>
			    <a href="main.pl?menupage=send_msg2&amp;rcptlist=16265" class=cmdlinks title="Privát üzenet küldése">(pü)</a></td>
		<td width=169><B>#3751 <a name=3751></a></td>
		<td width=209><B>  09:33:40</td>
		<td width=150 align=right>&nbsp;Elõzmeny:<A class=contlinks HREF=#3749 TITLE="">3749</A> </td>
		<td align=right width=70>
		<A class=cmdlinks href="main.pl?menupage=forum_comment_inapp&amp;comment_id=1465192" target=_blank title="Törlésre javaslom!">
		 X</A>&nbsp;
		 <a href=#top><img src=images/up.gif alt="" title="Vissza a tetejére"></td>
	</tr></table>
	<table width="100%" class=topictable>
	    <tr>
		<td colspan=2 valign=top class=topictable #dddddd>
	Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.<br>
		      </td>
		      </tr></table>
	<table width="100%" class=topictable border=0 cellpadding=0 cellspacing=0><tr>
		    <td colspan=3 align=right bgcolor=#dddddd>
		     &nbsp;&nbsp;&nbsp;&nbsp;
		    <A class=cmdlinks HREF="main.pl?menupage=forum_comment&amp;inreply=3751&amp;limit=30">Válasz erre</A>
		    </td>
		    </tr>
		 </table>`

func TestSelectNode(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(s1html))
	if err != nil {
		t.Fatal(err)
	}

	vts := []string{
		"html/body/table.commentheader/tbody/tr.commentheader/td[0]/a.commentmem",
		"html/body/table.commentheader/tbody/tr.commentheader/td[2],func=ts",
		"html/body/table.commentheader/tbody/tr.commentheader/td[1],func=idx",
		"html/body/table.commentheader/tbody/tr.commentheader/td[3]/a.contlinks,func=idx",
		"html/body/table.topictable[0],innerhtml",
	}

	for _, ts := range vts {
		tag := decodeTagString(ts)
		n := SelectFirst(doc, tag.sel)
		if n == nil || n == doc {
			t.Error(ts, tag.String(), "yields", n)
		}
	}

	if t.Failed() {
		dumpTree(doc)
	}
}
