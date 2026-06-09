package main

import (
	"log"
	"net/http"
)

func main() {
	routes()
	log.Println("Serveur démarré sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
