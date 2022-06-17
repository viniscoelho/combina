package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"combina/src/db"
	"combina/src/router"
	"combina/src/storage"
	"github.com/rs/cors"
)

func main() {
	initDB := flag.Bool("init-db", false, "creates a database and its tables")
	flag.Parse()
	if *initDB {
		db.InitializeDatabase()
	}

	ls, err := storage.NewLottoBacked()
	if err != nil {
		log.Fatalf("could not initialize storage: %s", err)
	}

	r := router.CreateRoutes(ls)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
	})
	handler := c.Handler(r)

	s := &http.Server{
		Handler:      handler,
		ReadTimeout:  0,
		WriteTimeout: 0,
		Addr:         ":3000",
		IdleTimeout:  time.Second * 60,
	}
	log.Fatal(s.ListenAndServe())
}
