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
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Données invalides",
		})
		return
	}

	if len(req.Username) < 3 || len(req.Password) < 6 {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Données invalides",
		})
		return
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Erreur hash",
		})
		return
	}

	var userID int
	err = DB.QueryRow(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id",
		req.Username, req.Email, hash,
	).Scan(&userID)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Utilisateur existe déjà",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user_id": userID,
	})
}
