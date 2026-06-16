package main

import (
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"forum_js/src"
)

func main() {
	src.InitDB()
	src.Routes()

	log.Println("Serveur démarré sur http://localhost:8080")

	go func() {
		time.Sleep(300 * time.Millisecond)
		openBrowser("http://localhost:8080")
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		log.Println("Erreur ouverture navigateur:", err)
	}
}
