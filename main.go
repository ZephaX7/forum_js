package main

import (
	"net/http"
	"os/exec"
)

func main() {
	routes()
	go exec.Command("cmd", "/c", "start", "http://localhost:8080").Run()
	http.ListenAndServe(":8080", nil)
}
