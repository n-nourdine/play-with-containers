package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/n-nourdine/play-with-containers/inventory-app/handlers"
)

func main() {

	l := log.New(os.Stdout, fmt.Sprintf("inventory-app running on port %s -> ", os.Getenv("INVENTORY_APP_PORT")), log.LstdFlags)

	h, err := handlers.NewHandler(l)
	if err != nil {
		l.Fatal(err.Error())
	}
	defer h.C.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthy", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("site ok")) })
	mux.HandleFunc("GET /api/movies", h.GetMovies)           // retrieve all the movies. or retrieve all the movies with name in the title.(GET /api/movies?title=[name])
	mux.HandleFunc("GET /api/movies/{id}", h.GetMovie)       // retrieve a single movie by id.
	mux.HandleFunc("POST /api/movies", h.AddMovie)           // create a new product entry.
	mux.HandleFunc("PUT /api/movies/{id}", h.UpdateMovie)    // update a single movie by id.
	mux.HandleFunc("DELETE /api/movies/{id}", h.DeleteMovie) // delete a single movie by id.
	mux.HandleFunc("DELETE /api/movies", h.DeleteMovies)     // delete all movies in the database.

	s := http.Server{
		Addr:         fmt.Sprintf(":%v", os.Getenv("INVENTORY_APP_PORT")),
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
			h.L.Fatalf("Error starting server: %s/n", err)
		}

	}()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	sig := <-c
	h.L.Println("Got signal:", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()
	s.Shutdown(ctx)
}


