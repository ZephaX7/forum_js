package src

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func verifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	if len(req.Username) < 3 || len(req.Password) < 6 {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		http.Error(w, "Erreur hash", http.StatusInternalServerError)
		return
	}

	var userID int
	err = DB.QueryRow(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id",
		req.Username, req.Email, hash,
	).Scan(&userID)

	if err != nil {
		http.Error(w, "Utilisateur existe déjà", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user_id": userID,
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	var id int
	var username, email, hash string

	err := DB.QueryRow(
		"SELECT id, username, email, password FROM users WHERE email=$1",
		req.Email,
	).Scan(&id, &username, &email, &hash)

	if err != nil {
		http.Error(w, "Identifiants invalides", http.StatusUnauthorized)
		return
	}

	if verifyPassword(hash, req.Password) != nil {
		http.Error(w, "Identifiants invalides", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"id":       id,
		"username": username,
		"email":    email,
	})
}
