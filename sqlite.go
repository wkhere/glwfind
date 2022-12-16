package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func dsn() string {
	return fmt.Sprintf("file:%s?mode=rw", dbfile())
}

func setupDB() (*sql.DB, error) {
	err := touch(dbfile())
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dsn())
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
	if _, err := os.Stat(dbfile()); os.IsNotExist(err) {
		f, err := os.Create(dbfile())
		if err != nil {
			return err
		}
		return f.Close()
	}
	return nil
}

func dbfile() string {
	if p := os.Getenv("GLWDB"); p != "" {
		return p
	}
	return filepath.Join(home(), ".glw.db")
}

func home() string {
	s := os.Getenv("HOME")
	if s == "" {
		u, err := user.Current()
		if err != nil {
			panic(err)
		}
		return u.HomeDir
	}
	return s
}
