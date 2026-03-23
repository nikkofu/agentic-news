package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nikkofu/agentic-news/internal/feedback"
)

func main() {
	stateDir := firstNonEmpty(os.Getenv("AGENTIC_NEWS_STATE_DIR"), "state")
	addr := firstNonEmpty(os.Getenv("AGENTIC_NEWS_FEEDBACK_ADDR"), ":8081")

	service := feedback.NewService(feedback.NewStore(stateDir))
	server := newServer(addr, feedback.NewHandler(service))
	log.Fatal(server.ListenAndServe())
}

func newServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
