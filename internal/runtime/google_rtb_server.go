package runtime
// This is your bridge between Google OpenRTB Protobuf and your core engine.
// Production adapter (Google OpenRTB server)
// This is your Authorized Buyers HTTP endpoint that:
// Receives binary Protobuf BidRequest
// Converts it into core.AuctionContext
// Calls Engine.EvaluateAuction
// Builds BidResponse and returns it

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/yourname/openrtb-framework/internal/core"

	pb "github.com/yourname/openrtb-framework/proto/openrtb"
	"google.golang.org/protobuf/proto"
)

type GoogleRTBServer struct {
	engine  *core.Engine
	metrics *core.Metrics

	seatID            string
	defaultCampaignID string
}

// NewGoogleRTBServer builds the HTTP handler for Google OpenRTB traffic.
func NewGoogleRTBServer(engine *core.Engine, metrics *core.Metrics, seatID, campaignID string) *GoogleRTBServer {
	return &GoogleRTBServer{
		engine:            engine,
		metrics:           metrics,
		seatID:            seatID,
		defaultCampaignID: campaignID,
	}
}

// Handler is the HTTP endpoint Google will call for bid requests.
func (s *GoogleRTBServer) Handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	s.metrics.IncAuctions()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		s.metrics.IncErrors()
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("read body error: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		s.metrics.IncErrors()
		return
	}
	defer r.Body.Close()

	req := &pb.BidRequest{}
	if err := proto.Unmarshal(body, req); err != nil {
		log.Printf("unmarshal BidRequest error: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		s.metrics.IncErrors()
		return
	}

	if len(req.GetImp()) == 0 {
		// No impressions → no bid
		w.WriteHeader(http.StatusNoContent)
		s.metrics.ObserveLatency(time.Since(start))
		return
	}

	// For simplicity, just handle the first impression.
	imp := req.GetImp()[0]
	ac := s.toAuctionContext(req, imp)

	ctx := r.Context()
	dec, err := s.engine.EvaluateAuction(ctx, s.defaultCampaignID, ac)
	if err != nil {
		log.Printf("engine evaluation error: %v", err)
		s.metrics.IncErrors()
		w.WriteHeader(http.StatusNoContent)
		s.metrics.ObserveLatency(time.Since(start))
		return
	}

	if !dec.ShouldBid {
		w.WriteHeader(http.StatusNoContent)
		s.metrics.ObserveLatency(time.Since(start))
		return
	}

	s.metrics.IncBids()

	resp := s.buildBidResponse(req, imp, dec)

	out, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("marshal BidResponse error: %v", err)
		s.metrics.IncErrors()
		w.WriteHeader(http.StatusNoContent)
		s.metrics.ObserveLatency(time.Since(start))
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(out); err != nil {
		log.Printf("write error: %v", err)
		s.metrics.IncErrors()
	}

	elapsed := time.Since(start)
	s.metrics.ObserveLatency(elapsed)

	log.Printf(
		"auction=%s imp=%s bidCPM=%.4f reason=%s latency=%s",
		req.GetId(),
		imp.GetId(),
		dec.BidCPM,
		dec.Reason,
		elapsed,
	)
}

// toAuctionContext maps a Google OpenRTB BidRequest + Imp to your internal AuctionContext.
// This is where you normalize Google-specific fields into your generic model.
func (s *GoogleRTBServer) toAuctionContext(req *pb.BidRequest, imp *pb.BidRequest_Imp) core.AuctionContext {
	ac := core.AuctionContext{
		AuctionID: req.GetId(),
		Timestamp: time.Now(), // you can add more exact timing later
		FloorCPM:  imp.GetBidfloor(),
	}

	// Device
	if d := req.GetDevice(); d != nil {
		ac.OS = d.GetOs()
		// Device type (phones, tablets, etc.) could be mapped from d.Devicetype
		// Example (pseudo):
		// switch d.GetDevicetype() { case 1: ac.DeviceType = "mobile"; ... }
	}

	// Geo: you can use device.geo or user.geo depending on what's populated
	if d := req.GetDevice(); d != nil && d.GetGeo() != nil {
		// Use country as simple geo for now
		ac.Geo = d.GetGeo().GetCountry()
	}

	// Site / App
	if site := req.GetSite(); site != nil {
		ac.SiteDomain = site.GetDomain()
		if ac.InventoryID == "" {
			ac.InventoryID = site.GetId()
		}
	}
	if app := req.GetApp(); app != nil {
		ac.AppBundle = app.GetBundle()
		if ac.InventoryID == "" {
			ac.InventoryID = app.GetId()
		}
	}

	// Format / size
	if b := imp.GetBanner(); b != nil {
		ac.InventoryFmt = "banner"
		if len(b.GetFormat()) > 0 {
			f := b.GetFormat()[0]
			ac.Width = int(f.GetW())
			ac.Height = int(f.GetH())
		} else {
			ac.Width = int(b.GetW())
			ac.Height = int(b.GetH())
		}
	} else if v := imp.GetVideo(); v != nil {
		ac.InventoryFmt = "video"
		ac.Width = int(v.GetW())
		ac.Height = int(v.GetH())
	} else if n := imp.GetNative(); n != nil {
		_ = n
		ac.InventoryFmt = "native"
		// Sizes for native are more abstract; you can add mapping later.
	}

	// User
	if u := req.GetUser(); u != nil {
		if u.GetId() != "" {
			ac.HasUserID = true
			ac.UserID = u.GetId()
		}
	}

	// Regulations / privacy (stubs for now)
	if regs := req.GetRegs(); regs != nil {
		_ = regs
		// Example: populate GDPR/CCPA flags after parsing regs.ext
	}

	return ac
}

// buildBidResponse maps your BidDecision back to a Google OpenRTB BidResponse.
func (s *GoogleRTBServer) buildBidResponse(
	req *pb.BidRequest,
	imp *pb.BidRequest_Imp,
	dec core.BidDecision,
) *pb.BidResponse {
	bid := &pb.BidResponse_SeatBid_Bid{
		Impid:   imp.GetId(),
		Price:   dec.BidCPM, // CPM in USD
		Crid:    dec.CreativeID,
		Adomain: []string{dec.AdvertiserDomain},
		// TODO:
		// - adm (ad markup) OR
		// - adid + use a pre-registered creative
		// - nurl/burl/impression_tracking_url for win + impression tracking
	}

	seatBid := &pb.BidResponse_SeatBid{
		Seat: s.seatID,
		Bid:  []*pb.BidResponse_SeatBid_Bid{bid},
	}

	return &pb.BidResponse{
		Id:      req.GetId(),
		Seatbid: []*pb.BidResponse_SeatBid{seatBid},
		// Optionally set currency, deal details, ext, etc.
	}
}