package src

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

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

var db *sql.DB

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
		_, err := DB.Exec(table)
		if err != nil {
			log.Printf("Erreur création table: %v\n", err)
		}
	}
}

func Routes() {

	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/login.html")
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/register.html")
	})

	http.HandleFunc("/forum", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/forum.html")
	})

	http.HandleFunc("/api/discussions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			getDiscussions(w, r)
		} else if r.Method == "POST" {
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
		if r.Method == "GET" {
			getReplies(w, r)
		} else if r.Method == "POST" {
			createReply(w, r)
		}
	})

	http.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			registerUser(w, r)
		}
	})

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			loginUser(w, r)
		}
	})
}

func getDiscussions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := DB.Query(`
        SELECT d.id, d.title, d.content, d.user_id, u.username, d.category,
               COUNT(r.id) as reply_count, d.created_at
        FROM discussions d
        LEFT JOIN users u ON d.user_id = u.id
        LEFT JOIN replies r ON d.id = r.discussion_id
        GROUP BY d.id, d.title, d.content, d.user_id, u.username, d.category, d.created_at
        ORDER BY d.created_at DESC
    `)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var discussions []Discussion

	for rows.Next() {
		var d Discussion
		err := rows.Scan(
			&d.ID,
			&d.Title,
			&d.Content,
			&d.UserID,
			&d.Username,
			&d.Category,
			&d.Replies,
			&d.Created,
		)
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

	result, err := DB.Exec(
		"INSERT INTO discussions (title, content, user_id, category) VALUES ($1, $2, $3, $4	)",
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
	err := DB.QueryRow(`
		SELECT d.id, d.title, d.content, d.user_id, u.username, d.category, 
		       COUNT(r.id) as reply_count, d.created_at
		FROM discussions d
		LEFT JOIN users u ON d.user_id = u.id
		LEFT JOIN replies r ON d.id = r.discussion_id
		WHERE d.id = $1
		GROUP BY d.id
	`, discussionID).Scan(&d.ID, &d.Title, &d.Content, &d.UserID, &d.Username, &d.Category, &d.Replies, &d.Created)

	if err != nil {
		http.Error(w, "Discussion non trouvée", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(d)
}

func getReplies(w http.ResponseWriter, r *http.Request) {
	discussionID := r.URL.Query().Get("discussion_id")

	rows, err := DB.Query(`
		SELECT id, content, user_id, (SELECT username FROM users WHERE id = replies.user_id), 
		       discussion_id, created_at
		FROM replies
		WHERE discussion_id = $1
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

func createReply(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content      string `json:"content"`
		UserID       int    `json:"user_id"`
		DiscussionID int    `json:"discussion_id"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	result, err := DB.Exec(
		"INSERT INTO replies (content, user_id, discussion_id) VALUES ($1, $2, $3)",
		req.Content, req.UserID, req.DiscussionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, _ := result.LastInsertId()
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "success": true})
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	result, err := DB.Exec(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		req.Username, req.Email, req.Password)
	if err != nil {
		http.Error(w, "Utilisateur déjà existant", http.StatusBadRequest)
		return
	}

	id, _ := result.LastInsertId()
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "success": true})
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	var user User
	err := DB.QueryRow(
		"SELECT id, username, email FROM users WHERE email = $1 AND password = $2",
		req.Email, req.Password).Scan(&user.ID, &user.Username, &user.Email)

	if err != nil {
		http.Error(w, "Email ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(user)
}
