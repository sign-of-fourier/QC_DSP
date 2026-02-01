// cmd/bidder/main.go - production bidder
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/yourname/qc-dsp/internal/config"
	"github.com/yourname/qc-dsp/internal/core"
	"github.com/yourname/qc-dsp/internal/logging"
	"github.com/yourname/qc-dsp/internal/runtime"
)

func main() {
	cfg := config.Load()

	// Simple campaign store for now
	store := core.NewInMemoryCampaignStore(map[string]core.CampaignState{
		cfg.DefaultCampaignID: {
			ID:          cfg.DefaultCampaignID,
			DailyBudget: 100.0,
			TotalBudget: 1000.0,
			IsActive:    true,
			TargetCPA:   20.0,
		},
	})

	strategy := core.NewSimpleValueStrategy()
	engine := core.NewEngine(strategy, store)
	metrics := core.NewMetrics()

	ctx := context.Background()
	firehoseLogger := logging.NewFirehoseLogger(ctx)

	server := runtime.NewGoogleRTBServer(
		engine,
		metrics,
		firehoseLogger,
		cfg.SeatID,
		cfg.DefaultCampaignID,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/openrtb", server.Handler)

	addr := ":" + cfg.Port
	log.Printf("Starting bidder on %s, seat=%s, firehose=%s",
		addr, cfg.SeatID, os.Getenv("FIREHOSE_STREAM_NAME"))

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}