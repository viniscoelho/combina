package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/combina/src/db"
	"github.com/combina/src/router"
	"github.com/combina/src/storage/lottostore"
	"github.com/rs/cors"
)

func main() {
	initDB := flag.Bool("init-db", false, "creates a database and its tables")
	flag.Parse()
	if *initDB {
		db.InitializeDatabase()
	}

	ls, err := lottostore.NewLottoBacked()
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
