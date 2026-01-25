package core
// The core “bidding brain” interface

import "context"

// Strategy is the core interface that BOTH prod and sim will use.
// You can implement multiple strategies and A/B test them.
type Strategy interface {
	// DecideBid returns a BidDecision for a given auction + campaign state.
	DecideBid(ctx context.Context, ac AuctionContext, cs CampaignState) (BidDecision, error)
}
