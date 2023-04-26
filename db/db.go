package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/saparaly/snippentbox/pkg/models"
)

const (
	users = `CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE,
		email TEXT UNIQUE,
		password TEXT
		);`
	post = `CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER ,
			username TEXT,
			title TEXT,
			content TEXT,
			created DATETIME,
			tags TEXT,
			like INTEGER DEFAULT 0,
			dislike INTEGER DEFAULT 0,
			FOREIGN KEY (user_id) REFERENCES Users(id)
		);`

	postIndex = `CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created);`
	comment   = `CREATE TABLE IF NOT EXISTS Comments (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		post_id INTEGER,
		username TEXT,
		content TEXT ,
		like INTEGER DEFAULT 0,
		dislike INTEGER DEFAULT 0,
		FOREIGN KEY (user_id) REFERENCES Users(id),
		FOREIGN KEY (post_id) REFERENCES Posts(id)
	);`
	categories = `CREATE TABLE IF NOT EXISTS Categories (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		post_id INTEGER,
		FOREIGN KEY (post_id) REFERENCES Post(id)
	);`
	session = `CREATE TABLE IF NOT EXISTS session (
		ID INTEGER PRIMARY KEY,
		UserID INTEGER,
		Token TEXT UNIQUE,
		ExpirationDate TIMESTAMP
	);`
	postLikes = `CREATE TABLE IF NOT EXISTS post_likes (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		post_id INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES Users(id),
		FOREIGN KEY (post_id) REFERENCES Posts(id),
		UNIQUE (user_id, post_id)
	);`
	postDislikes = ` CREATE TABLE IF NOT EXISTS post_dislikes (
		id INTEGER PRIMARY KEY,
		post_id INTEGER,
		user_id INTEGER,
		FOREIGN KEY(post_id) REFERENCES posts(id),
		FOREIGN KEY(user_id) REFERENCES users(id),
		UNIQUE (user_id, post_id)
	);`
	commentLikes = `CREATE TABLE IF NOT EXISTS comment_likes (
		id INTEGER PRIMARY KEY,
		comment_id INTEGER,
		user_id INTEGER,
		FOREIGN KEY(comment_id) REFERENCES Comments(id),
		FOREIGN KEY(user_id) REFERENCES users(id),
		UNIQUE (user_id, comment_id)
	);`
	commentDislikes = `CREATE TABLE IF NOT EXISTS comment_dislikes (
		id INTEGER PRIMARY KEY,
		comment_id INTEGER,
		user_id INTEGER,
		FOREIGN KEY(comment_id) REFERENCES Comments(id),
		FOREIGN KEY(user_id) REFERENCES users(id),
		UNIQUE (user_id, comment_id)
	);`
)

func CreateDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func CreateTables(b *sql.DB) error {
	var quesries []string = []string{users, post, postIndex, comment, postLikes, postDislikes, commentLikes, commentDislikes, categories, session}
	for _, each := range quesries {
		_, err := b.Exec(each)
		if err != nil {
			fmt.Println("db has not created")
			return err
		}
	}
	return nil
}

func GetUserByUsername(b *sql.DB, username string) (*models.User, error) {
	stmt, err := b.Prepare("SELECT id, name, email, password FROM users WHERE name = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var user models.User
	err = stmt.QueryRow(username).Scan(&user.Id, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func CreateSession(b *sql.DB, session models.Session) error {
	exists, err := GerSessionByUserId(b, session.UserID)
	if err != nil {
		return err
	}
	if exists != nil {
		getId, err := b.Query("SELECT ID from session WHERE UserID = ?", session.UserID)
		var sesId int
		for getId.Next() {
			var id int
			getId.Scan(&id)
			sesId = id
		}

		stmt, err := b.Prepare("UPDATE session Set Token = ?, ExpirationDate = ? WHERE ID = ?")
		if err != nil {
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(session.Token, session.ExpirationDate, sesId)
		if err != nil {
			return err
		}
		return nil
	}
	stmt, err := b.Prepare("INSERT INTO session(UserID, Token, ExpirationDate) VALUES(?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(session.UserID, session.Token, session.ExpirationDate)
	if err != nil {
		return err
	}
	return nil
}

func DeleteSession(db *sql.DB, token string) error {
	stmt, err := db.Prepare("DELETE FROM session WHERE Token = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(token)
	if err != nil {
		return err
	}

	return nil
}

func GerSessionByUserId(b *sql.DB, userId int) (*models.Session, error) {
	row := b.QueryRow("SELECT * FROM session WHERE UserID = ? ORDER BY ExpirationDate DESC LIMIT 1", userId)

	var session models.Session
	err := row.Scan(&session.ID, &session.UserID, &session.Token, &session.ExpirationDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func GetSessionByToken(db *sql.DB, token string) (*models.Session, error) {
	session := &models.Session{}
	err := db.QueryRow("SELECT ID, UserID, Token, ExpirationDate FROM session WHERE Token = ?", token).Scan(&session.ID, &session.UserID, &session.Token, &session.ExpirationDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return session, nil
}
