package main

import (
	"net/http"
)

func main() {
	routes()

	http.ListenAndServe(":8080", nil)
}
