// internal/logging/bidrequest_log.go
package logging

import "time"

type BidRequestLog struct {
	Timestamp string `json:"ts"`

	AuctionID string `json:"auction_id"`
	SeatID    string `json:"seat_id"`

	Imp struct {
		ID       string  `json:"id"`
		Format   string  `json:"fmt"`
		Width    int     `json:"w"`
		Height   int     `json:"h"`
		FloorCPM float64 `json:"floor_cpm"`
	} `json:"imp"`

	Device struct {
		OS         string `json:"os"`
		DeviceType string `json:"devicetype"`
	} `json:"device"`

	Geo  string `json:"geo"`
	Site struct {
		Domain string `json:"domain"`
		ID     string `json:"id"`
	} `json:"site"`

	User struct {
		ID       string `json:"id"`
		BuyerUID string `json:"buyeruid"`
	} `json:"user"`
}

// Helper to set timestamp to now if not provided
func NewBidRequestLog() BidRequestLog {
	return BidRequestLog{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
}