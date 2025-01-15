package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/wkhere/htmlx"
	. "github.com/wkhere/htmlx/pred"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func feedAll(db *sql.DB) (err error) {
	url1 := "https://golangweekly.com/latest"
	// Note: redirect works only from /latest, not from /issues/latest
	url, err := redirect(url1)
	if err != nil {
		return fmt.Errorf("redirect of %s: %w", url1, err)
	}

	inum, err := parseIssueNum(url)
	if err != nil {
		return fmt.Errorf("latest issue num not found: %w", err)
	}

	const earliestIssueNum = 41 // we know that

	for ; inum >= earliestIssueNum; inum-- {
		if inum == 187 {
			continue // the one missing
		}

		url := fmt.Sprintf("https://golangweekly.com/issues/%d", inum)

		done, err := upsertIssue(db, inum, url)
		if err != nil {
			return fmt.Errorf("upsert issue#%d: %w", inum, err)
		}
		if done {
			continue
		}

		consolef("\rprocessing issue#%d\t", inum)

		all, err := feed1(db, inum, url)
		if err != nil || !all {
			consoleln()
		}
		if err != nil {
			return fmt.Errorf("issue#%d: %w", inum, err)
		}
	}

	consoleln()
	return vacuum(db)
}

func feed1(db *sql.DB, inum int, url string) (all bool, _ error) {

	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("get error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("GET status: %s", resp.Status)
	}

	top, err := htmlx.FinderFromData(resp.Body)
	if err != nil {
		return false, fmt.Errorf("html parse error: %w", err)
	}

	all, err = feedRefs(db, inum, top)
	if err != nil {
		return all, fmt.Errorf("refs feed error: %w", err)
	}
	if all {
		err = finishIssue(db, inum)
		if err != nil {
			return all, fmt.Errorf("update-as-done error: %w", err)
		}
	}

	return all, nil
}

func feedRefs(db *sql.DB, inum int, top htmlx.Finder) (all bool, _ error) {

	var log1st = true
	var logf = func(format string, a ...any) {
		if log1st {
			log1st = false
			consoleln()
		}
		consolef("\t"+format+"\n", a...)
	}

	var selector = struct {
		items func() htmlx.FinderStream
		desc  func(item htmlx.Finder) htmlx.FinderStream
		link  func(item htmlx.Finder) htmlx.Finder
	}{}

	switch {
	default:
		// items: body/main div#content table.item
		// item:
		//   desc: all text children of: tbody/tr/td/p.desc
		//   link from p.desc: span/a
		//
		selector.items = func() htmlx.FinderStream {
			return top.
				Find(Element(atom.Main)).
				Find(Element(atom.Div, ID("content"))).
				FindWithSiblings(Element(atom.Table, Class("item")))
		}
		p := func(item htmlx.Finder) htmlx.Finder {
			return item.
				Find(Element(atom.Td)).
				Find(Element(atom.P, Class("desc")))
		}
		selector.desc = func(item htmlx.Finder) htmlx.FinderStream {
			return p(item).
				FindAll(AnyText())
		}
		selector.link = func(item htmlx.Finder) htmlx.Finder {
			return p(item).
				Find(Element(atom.Span)).
				Find(Element(atom.A))
		}

	case inum == 307:
		// one exceptional issue #307:
		// items: body/main div#content/table.content td/ul/li
		// item:
		//   desc: all text children
		//   link from item: a
		//
		selector.items = func() htmlx.FinderStream {
			return top.
				Find(Element(atom.Main)).
				Find(Element(atom.Div, ID("content"))).
				Find(Element(atom.Table, Class("content"))).
				Find(Element(atom.Td)).
				Find(Element(atom.Ul)).
				FindWithSiblings(Element(atom.Li))
		}
		selector.desc = func(item htmlx.Finder) htmlx.FinderStream {
			return item.
				FindAll(AnyText())
		}
		selector.link = func(item htmlx.Finder) htmlx.Finder {
			return item.
				Find(Element(atom.A))
		}

	case inum <= 204:
		// older issues:
		// items: body/main table.container table.item(findall)
		// item:
		//   desc: all non-ws text children of: tbody/tr[0..n-1] (*)
		//   (*) applies to len(tr)>1, if =1 then that tr is descended
		//   link from item: tbody/tr/td.link a
		selector.items = func() htmlx.FinderStream {
			return top.
				Find(Element(atom.Main)).
				Find(Element(atom.Table, Class("container"))).
				FindAll(Element(atom.Table, Class("item")))
		}
		selector.desc = func(item htmlx.Finder) htmlx.FinderStream {
			trs := item.
				FindWithSiblings(Element(atom.Tr)).
				Collect()
			switch len(trs) {
			case 0:
				return htmlx.Inject(nil)
			case 1:
				return htmlx.AllText(trs[0])
			default:
				// btw, the last tr is author/source
				return htmlx.
					Inject(trs[:len(trs)-1]).
					Join(htmlx.AllText)
			}
		}
		selector.link = func(item htmlx.Finder) htmlx.Finder {
			return item.
				Find(Element(atom.Td, Class("link"))).
				Find(Element(atom.A))
		}
	}

	missing := false
	refnum := 0

	for item := range selector.items() {
		refnum++

		link, ok := selector.link(item).Attr().Val("href")
		if !ok {
			logf("ref#%d: no link", refnum)
			missing = true
			continue
		}
		var rid string
		{
			lid, err := parseReflinkID(link)
			if err == nil {
				rid = fmt.Sprintf("l:%d", lid)
			} else {
				rid = fmt.Sprintf("g:%d:%s", inum, minihash(link))
			}
		}

		s := dumpAll(selector.desc(item).Collect())

		err := upsertRef(db, rid, inum, refnum, link, s)
		if err != nil {
			// upsert errors terminate refs loop
			// - if db goes funky it probably will also for the next ref
			return false, fmt.Errorf("refid#%s ref#%d upsert: %w", rid, refnum, err)
		}
	}

	if refnum == 0 {
		logf("unknown html for items")
		return false, nil
	}

	return !missing, nil
}

func dumpAll(ff []htmlx.Finder) string {
	b := new(strings.Builder)
	for _, f := range ff {
		b.WriteString(f.Data)
		if f.Type == html.TextNode && !tailWS(f.Data) {
			b.WriteString("  ")
		}
	}
	return b.String()
}

func redirect(url string) (string, error) {
	resp, err := http.Head(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HEAD status: %s", resp.Status)
	}
	return resp.Request.URL.String(), nil
}
