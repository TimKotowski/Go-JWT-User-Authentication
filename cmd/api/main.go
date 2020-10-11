package main

import (
	"Go-JWT-Auth/api/config"
	"Go-JWT-Auth/api/middleware/auth"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.ParseConfigFile("config.json")
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nsdsd %+v", cfg)

	cfg.DBHost = os.Getenv("DB_HOST")
	cfg.DBPort = os.Getenv("DB_PORT")
	cfg.DBName = os.Getenv("DB_NAME")
	cfg.DBUser = os.Getenv("DB_USER")
	cfg.DBPass = os.Getenv("DB_PASS")

	postgres := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	db, err := sql.Open("postgres", postgres)

	fmt.Printf("\nsdsd %+v", cfg)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("connected")

	r := chi.NewRouter()

	fs := http.FileServer(http.Dir("./views/"))
	r.Handle("/", http.StripPrefix("", fs))

	auth.New(db, r)

	// Create a new HTTP server.
	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Printf("Running server...")

	// Start the HTTP server.
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}

}
