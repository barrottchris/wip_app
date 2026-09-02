package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"wip/internal/app"
	"wip/internal/config"
	"wip/internal/db"
	"wip/internal/httpapi"
)

const serverAddress = "localhost:34115"

func main() {
	fmt.Println("Opening local SQLite database...")
	database, err := db.Start()
	if err != nil {
		log.Fatalf("failed to start database: %v", err)
	}
	defer database.Stop()
	handleGracefulShutdown(database)

	store := app.NewStore(database.Conn)
	if err := store.SeedIfEmpty(); err != nil {
		log.Fatalf("failed to seed initial data: %v", err)
	}
	runtimeTracker := app.NewRuntimeTracker()
	api := httpapi.NewServer(
		store,
		config.NewStore(database.Conn),
		runtimeTracker,
		app.NewProcessManager(runtimeTracker.SetRunning),
	)

	mux := http.NewServeMux()
	mux.Handle("/api/", api.Handler())
	mux.Handle("/", http.FileServer(http.Dir("./frontend/src")))

	fmt.Printf("WIP running at http://%s\n", serverAddress)
	log.Fatal(http.ListenAndServe(serverAddress, mux))
}

func handleGracefulShutdown(database *db.DB) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		_ = database.Stop()
		os.Exit(0)
	}()
}
