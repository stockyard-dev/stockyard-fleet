package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-fleet/internal/server"
	"github.com/stockyard-dev/stockyard-fleet/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9809"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./fleet-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("fleet: %v", err)
	}
	defer db.Close()

	srv := server.New(db, server.DefaultLimits())

	fmt.Printf("\n  Fleet — Self-hosted vehicle and fleet management\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Questions? hello@stockyard.dev — I read every message\n\n", port, port)
	log.Printf("fleet: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
