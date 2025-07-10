// billing
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

	"github.com/n-nourdine/play-with-containers/billing-app/handler"
)

func main() {
	logger := log.New(os.Stdout, fmt.Sprintf("billing-app running on port %s -> ",
		os.Getenv("BILLING_APP_PORT")), log.LstdFlags)

	// // Connect to database
	// store, err := database.NewConn()
	// if err != nil {
	// 	logger.Fatalf("Erreur de connexion à la base de données: %v", err)
	// }
	// defer store.Close()

	// // Create RabbitMQ consumer
	// consumer, err := rabbitmq.NewConsumer(logger, store)
	// if err != nil {
	// 	logger.Fatalf("Erreur de création du consommateur RabbitMQ: %v", err)
	// }
	// defer consumer.Close()

	// // Create context for graceful shutdown
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// // Start consuming messages
	// err = consumer.StartConsuming(ctx)
	// if err != nil {
	// 	logger.Fatalf("Erreur lors du démarrage de la consommation: %v", err)
	// }

	h, err := handler.NewHandler(logger)
	if err != nil {
		logger.Fatalf("Erreur de connexion à la base de données: %v", err)
	}
	defer h.C.Close()
	
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/order", h.Add)
	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("GET /api/orders", h.GetAllOrders)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", os.Getenv("BILLING_APP_PORT")),
		Handler:      mux,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		log.Println()
		err := server.ListenAndServe()
		if err == nil {
			log.Fatalf("Error starting server: %s/n", err)
		}

	}()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	sig := <-c
	log.Println("Got signal:", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()
	server.Shutdown(ctx)
}
