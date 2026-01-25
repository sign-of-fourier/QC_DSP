package sim
// Simulation adapter
// This lets you reuse the same Engine & Strategy with synthetic auctions.
import (
	"math/rand"
	"time"

	"github.com/yourname/openrtb-framework/internal/core"
)

type Generator struct {
	r *rand.Rand
}

func NewGenerator(seed int64) *Generator {
	return &Generator{r: rand.New(rand.NewSource(seed))}
}

func (g *Generator) SampleAuction(id int) core.AuctionContext {
	// Super simple synthetic distribution; you’ll refine this
	geo := "US"
	if g.r.Float64() < 0.3 {
		geo = "EU"
	}

	device := "mobile"
	if g.r.Float64() < 0.4 {
		device = "desktop"
	}

	floor := 0.5 + g.r.Float64()*2.0 // $0.5–$2.5 CPM

	return core.AuctionContext{
		AuctionID:   formatAuctionID(id),
		Timestamp:   time.Now(),
		Geo:         geo,
		DeviceType:  device,
		InventoryID: "slot-1",
		InventoryFmt: "banner",
		Width:       300,
		Height:      250,
		FloorCPM:    floor,
	}
}

func formatAuctionID(i int) string {
	return "sim-auction-" + strconv.Itoa(i)
}
