package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type PGRepo struct {
	db *sql.DB
}

func NewPGRepo(dbURL string) *PGRepo {
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		log.Println("can't open database")
		log.Println(err)
		os.Exit(1)
	}

	return &PGRepo{db}
}

func (r *PGRepo) FindBySlug(slug string) (*Page, error) {
	stmt, err := r.db.Prepare(
		"select title, body, created, modified from pages where slug = $1")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	row := stmt.QueryRow(slug)
	if row == nil {
		// if it's not in the database, we make a blank one
		now := time.Now()
		return &Page{slug, "", "", now, now}, nil
	}
	var title string
	var body string
	var created time.Time
	var modified time.Time

	row.Scan(&title, &body, &created, &modified)

	p := Page{slug, title, body, created, modified}
	return &p, nil
}

func (r *PGRepo) Store(page *Page) error {
	tx, err := r.db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}

	cstmt, err := tx.Prepare("select count(*) from pages where slug = $1")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}
	row := cstmt.QueryRow(page.Slug)
	var cnt int
	row.Scan(&cnt)
	now := time.Now()

	if cnt > 0 {
		stmt, err := tx.Prepare(
			"update pages set title = $1, body = $2, modified = $3 where slug = $4")

		if err != nil {
			log.Println(err)
			tx.Rollback()
			return err
		}
		_, err = stmt.Exec(page.Title, page.Body, now, page.Slug)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return err
		}
	} else {
		stmt, err := tx.Prepare(
			`insert into pages (slug, title, body, modified, created)
                   values($1,   $2,    $3,   $4,       $5)`)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return err
		}
		_, err = stmt.Exec(page.Slug, page.Title, page.Body, page.Modified, page.Created)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
