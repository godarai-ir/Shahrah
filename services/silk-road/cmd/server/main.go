package main

import (
	"log"
	"net/http"
	"os"
	"time"

	appserver "github.com/Delpak/shahrah/services/silk-road/pkg/server"
)

func main() {
	addr := getenv("SILK_ROAD_ADDR", ":8083")
	client := &http.Client{Timeout: 5 * time.Minute}

	log.Printf("silk-road listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, appserver.NewHandler(client)))
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
