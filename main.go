package main

import (
	"net/http"
	"os/exec"
)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/index.html")
	})
	go exec.Command("cmd", "/c", "start", "http://localhost:8080").Run()
	http.ListenAndServe(":8080", nil)
}
