// a very simple baseline strategy
package core

import (
	"context"
	"math"
)

type SimpleValueStrategy struct{}

func NewSimpleValueStrategy() *SimpleValueStrategy {
	return &SimpleValueStrategy{}
}

func (s *SimpleValueStrategy) DecideBid(ctx context.Context, ac AuctionContext, cs CampaignState) (BidDecision, error) {
	// Very naive: just use TargetCPA and an assumed CVR
	// In reality you'll plug in ML models here.

	// Example: assume constant CVR per impression (toy)
	const assumedCVR = 0.02 // 2% conversion rate

	if cs.TargetCPA <= 0 {
		// If no target CPA, skip bidding
		return BidDecision{ShouldBid: false, Reason: "no_target_cpa"}, nil
	}

	valuePerConversion := cs.TargetCPA
	valuePerImpression := assumedCVR * valuePerConversion // in USD

	bidCPM := valuePerImpression * 1000.0

	// Apply some safety / margin
	bidCPM = bidCPM * 0.7

	// Don't bid below floor (if known)
	if ac.FloorCPM > 0 {
		bidCPM = math.Max(bidCPM, ac.FloorCPM)
	}

	if bidCPM <= 0 {
		return BidDecision{ShouldBid: false, Reason: "computed_zero_bid"}, nil
	}

	return BidDecision{
		ShouldBid:        true,
		BidCPM:           bidCPM,
		CreativeID:       "demo-creative-1",
		AdvertiserDomain: "example.com",
		Reason:           "simple_value_strategy",
	}, nil
}
