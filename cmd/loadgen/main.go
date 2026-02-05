package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	pb "github.com/sign-of-fourier/QC_DSP/proto"
	"google.golang.org/protobuf/proto"
)

func main() {
	// Target bidder URL, default to local bidder
	targetURL := os.Getenv("BIDDER_URL")
	if targetURL == "" {
		targetURL = "http://localhost:8080/openrtb"
	}

	// Number of requests to send (can override via env)
	numRequests := 50
	if nStr := os.Getenv("LOADGEN_NUM_REQUESTS"); nStr != "" {
		var n int
		fmt.Sscanf(nStr, "%d", &n)
		if n > 0 {
			numRequests = n
		}
	}

	log.Printf("Loadgen sending %d synthetic BidRequests to %s", numRequests, targetURL)

	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	for i := 0; i < numRequests; i++ {
		reqMsg := makeSyntheticBidRequest()
		if err := sendBidRequest(ctx, client, targetURL, reqMsg); err != nil {
			log.Printf("request %d: error: %v", i, err)
		} else {
			log.Printf("request %d: sent ok (auction_id=%s)", i, reqMsg.GetId())
		}
		time.Sleep(50 * time.Millisecond) // small delay so we don't hammer
	}
}

func makeSyntheticBidRequest() *pb.BidRequest {
	// Simple synthetic OpenRTB BidRequest: just id + one banner imp.
	auctionID := fmt.Sprintf("auction-%d", time.Now().UnixNano())

	req := &pb.BidRequest{
		Id: proto.String(auctionID),

		Imp: []*pb.BidRequest_Imp{
			{
				Id:       proto.String("1"),
				Bidfloor: proto.Float64(0.5 + rand.Float64()*1.5), // 0.5 to 2.0 CPM
				Banner: &pb.BidRequest_Imp_Banner{
					W: proto.Int32(300),
					H: proto.Int32(250),
					Format: []*pb.BidRequest_Imp_Banner_Format{
						{
							W: proto.Int32(300),
							H: proto.Int32(250),
						},
					},
				},
			},
		},
	}

	// NOTE:
	// We are *not* setting Site, Device, User, etc. here because the exact
	// field names and types in Google's openrtb.proto/openrtb-adx.proto
	// can differ from generic examples.
	//
	// This minimal request is enough to:
	//   - be a valid BidRequest
	//   - exercise your /openrtb handler
	//   - get logged to Firehose → S3
	//
	// You can later enrich this by:
	//   - opening proto/openrtb.pb.go
	//   - looking at type BidRequest struct { ... }
	//   - adding fields using the exact names/types from that struct.

	return req
}


func randomDomain() string {
	domains := []string{
		"news.example.com",
		"sports.example.com",
		"travel.example.com",
		"tech.example.com",
		"finance.example.com",
	}
	return domains[rand.Intn(len(domains))]
}

func randomOS() string {
	oses := []string{
		"iOS",
		"Android",
		"Windows",
		"macOS",
	}
	return oses[rand.Intn(len(oses))]
}

func randomCountry() string {
	countries := []string{
		"US",
		"CA",
		"GB",
		"DE",
		"FR",
		"JP",
	}
	return countries[rand.Intn(len(countries))]
}

func sendBidRequest(ctx context.Context, client *http.Client, targetURL string, reqMsg *pb.BidRequest) error {
	data, err := proto.Marshal(reqMsg)
	if err != nil {
		return fmt.Errorf("proto marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	// For now we expect 204 No Content from our no-bid bidder
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}