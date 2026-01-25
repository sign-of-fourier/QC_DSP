package main
// running the prod server

import (
	"log"
	"net/http"

	"github.com/yourname/openrtb-framework/internal/config"
	"github.com/yourname/openrtb-framework/internal/core"
	"github.com/yourname/openrtb-framework/internal/runtime"
)

func main() {
	cfg := config.Load()

	// Simple in-memory campaign store (you’d implement this)
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

	server := runtime.NewGoogleRTBServer(engine, metrics, cfg.SeatID, cfg.DefaultCampaignID)

	mux := http.NewServeMux()
	mux.HandleFunc("/openrtb", server.Handler)

	addr := ":" + cfg.Port
	log.Printf("Starting Google OpenRTB bidder on %s seat=%s campaign=%s",
		addr, cfg.SeatID, cfg.DefaultCampaignID)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
