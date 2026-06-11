package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Init DB
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
    CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        content TEXT
    );
    `

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
}

// Struct Post
type Post struct {
	ID      int
	Title   string
	Content string
}

// Handler : afficher forum
func forumHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, content FROM posts")
	if err != nil {
		http.Error(w, "Erreur DB", 500)
		return
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post
		rows.Scan(&p.ID, &p.Title, &p.Content)
		posts = append(posts, p)
	}

	tmpl, err := template.ParseFiles("templates/forum.html")
	if err != nil {
		http.Error(w, "Template error", 500)
		return
	}

	tmpl.Execute(w, posts)
}

// Handler : créer un post
func createPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/forum", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	_, err := db.Exec("INSERT INTO posts(title, content) VALUES (?, ?)", title, content)
	if err != nil {
		http.Error(w, "Erreur insertion", 500)
		return
	}

	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

// Fonction à appeler dans routes()
func forumRoutes() {
	initDB()

	http.HandleFunc("/forum", forumHandler)
	http.HandleFunc("/create-post", createPostHandler)
}
