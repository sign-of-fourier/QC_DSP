// internal/runtime/google_rtb_server.go
package runtime

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/yourname/qc-dsp/internal/core"
	"github.com/yourname/qc-dsp/internal/logging"

	pb "github.com/yourname/qc-dsp/proto/openrtb"
	"google.golang.org/protobuf/proto"
)

type GoogleRTBServer struct {
	engine        *core.Engine
	metrics       *core.Metrics
	firehose      *logging.FirehoseLogger
	seatID        string
	defaultCampID string
}

func NewGoogleRTBServer(
	engine *core.Engine,
	metrics *core.Metrics,
	firehose *logging.FirehoseLogger,
	seatID string,
	defaultCampaignID string,
) *GoogleRTBServer {
	return &GoogleRTBServer{
		engine:        engine,
		metrics:       metrics,
		firehose:      firehose,
		seatID:        seatID,
		defaultCampID: defaultCampaignID,
	}
}

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
		w.WriteHeader(http.StatusNoContent)
		s.metrics.ObserveLatency(time.Since(start))
		return
	}

	imp := req.GetImp()[0]

	// Build and send log record to Firehose
	logRec := s.buildBidRequestLog(req, imp, start)
	partitionKey := req.GetId()
	s.firehose.Log(r.Context(), partitionKey, logRec)

	// Normal bidding flow
	ac := s.toAuctionContext(req, imp)
	dec, err := s.engine.EvaluateAuction(r.Context(), s.defaultCampID, ac)
	if err != nil {
		log.Printf("engine error: %v", err)
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
		log.Printf("write response error: %v", err)
		s.metrics.IncErrors()
	}

	elapsed := time.Since(start)
	s.metrics.ObserveLatency(elapsed)
	log.Printf("auction=%s bid=%.4f latency=%s", req.GetId(), dec.BidCPM, elapsed)
}

func (s *GoogleRTBServer) buildBidRequestLog(
	req *pb.BidRequest,
	imp *pb.BidRequest_Imp,
	now time.Time,
) logging.BidRequestLog {
	rec := logging.NewBidRequestLog()
	rec.AuctionID = req.GetId()
	rec.SeatID = s.seatID

	// Imp
	rec.Imp.ID = imp.GetId()
	rec.Imp.FloorCPM = imp.GetBidfloor()

	if b := imp.GetBanner(); b != nil {
		rec.Imp.Format = "banner"
		if len(b.GetFormat()) > 0 {
			f := b.GetFormat()[0]
			rec.Imp.Width = int(f.GetW())
			rec.Imp.Height = int(f.GetH())
		} else {
			rec.Imp.Width = int(b.GetW())
			rec.Imp.Height = int(b.GetH())
		}
	} else if v := imp.GetVideo(); v != nil {
		rec.Imp.Format = "video"
		rec.Imp.Width = int(v.GetW())
		rec.Imp.Height = int(v.GetH())
	} else if n := imp.GetNative(); n != nil {
		_ = n
		rec.Imp.Format = "native"
	}

	// Device & Geo
	if d := req.GetDevice(); d != nil {
		rec.Device.OS = d.GetOs()
		if d.GetGeo() != nil {
			rec.Geo = d.GetGeo().GetCountry()
		}
	}

	// Site / App
	if site := req.GetSite(); site != nil {
		rec.Site.Domain = site.GetDomain()
		rec.Site.ID = site.GetId()
	}
	if app := req.GetApp(); app != nil {
		if rec.Site.Domain == "" {
			rec.Site.Domain = app.GetBundle()
		}
		if rec.Site.ID == "" {
			rec.Site.ID = app.GetId()
		}
	}

	// User
	if u := req.GetUser(); u != nil {
		rec.User.ID = u.GetId()
		rec.User.BuyerUID = u.GetBuyeruid()
	}

	return rec
}