// internal/config/config.go
// Simple env-based config with sane defaults.
package config

import (
	"log"
	"os"
)

type Config struct {
	// HTTP port the bidder listens on, e.g. "8080"
	Port string

	// Authorized Buyers seat ID you’re bidding as
	SeatID string

	// Default campaign ID used by this bidder instance
	DefaultCampaignID string

	// Optional log level, if you want to use it later
	LogLevel string
}

func getenvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	port := getenvOrDefault("BIDDER_PORT", "8080")
	seat := getenvOrDefault("BIDDER_SEAT_ID", "1234")
	campaign := getenvOrDefault("BIDDER_DEFAULT_CAMPAIGN_ID", "default")
	logLevel := getenvOrDefault("BIDDER_LOG_LEVEL", "INFO")

	if seat == "1234" {
		log.Println("WARN: BIDDER_SEAT_ID not set, using placeholder '1234'")
	}

	return Config{
		Port:              port,
		SeatID:            seat,
		DefaultCampaignID: campaign,
		LogLevel:          logLevel,
	}
}
