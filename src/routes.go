package src

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
)

/* =========================
   MODELS
========================= */

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

func Routes() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Méthode non autorisée"})
			return
		}

		LoginUser(w, r)
	})

	http.HandleFunc("/api/discussions", discussionsHandler)
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/register.html")
	})

	http.HandleFunc("/api/discussions/", getDiscussionDetail)
	http.HandleFunc("/api/replies", repliesHandler)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/login.html")
	})

	http.HandleFunc("/forum", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/forum.html")
	})

	http.HandleFunc("/discussion", discussionPage)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "templates/index.html")
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	var user User
	var storedPassword string

	err := DB.QueryRow(
		"SELECT id, username, email, password FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Username, &user.Email, &storedPassword)

	if err != nil || storedPassword != req.Password {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Identifiant incorrect",
		})
		return
	}

	json.NewEncoder(w).Encode(user)
}

func discussionPage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	tmpl, _ := template.ParseFiles("templates/discussion.html")
	tmpl.Execute(w, map[string]string{"id": id})
}

func discussionsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {

		rows, err := DB.Query(`
            SELECT d.id, d.title, d.content, d.user_id, u.username, d.category,
                   COUNT(r.id), d.created_at
            FROM discussions d
            LEFT JOIN users u ON d.user_id = u.id
            LEFT JOIN replies r ON d.id = r.discussion_id
            GROUP BY d.id, d.title, d.content, d.user_id, u.username, d.category, d.created_at
            ORDER BY d.created_at DESC
        `)
		if err != nil {
			json.NewEncoder(w).Encode([]Discussion{})
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
		return
	}

	if r.Method == "POST" {

		var req Discussion
		json.NewDecoder(r.Body).Decode(&req)

		var id int
		err := DB.QueryRow(
			"INSERT INTO discussions (title, content, user_id, category) VALUES ($1,$2,$3,$4) RETURNING id",
			req.Title, req.Content, req.UserID, req.Category,
		).Scan(&id)

		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": "erreur création"})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      id,
			"success": true,
		})
	}
}

func getDiscussionDetail(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Path[len("/api/discussions/"):]
	discussionID, _ := strconv.Atoi(id)

	var d Discussion

	err := DB.QueryRow(`
        SELECT d.id, d.title, d.content, d.user_id, u.username, d.category,
               COUNT(r.id), d.created_at
        FROM discussions d
        LEFT JOIN users u ON d.user_id = u.id
        LEFT JOIN replies r ON d.id = r.discussion_id
        WHERE d.id = $1
        GROUP BY d.id, d.title, d.content, d.user_id, u.username, d.category, d.created_at
    `, discussionID).Scan(
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
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return
	}

	json.NewEncoder(w).Encode(d)
}

func repliesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {

		id := r.URL.Query().Get("discussion_id")

		rows, _ := DB.Query(`
            SELECT id, content, user_id,
            (SELECT username FROM users WHERE id = replies.user_id),
            discussion_id, created_at
            FROM replies WHERE discussion_id = $1
            ORDER BY created_at ASC
        `, id)

		var replies []Reply

		for rows.Next() {
			var r Reply
			rows.Scan(&r.ID, &r.Content, &r.UserID, &r.Username, &r.DiscussionID, &r.Created)
			replies = append(replies, r)
		}

		json.NewEncoder(w).Encode(replies)
	}

	if r.Method == "POST" {

		var req Reply
		json.NewDecoder(r.Body).Decode(&req)

		DB.Exec(
			"INSERT INTO replies (content, user_id, discussion_id) VALUES ($1,$2,$3)",
			req.Content, req.UserID, req.DiscussionID,
		)

		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}
}
