package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// Structures de données
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Discussion struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Category string `json:"category"`
	Replies  int    `json:"replies"`
	Created  string `json:"created_at"`
}

type Reply struct {
	ID           int    `json:"id"`
	Content      string `json:"content"`
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	DiscussionID int    `json:"discussion_id"`
	Created      string `json:"created_at"`
}

var Db *sql.DB
var Tmpl *template.Template

func init() {
	var err error
	Db, err = sql.Open("sqlite3", "forum.db")
	if err != nil {
		log.Fatal(err)
	}
	if err = Db.Ping(); err != nil {
		log.Fatal(err)
	}
	createTables()

	// Charger les templates
	Tmpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Erreur chargement templates:", err)
	}
}

func createTables() {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS discussions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			category TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS replies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			discussion_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id),
			FOREIGN KEY(discussion_id) REFERENCES discussions(id)
		)`,
	}

	for _, table := range tables {
		_, err := Db.Exec(table)
		if err != nil {
			log.Printf("Erreur création table: %v\n", err)
		}
	}
}

func Routes() {

	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		Tmpl.ExecuteTemplate(w, "index.html", nil)
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		Tmpl.ExecuteTemplate(w, "login.html", nil)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		Tmpl.ExecuteTemplate(w, "register.html", nil)
	})

	http.HandleFunc("/forum", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		Tmpl.ExecuteTemplate(w, "forum.html", nil)
	})

	http.HandleFunc("/api/discussions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "GET":
			getDiscussions(w, r)
		case "POST":
			createDiscussion(w, r)
		}
	})

	http.HandleFunc("/api/discussions/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			getDiscussionDetail(w, r)
		}
	})

	http.HandleFunc("/api/replies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "GET":
			getReplies(w, r)
		case "POST":
			createReply(w, r)
		}
	})

	http.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			RegisterUser(w, r)
		}
	})

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			LoginUser(w, r)
		}
	})
}

func getDiscussions(w http.ResponseWriter, _ *http.Request) {
	rows, err := Db.Query(`
		SELECT d.id, d.title, d.content, d.user_id, u.username, d.category, 
		       COUNT(r.id) as reply_count, d.created_at
		FROM discussions d
		LEFT JOIN users u ON d.user_id = u.id
		LEFT JOIN replies r ON d.id = r.discussion_id
		GROUP BY d.id
		ORDER BY d.created_at DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	discussions := []Discussion{}
	for rows.Next() {
		var d Discussion
		err := rows.Scan(&d.ID, &d.Title, &d.Content, &d.UserID, &d.Username, &d.Category, &d.Replies, &d.Created)
		if err != nil {
			continue
		}
		discussions = append(discussions, d)
	}

	json.NewEncoder(w).Encode(discussions)
}

// Créer une discussion
func createDiscussion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		UserID   int    `json:"user_id"`
		Category string `json:"category"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	result, err := Db.Exec(
		"INSERT INTO discussions (title, content, user_id, category) VALUES (?, ?, ?, ?)",
		req.Title, req.Content, req.UserID, req.Category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, _ := result.LastInsertId()
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "success": true})
}

// Récupérer les détails d'une discussion
func getDiscussionDetail(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/discussions/"):]
	discussionID, _ := strconv.Atoi(id)

	var d Discussion
	err := Db.QueryRow(`
		SELECT d.id, d.title, d.content, d.user_id, u.username, d.category, 
		       COUNT(r.id) as reply_count, d.created_at
		FROM discussions d
		LEFT JOIN users u ON d.user_id = u.id
		LEFT JOIN replies r ON d.id = r.discussion_id
		WHERE d.id = ?
		GROUP BY d.id
	`, discussionID).Scan(&d.ID, &d.Title, &d.Content, &d.UserID, &d.Username, &d.Category, &d.Replies, &d.Created)

	if err != nil {
		http.Error(w, "Discussion non trouvée", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(d)
}

// Récupérer les réponses d'une discussion
func getReplies(w http.ResponseWriter, r *http.Request) {
	discussionID := r.URL.Query().Get("discussion_id")

	rows, err := Db.Query(`
		SELECT id, content, user_id, (SELECT username FROM users WHERE id = replies.user_id), 
		       discussion_id, created_at
		FROM replies
		WHERE discussion_id = ?
		ORDER BY created_at ASC
	`, discussionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	replies := []Reply{}
	for rows.Next() {
		var r Reply
		err := rows.Scan(&r.ID, &r.Content, &r.UserID, &r.Username, &r.DiscussionID, &r.Created)
		if err != nil {
			continue
		}
		replies = append(replies, r)
	}

	json.NewEncoder(w).Encode(replies)
}

// Créer une réponse
func createReply(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content      string `json:"content"`
		UserID       int    `json:"user_id"`
		DiscussionID int    `json:"discussion_id"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	result, err := Db.Exec(
		"INSERT INTO replies (content, user_id, discussion_id) VALUES (?, ?, ?)",
		req.Content, req.UserID, req.DiscussionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, _ := result.LastInsertId()
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "success": true})
}
