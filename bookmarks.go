package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func verifyApiKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("API_KEY")
		apiKeyHeader := r.Header.Get("X-BOOKMARKS-API-KEY")
		if apiKeyHeader != apiKey {
			http.Error(w, errorMessage, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	dbpool, db_err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if db_err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to DB: %v\n", db_err)
		os.Exit(1)
	}
	defer dbpool.Close()
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.AllowContentType("application/json"))
	router.Use(middleware.Heartbeat("/app/health"))
	router.Use(verifyApiKey)
	router.Get("/show/{id}", getBookmark(dbpool))
	router.Post("/create", createBookmark(dbpool))
	router.Put("/update/{id}", updateBookmark(dbpool))
	router.Delete("/remove/{id}", deleteBookmark(dbpool))
	router.Get("/list", listBookmarks(dbpool))
	router.Post("/search", searchBookmark(dbpool))
	router.Mount("/bookmarks/action", router)
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(errorMessage))
	})
	log.Println("Sever running at Port 8085")
	log.Fatal(http.ListenAndServe("127.0.0.1:8085", router))
}
