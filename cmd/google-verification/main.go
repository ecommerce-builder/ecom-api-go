package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

func googlePropertyVerificationHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("google-site-verification: googleae3066b83fcfab45.html"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()
	srv := &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: r}
	r.Route("/googleae3066b83fcfab45.html", func(r chi.Router) {
		r.Get("/", googlePropertyVerificationHandler)
	})

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
