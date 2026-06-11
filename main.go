package main

import (
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

func main() {
	Routes()
	log.Println("Serveur démarré sur http://localhost:8080")

	go func() {
		time.Sleep(300 * time.Millisecond)
		url := "http://localhost:8080"
		if err := openBrowser(url); err != nil {
			log.Printf("Impossible d'ouvrir le navigateur: %v", err)
		}
	}()

	http.ListenAndServe(":8080", nil)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
