package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/wkhere/tilde"
	_ "modernc.org/sqlite"
)

func dbfile() (string, error) {
	p := os.Getenv("GLWDB")
	if p == "" {
		p = "~/.glw.db"
	}
	return tilde.Expand(p)
}

func dsn(path string) string {
	return fmt.Sprintf("file:%s?mode=rw", path)
}

func setupDB() (*sql.DB, error) {
	p, err := dbfile()
	if err != nil {
		return nil, err
	}
	err = touch(p)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dsn(p))
	if err != nil {
		return nil, err
	}
	err = createMissingSchema(db)
	if err != nil {
		return db, fmt.Errorf("create missing schema: %w", err)
	}
	return db, nil
}

func createMissingSchema(db *sql.DB) (err error) {
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS issues (
			num INTEGER NOT NULL,
			url  TEXT NOT NULL,
			done BOOLEAN NOT NULL,
			PRIMARY KEY (num)
		);
		CREATE TABLE IF NOT EXISTS refs (
			linkid   INTEGER NOT NULL,
			issuenum INTEGER NOT NULL,
			refnum	 INTEGER NOT NULL,
			link     TEXT NOT NULL,
			desc     TEXT NOT NULL,
			FOREIGN KEY (issuenum) REFERENCES issues(num),
			PRIMARY KEY (linkid)
		);
	`)
	return err
}

func upsertIssue(db *sql.DB, inum int, url string) (done bool, err error) {
	_, err = db.Exec(
		`INSERT OR IGNORE INTO issues (num, url, done) VALUES (?, ?, false)`,
		inum, url,
	)
	if err != nil {
		return false, err
	}
	row := db.QueryRow(
		`SELECT done FROM issues WHERE num=?`, inum,
	)
	err = row.Scan(&done)
	return done, err
}

func finishIssue(db *sql.DB, inum int) (err error) {
	_, err = db.Exec(
		`UPDATE issues SET done=true WHERE num=?`, inum,
	)
	return err
}

func upsertRef(db *sql.DB, lid, inum, refnum int, link, desc string) (
	err error) {
	_, err = db.Exec(`
		INSERT INTO refs (linkid, issuenum, refnum, link, desc)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (linkid) DO UPDATE SET refnum=?, link=?, desc=?
		`,
		lid, inum, refnum, link, desc,
		refnum, link, desc,
	)
	return err
}

func vacuum(db *sql.DB) (err error) {
	_, err = db.Exec(`VACUUM`)
	return err
}

func touch(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		return f.Close()
	}
	return nil
}
