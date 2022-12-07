package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/wkhere/htmlx"
	. "github.com/wkhere/htmlx/pred"
	"golang.org/x/net/html/atom"
)

func findInAll(w io.Writer, conf *config) (err error) {
	u, err := url.Parse(conf.startURL)
	if err != nil {
		return err
	}

	matchers := make([]*regexp.Regexp, len(conf.terms))
	for i, t := range conf.terms {
		matchers[i] = regexp.MustCompile("(?i)" + t)
	}

	for u != nil {
		var next *url.URL
		next, err = find(w, u, matchers)
		if err != nil {
			return fmt.Errorf("%s: %v", u, err)
		}
		u = next
	}
	return nil
}

func find(w io.Writer, u *url.URL, matchers []*regexp.Regexp) (next *url.URL,
	_ error) {
	req := &http.Request{Method: "GET", URL: u}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected get status: %s", resp.Status)
	}

	top, err := htmlx.FinderFromData(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// main div#content table.el-item.item. tbody tr td p.desc
	// + span.mainlink a -> link

	items, ok := top.
		Find(Element(atom.Main)).
		Find(Element(atom.Div, ID("content"))).
		FindWithSiblings(Element(atom.Table, Class("item")))

	if !ok {
		log.Println("unknown format for items in the issue:", u)
	}

	for section := range items {

		p := section.
			Find(Element(atom.Td)).
			Find(Element(atom.P, Class("desc")))

		s := dumpAllText(p)

		match := all(matchers, func(rx *regexp.Regexp) bool {
			return rx.MatchString(s)
		})
		if !match {
			continue
		}

		fmt.Fprintf(w, "** match in %s\n", u)

		link, ok := p.
			Find(Element(atom.Span, Class("mainlink"))).
			Find(Element(atom.A)).
			Attr().Val("href")

		if !ok || link == "" {
			fmt.Fprintln(w, "** link: not found")
		} else {
			fmt.Fprintln(w, "** link:", link)
		}

		fmt.Fprintln(w, "** text content:")
		io.WriteString(w, s)
		io.WriteString(w, "\n\n")
	}

	prev := findPrev(top)
	if prev.IsEmpty() {
		log.Println("last issue processed:", u)
		return nil, nil
	}

	link, ok := prev.Attr().Val("href")
	if !ok || link == "" {
		return nil, fmt.Errorf("invalid prev link: %v", prev)
	}
	return u.ResolveReference(&url.URL{Path: link}), nil
}

func findPrev(f htmlx.Finder) htmlx.Finder {
	return f.
		Find(Element(atom.Main)).
		Find(Element(atom.Div, Class("pager"))).
		Find(Element(atom.Div, Class("prev"))).
		Find(Element(atom.A))
}

func dumpAllText(f htmlx.Finder) string {
	b := new(strings.Builder)
	for x := range f.FindAll(AnyText()) {
		b.WriteString(x.Data)
	}
	return b.String()
}

func all[T any](xx []T, f func(T) bool) bool {
	for _, x := range xx {
		if !f(x) {
			return false
		}
	}
	return true
}
