// internal/runtime/google_rtb_server.go
package runtime

import (
	"io"
	"log"
	"net/http"

	"github.com/sign-of-fourier/QC_DSP/internal/logging"
	pb "github.com/sign-of-fourier/QC_DSP/proto"
	"google.golang.org/protobuf/proto"
)

type GoogleRTBServer struct {
	firehose *logging.FirehoseLogger
	seatID   string
}

func NewGoogleRTBServer(fh *logging.FirehoseLogger, seatID string) *GoogleRTBServer {
	return &GoogleRTBServer{
		firehose: fh,
		seatID:   seatID,
	}
}

// Handler: main OpenRTB endpoint for Google Authorized Buyers
func (s *GoogleRTBServer) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("read body error: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	req := &pb.BidRequest{}
	if err := proto.Unmarshal(body, req); err != nil {
		log.Printf("unmarshal BidRequest error: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if len(req.GetImp()) == 0 {
		// No impressions – just no-bid
		w.WriteHeader(http.StatusNoContent)
		return
	}
	imp := req.GetImp()[0]

	// Build log record
	logRec := s.buildBidRequestLog(req, imp)
	partitionKey := req.GetId()
	s.firehose.Log(r.Context(), partitionKey, logRec)

	// For now, always no-bid (204)
	w.WriteHeader(http.StatusNoContent)
}

// Optional debug endpoint to test Firehose without real RTB
func (s *GoogleRTBServer) TestLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rec := logging.NewBidRequestLog()
	rec.AuctionID = "test-auction-id"
	rec.SeatID = s.seatID
	rec.Imp.ID = "1"
	rec.Imp.Format = "banner"
	rec.Imp.Width = 300
	rec.Imp.Height = 250
	rec.Imp.FloorCPM = 1.23
	rec.Device.OS = "iOS"
	rec.Device.DeviceType = "mobile"
	rec.Geo = "US-MA"
	rec.Site.Domain = "example.com"
	rec.Site.ID = "site-123"
	rec.User.ID = "user-abc"
	rec.User.BuyerUID = "buyeruid-xyz"

	s.firehose.Log(r.Context(), rec.AuctionID, rec)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("test log sent to Firehose\n"))
}

func (s *GoogleRTBServer) buildBidRequestLog(
	req *pb.BidRequest,
	imp *pb.BidRequest_Imp,
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