package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Hasher le mot de passe avec bcrypt
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

// Vérifier le mot de passe
func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Enregistrer un utilisateur
func registerUser(w http.ResponseWriter, r *http.Request) {
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

	// Valider les champs
	if len(req.Username) < 3 || len(req.Email) < 5 || len(req.Password) < 6 {
		http.Error(w, "Username (3+ caractères), email valide et password (6+ caractères) requis", http.StatusBadRequest)
		return
	}

	// Hasher le mot de passe
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		http.Error(w, "Erreur lors du hachage du mot de passe", http.StatusInternalServerError)
		return
	}

	// Insérer dans la base de données
	result, err := db.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		req.Username, req.Email, hashedPassword)

	if err != nil {
		http.Error(w, "Utilisateur déjà existant ou erreur base de données", http.StatusBadRequest)
		return
	}

	id, _ := result.LastInsertId()
	user := User{
		ID:       int(id),
		Username: req.Username,
		Email:    req.Email,
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "user": user})
}

// Connecter un utilisateur
func loginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	var user User
	var hashedPassword string

	// Récupérer l'utilisateur de la base de données
	err := db.QueryRow(
		"SELECT id, username, email, password FROM users WHERE email = ?",
		req.Email).Scan(&user.ID, &user.Username, &user.Email, &hashedPassword)

	if err != nil {
		http.Error(w, "Email ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	// Vérifier le mot de passe
	if err := verifyPassword(hashedPassword, req.Password); err != nil {
		http.Error(w, "Email ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "user": user})
}
