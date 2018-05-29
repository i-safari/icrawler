package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/ahmdrz/goinsta"
	_ "github.com/mattn/go-sqlite3"
)

var schemas = []string{
	`CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY,
	username VARCHAR(32) NULL,
	fullname VARCHAR(64) NULL,
	biography TEXT NULL,
	email VARCHAR(64) NULL,
	phone VARCHAR(16) NULL,
	followers INT NULL,
	following INT NULL,
	media INT NULL,
	highlights INT NULL,
	isprivate INT NULL
);`,
	`CREATE TABLE IF NOT EXISTS feed (
	userid INTEGER NOT NULL,
	id INTEGER NOT NULL,
	username VARCHAR(32) NULL,
	path VARCHAR(256) NOT NULL,
	deleted INT NULL,
	PRIMARY KEY(userid, id),
	FOREIGN KEY(userid) REFERENCES users
);`,
	`CREATE TABLE IF NOT EXISTS stories (
	userid INTEGER NOT NULL,
	id INTEGER NOT NULL,
	title VARCHAR(32) NULL,
	username VARCHAR(32) NULL,
	path VARCHAR(256) NOT NULL,
	highlight INT NULL,
	deleted INT NULL,
	PRIMARY KEY(userid, id),
	FOREIGN KEY(userid) REFERENCES users
);`,
}

var (
	errNotFound = errors.New("not found")
)

type dbConn struct {
	conn *sql.DB
	file string
}

func dbOpen(file string) *dbConn {
	conn, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatalln(err)
	}
	db := &dbConn{conn, file}
	db.Create()
	return db
}

func (db *dbConn) Close() {
	db.conn.Close()
}

func (db *dbConn) ReConn() {
	err := db.conn.Ping()
	if err == nil {
		return
	}

	db.conn, err = sql.Open("sqlite3", db.file)
	if err != nil {
		log.Println(err)
	}
}

func (db *dbConn) Create() {
	for i := range schemas {
		_, err := db.conn.Exec(schemas[i])
		if err != nil {
			log.Println(err)
		}
	}
}

func (db *dbConn) Get(name string) (*goinsta.User, error) {
	db.ReConn()

	stmt, err := db.conn.Prepare("SELECT * FROM users WHERE username = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	user := insta.NewUser()

	rows, err := stmt.Query(name)
	if err == nil {
		defer rows.Close()
		if !rows.Next() {
			log.Println(name, "does not exists")
			return user, errNotFound
		}
		err = rows.Scan(
			&user.ID, &user.Username, &user.FullName, &user.Biography,
			&user.PublicEmail, &user.PublicPhoneNumber, &user.FollowerCount,
			&user.FollowingCount, &user.MediaCount, &user.Gender, &user.IsPrivate,
		)
	}
	return user, err
}

func (db *dbConn) Put(user *goinsta.User) error {
	db.ReConn()
	stmt, err := db.conn.Prepare(`INSERT INTO users (
		id, username, fullname, biography, email, phone, followers,
		following, media, highlights, isprivate) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err == nil {
		_, err = stmt.Exec(
			user.ID, user.Username, user.FullName, user.Biography,
			user.PublicEmail, user.PublicPhoneNumber, user.FollowerCount,
			user.FollowingCount, user.MediaCount, user.Gender, user.IsPrivate,
		)
		stmt.Close()
	}
	return err
}

func (db *dbConn) Update(user *goinsta.User) error {
	db.ReConn()
	stmt, err := db.conn.Prepare(
		`UPDATE users SET id = ?,
		username = ?, fullname = ?, biography = ?,
		email = ?, phone = ?, followers = ?, following = ?,
		media = ?, highlights = ?, isprivate = ? WHERE 
		id = ?`,
	)
	if err == nil {
		_, err = stmt.Exec(
			user.ID, user.Username, user.FullName, user.Biography,
			user.PublicEmail, user.PublicPhoneNumber, user.FollowerCount,
			user.FollowingCount, user.MediaCount, user.Gender, user.IsPrivate,
			user.ID,
		)
		stmt.Close()
	}
	return err
}

func (db *dbConn) SetTitle(item *goinsta.Item, title string) error {
	db.ReConn()
	stmt, err := db.conn.Prepare(
		`UPDATE stories SET title = ? WHERE id = ?`,
	)
	if err == nil {
		_, err = stmt.Exec(title, item.ID)
		stmt.Close()
	}
	return err
}

func (db *dbConn) ExistsStory(item *goinsta.Item) bool {
	stmt, err := db.conn.Prepare(
		`SELECT path FROM stories WHERE id = ?`,
	)
	if err == nil {
		defer stmt.Close()
		rows, err := stmt.Query(item.ID)
		if err == nil {
			defer rows.Close()
			path := ""
			if rows.Next() {
				rows.Scan(&path)
			}
			if path == "" {
				return false
			}
			_, err = os.Stat(path)
			if err != nil {
				return false
			}
		}
	}
	return true
}

func (db *dbConn) ExistsMedia(item *goinsta.Item) bool {
	stmt, err := db.conn.Prepare(
		`SELECT path FROM feed WHERE id = ?`,
	)
	if err == nil {
		defer stmt.Close()
		rows, err := stmt.Query(item.ID)
		if err == nil {
			defer rows.Close()
			path := ""
			if rows.Next() {
				rows.Scan(&path)
			}
			if path == "" {
				return false
			}
			_, err = os.Stat(path)
			if err != nil {
				return false
			}
		}
	}
	return true
}

func (db *dbConn) PutStory(user *goinsta.User, item *goinsta.Item, h bool, o string) error {
	db.ReConn()
	stmt, err := db.conn.Prepare(
		`INSERT INTO stories (userid, id, username, highlight, path) VALUES (?, ?, ?, ?, ?)`,
	)
	if err == nil {
		_, err = stmt.Exec(user.ID, item.ID, user.Username, h, o)
		stmt.Close()
	}
	return err
}

func (db *dbConn) PutMedia(user *goinsta.User, item *goinsta.Item, o string) error {
	db.ReConn()
	stmt, err := db.conn.Prepare(
		`INSERT INTO feed (userid, id, username, path) VALUES (?, ?, ?, ?)`,
	)
	if err == nil {
		_, err = stmt.Exec(user.ID, item.ID, user.Username, o)
		stmt.Close()
	}
	return err
}
