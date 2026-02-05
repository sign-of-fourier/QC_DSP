// cmd/bidder/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/sign-of-fourier/QC_DSP/internal/logging"
	"github.com/sign-of-fourier/QC_DSP/internal/runtime"
)

func main() {
	port := os.Getenv("BIDDER_PORT")
	if port == "" {
		port = "8080"
	}
	seatID := os.Getenv("BIDDER_SEAT_ID")
	if seatID == "" {
		seatID = "test-seat"
	}

	ctx := context.Background()
	firehoseLogger := logging.NewFirehoseLogger(ctx)

	server := runtime.NewGoogleRTBServer(firehoseLogger, seatID)

	mux := http.NewServeMux()
	mux.HandleFunc("/openrtb", server.Handler)
	mux.HandleFunc("/test-log", server.TestLogHandler)

	addr := ":" + port
	log.Printf("Starting bidder on %s, seat=%s, firehose=%s",
		addr, seatID, os.Getenv("FIREHOSE_STREAM_NAME"))

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}