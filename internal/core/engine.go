package core
// Orchestrator that wires strategy + campaign state
import (
	"context"
	"errors"
)

// Engine holds dependencies: strategy, campaign store, models, etc.
type Engine struct {
	strategy      Strategy
	campaignStore CampaignStore
}

// CampaignStore abstracts how you get/update campaign state.
// In production this may hit Redis/DB; in simulation it's in-memory.
type CampaignStore interface {
	GetCampaign(ctx context.Context, id string) (CampaignState, error)
	UpdateSpend(ctx context.Context, id string, delta float64) error
}

func NewEngine(strategy Strategy, store CampaignStore) *Engine {
	return &Engine{
		strategy:      strategy,
		campaignStore: store,
	}
}

// EvaluateAuction is the central function used by both prod and sim.
func (e *Engine) EvaluateAuction(ctx context.Context, campaignID string, ac AuctionContext) (BidDecision, error) {
	cs, err := e.campaignStore.GetCampaign(ctx, campaignID)
	if err != nil {
		return BidDecision{}, err
	}
	if !cs.IsActive {
		return BidDecision{ShouldBid: false, Reason: "campaign_inactive"}, nil
	}
	if cs.SpentToday >= cs.DailyBudget {
		return BidDecision{ShouldBid: false, Reason: "daily_budget_exhausted"}, nil
	}

	dec, err := e.strategy.DecideBid(ctx, ac, cs)
	if err != nil {
		return BidDecision{}, err
	}
	if !dec.ShouldBid {
		return dec, nil
	}

	// Optional safety check: don't bid negative or zero
	if dec.BidCPM <= 0 {
		dec.ShouldBid = false
		dec.Reason = "non_positive_bid"
		return dec, nil
	}

	// NOTE: cost update should happen only on win (in prod) or win simulation (in sim).
	// So we don't adjust spend here — we just decide.
	return dec, nil
}
