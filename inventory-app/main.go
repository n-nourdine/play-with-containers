package main

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/n-nourdine/play-with-containers/inventory-app/database"
)

func main() {
	l := log.New(os.Stdout, "inventory-app ", log.LstdFlags)

	err := Load("/home/nasser/nasdev/play-with-containers/.env")
	if err != nil {
		l.Println(err)
		os.Exit(1)
	}

	db, err := database.NewConn(os.Getenv("DATABASE_URL"))
	if err != nil {
		l.Println(err.Error())
		os.Exit(1)
	}

	defer db.Close()
	mux := http.NewServeMux()

	// mux.HandleFunc("GET /api/movies",) // retrieve all the movies. or retrieve all the movies with name in the title.(GET /api/movies?title=[name])
	// mux.HandleFunc("GET /api/movies/{id}",) // retrieve a single movie by id.
	// mux.HandleFunc("POST /api/movies",) // create a new product entry.
	// mux.HandleFunc("PUT /api/movies/{id}",) // update a single movie by id.
	// mux.HandleFunc("DELETE /api/movies/{id}",) // delete a single movie by id.
	// mux.HandleFunc("DELETE /api/movies",) // delete all movies in the database.

	s := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ErrorLog:     l,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		l.Println()
		err := s.ListenAndServe()
		if err == nil {
			log.Printf("Error starting server: %s/n", err)
			os.Exit(1)
		}

	}()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	sig := <-c
	log.Println("Got signal:", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()
	s.Shutdown(ctx)
}

func Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key != "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
