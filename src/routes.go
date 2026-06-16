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

/* =========================
   ROUTES
========================= */

func Routes() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})

	http.HandleFunc("/forum", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/forum.html")
	})

	// ✅ NOUVELLE PAGE DISCUSSION
	http.HandleFunc("/discussion", discussionPage)

	// API
	http.HandleFunc("/api/discussions", discussionsHandler)
	http.HandleFunc("/api/discussions/", getDiscussionDetail)
	http.HandleFunc("/api/replies", repliesHandler)
}

/* =========================
   PAGE DISCUSSION HTML
========================= */

func discussionPage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	tmpl, _ := template.ParseFiles("templates/discussion.html")
	tmpl.Execute(w, map[string]string{"id": id})
}

/* =========================
   API DISCUSSIONS
========================= */

func discussionsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		rows, err := DB.Query(`
            SELECT d.id, d.title, d.content, d.user_id, u.username, d.category,
                   COUNT(r.id), d.created_at
            FROM discussions d
            LEFT JOIN users u ON d.user_id = u.id
            LEFT JOIN replies r ON d.id = r.discussion_id
            GROUP BY d.id, u.username
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
			rows.Scan(&d.ID, &d.Title, &d.Content, &d.UserID, &d.Username, &d.Category, &d.Replies, &d.Created)
			discussions = append(discussions, d)
		}

		json.NewEncoder(w).Encode(discussions)
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
			http.Error(w, err.Error(), 500)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      id,
			"success": true,
		})
	}
}

/* =========================
   API DETAIL DISCUSSION
========================= */

func getDiscussionDetail(w http.ResponseWriter, r *http.Request) {

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
        GROUP BY d.id, u.username
    `, discussionID).Scan(
		&d.ID, &d.Title, &d.Content, &d.UserID,
		&d.Username, &d.Category, &d.Replies, &d.Created,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(d)
}

/* =========================
   API REPLIES
========================= */

func repliesHandler(w http.ResponseWriter, r *http.Request) {

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
