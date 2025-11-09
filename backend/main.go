package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mongoURI := getenv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := getenv("MONGODB_DB", "mewakir")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := NewMeetingStore(ctx, mongoURI, dbName)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer store.Disconnect(context.Background())

	hub := NewHub(store)
	go hub.Run()

	srv := NewServer(store, hub)

	port := getenv("PORT", "8080")

	go func() {
		log.Printf("server listening on :%s", port)
		if err := srv.Start(":" + port); err != nil {
			log.Fatalf("listen: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
