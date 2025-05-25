package main

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
)

type result struct {
	issueURL string
	refURL   string
	desc     string
}

const termsLimit = 20

func find(w io.Writer, db *sql.DB, terms []string) (err error) {
	if len(terms) < 1 {
		return fmt.Errorf("need at least 1 term")
	}
	if len(terms) > termsLimit {
		return fmt.Errorf("terms limit exceeded")
	}

	q := `
		SELECT i.url, r.link, r.desc
		FROM refs r JOIN issues i on (i.num=r.issuenum)
		WHERE `
	conds := make([]string, len(terms))
	for i, _ := range terms {
		conds[i] = `r.desc LIKE ?`
	}
	qtail := `
		ORDER BY i.num DESC, r.refnum`

	q = fmt.Sprint(q, strings.Join(conds, ` AND `), qtail)

	values := make([]any, len(terms))
	for i, term := range terms {
		values[i] = `%` + term + `%`
	}

	rows, err := db.Query(q, values...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var r result
		err = rows.Scan(&r.issueURL, &r.refURL, &r.desc)
		if err != nil {
			return err
		}
		print(w, &r)
	}

	return rows.Err()
}

func print(w io.Writer, r *result) {
	fmt.Fprintln(w, "* issue:", r.issueURL)
	fmt.Fprintln(w, "* ref:  ", r.refURL)
	fmt.Fprintln(w, "* desc: ", r.desc)
	fmt.Fprintln(w)
}
