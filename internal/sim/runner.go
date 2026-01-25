package sim

import (
	"context"
	"math/rand"

	"github.com/yourname/openrtb-framework/internal/core"
)

type Result struct {
	Auctions    int
	Bids        int
	Impressions int
	Spend       float64
}

type Runner struct {
	engine       *core.Engine
	campaignID   string
	rand         *rand.Rand
	ctrBaseline  float64
	cvrBaseline  float64
}

func NewRunner(engine *core.Engine, campaignID string) *Runner {
	return &Runner{
		engine:      engine,
		campaignID:  campaignID,
		rand:        rand.New(rand.NewSource(42)),
		ctrBaseline: 0.02,
		cvrBaseline: 0.05,
	}
}

func (r *Runner) Run(ctx context.Context, auctions []core.AuctionContext) (Result, error) {
	res := Result{}
	for _, ac := range auctions {
		res.Auctions++

		dec, err := r.engine.EvaluateAuction(ctx, r.campaignID, ac)
		if err != nil {
			return res, err
		}
		if !dec.ShouldBid {
			continue
		}
		res.Bids++

		// Simple synthetic market: clearing price around floor+0.5
		clearingCPM := ac.FloorCPM + r.rand.Float64()*1.0

		if dec.BidCPM >= clearingCPM {
			// Win
			res.Impressions++
			cost := clearingCPM / 1000.0
			_ = cost // you could call UpdateSpend on the CampaignStore here

			res.Spend += cost

			// Sample click
			if r.rand.Float64() < r.ctrBaseline {
				// Sample conversion
				if r.rand.Float64() < r.cvrBaseline {
					// Update some conversion metric if desired
				}
			}
		}
	}
	return res, nil
}