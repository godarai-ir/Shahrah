package main

import (
	"log"
	"net/http"
	"os"
	"time"

	appserver "github.com/Delpak/shahrah/services/tunnel-server/pkg/server"
)

func main() {
	silkRoadURL := getenv("SILK_ROAD_URL", "http://127.0.0.1:8083")
	addr := getenv("TUNNEL_SERVER_ADDR", ":8082")
	client := &http.Client{Timeout: 5 * time.Minute}

	log.Printf("tunnel-server listening on %s -> %s", addr, silkRoadURL)
	log.Fatal(http.ListenAndServe(addr, appserver.NewHandler(silkRoadURL, client)))
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
