package core

import "time"

// AuctionContext is a normalized view of an impression opportunity.
// It is filled from either:
//   - Google OpenRTB BidRequest (prod)
//   - Synthetic generator / log replay (sim)
type AuctionContext struct {
	AuctionID string
	Timestamp time.Time

	// Basic context
	Geo          string
	DeviceType   string
	OS           string
	AppBundle    string
	SiteDomain   string
	InventoryID  string // publisher slot / placement id
	InventoryFmt string // "banner", "video", "native"
	Width        int
	Height       int

	// Economics / constraints
	FloorCPM float64 // USD CPM floor if known

	// User / privacy
	HasUserID bool
	UserID    string
	RegGDPR   bool
	RegCOPPA  bool
	RegCCPA   bool
	// etc: consent strings, segments, etc.
}

// CampaignState is the current state of a campaign or strategy.
// In production, this can be hydrated from Redis/DB/etc.
type CampaignState struct {
	ID           string
	DailyBudget  float64
	SpentToday   float64
	TotalBudget  float64
	TotalSpent   float64
	TargetCPA    float64
	TargetROAS   float64
	IsActive     bool
	Segments     []string
	AllowedGeos  []string
	AllowedSites []string
}

// BidDecision is the output of your bidding logic.
type BidDecision struct {
	ShouldBid bool
	// CPM in USD you want to bid.
	BidCPM float64

	// Which creative to serve if you win
	CreativeID string
	// Landing page domain
	AdvertiserDomain string

	// Optional: per-imp / strategy metadata for logging/debug
	Reason string
}