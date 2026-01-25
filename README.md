# Quante Carlo's DCO Engine
Optimizes Ads in real-time.
More clicks in fewer experiements
<img src="AdLearning.png"></img>
=======
# OpenRTB Bidder Framework

A Go framework for building a **Google Authorized Buyers (AdX) OpenRTB bidder** that also supports **offline simulation** using the *same* bidding logic.

- ✅ **Production mode:** Handle real-time Google OpenRTB Protobuf bid requests and respond in under ~100ms.  
- ✅ **Simulation mode:** Run high-volume synthetic auctions for experimentation, budgeting, and ML strategy testing.  
- ✅ **Unified bidding logic:** Same engine and strategy used in both prod and sim—no duplicated logic.

## Bayesian Optimization
- GPR: optimization of ads based on semantic content
- Massively scalable batch Expected Imprtovement
- Warping function

(More Details)[data.md]
=======
---

## 📐 Architecture Overview

This project uses a **hexagonal (ports-and-adapters)** architecture to separate:

### **Core Domain (shared)**
Located in: `internal/core`

Contains the business logic:

- Auction normalization (`AuctionContext`)
- Campaign budgeting & pacing (`CampaignState`)
- Bidding strategy interface (`Strategy`)
- The `Engine` that orchestrates decisions
- Metrics (`Metrics`)
- Campaign storage (`CampaignStore`)

### **Production Adapter**
Located in: `internal/runtime/google_rtb_server.go`

- Receives **Google Authorized Buyers** Protobuf `BidRequest`
- Converts to `AuctionContext`
- Uses the shared `Engine` to compute a bid
- Builds a Protobuf `BidResponse` and returns it to Google

### **Simulation Adapter**
Located in: `internal/sim`

- Generates synthetic auctions (`Generator`)
- Runs them through the same engine (`Runner`)
- Computes wins, spend, CTR/CVR outcomes

---

## 📁 Directory Layout
openrtb-framework/
├─ go.mod
├─ proto/
│ ├─ openrtb.proto # Google’s main OpenRTB schema
│ ├─ openrtb-adx.proto # Google-specific extensions
│ └─ openrtb/ # Generated Go protobuf code
│ ├─ openrtb.pb.go
│ └─ openrtb_adx.pb.go
├─ cmd/
│ ├─ bidder/ # Production bidder entrypoint
│ │ └─ main.go
│ └─ simulate/ # Offline simulator CLI
│ └─ main.go
└─ internal/
├─ core/ # Shared domain logic
│ ├─ types.go
│ ├─ strategy.go
│ ├─ strategy_simple.go
│ ├─ engine.go
│ ├─ metrics.go
│ └─ store_inmemory.go # In-memory CampaignStore
├─ runtime/ # Production OpenRTB server
│ └─ google_rtb_server.go
├─ sim/ # Simulation engine
│ ├─ generator.go
│ └─ runner.go
└─ config/
└─ config.go

## 🚀 Getting Started

### **Prerequisites**

- Go **1.22+**
- `protoc` (Protocol Buffer compiler)
- Protobuf plugins:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

1. Clone the repo
```bash
git clone https://github.com/yourname/openrtb-framework.git
cd openrtb-framework
go mod tidy
```
📦 Download and Generate OpenRTB Protos
1. Download required files from Google

Download from the Google Authorized Buyers documentation:

openrtb.proto

openrtb-adx.proto

Place them inside:
./proto/

2. Generate Go protobuf code

From inside proto/:
```bash
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  openrtb.proto openrtb-adx.proto
```

This generates:

proto/openrtb.pb.go

proto/openrtb-adx.pb.go

These files contain the native Go structs used to decode/encode Protobuf RTB messages.

🏭 Running the Bidder (Production Mode)

The production bidder listens for Google OpenRTB requests.

Set environment variables
```bash
export BIDDER_PORT=8080
export BIDDER_SEAT_ID=YOUR_SEAT_ID
export BIDDER_DEFAULT_CAMPAIGN_ID=default
export BIDDER_LOG_LEVEL=INFO
```

Run
```bash
go run ./cmd/bidder
```

Your bidder now exposes:
```bash
POST http://localhost:8080/openrtb
Content-Type: application/octet-stream

<protobuf binary BidRequest>
```

And returns:

200 OK + Protobuf BidResponse on bid

204 No Content for no-bid

This endpoint is where Google Authorized Buyers will call your bidder.

🧪 Running the Simulator

Simulations allow offline experimentation using the exact same bidding engine.

```bash
go run ./cmd/simulate
```

This will:

Generate 100k synthetic auctions

Run them through your bidding engine

Simulate clearing prices & outcomes

Print a summary (spend, impressions, bids, etc.)

Update:

generator.go → auction distribution

runner.go → market logic

strategy_simple.go → your bidding logic

🧠 Core Concepts
AuctionContext

A normalized representation of an impression opportunity.

Built from:

Prod: Google OpenRTB BidRequest + Imp

Sim: Synthetic generator

Contains fields like:

Geo, device, OS

Site/app context

Ad slot format (banner/video/native)

Size

Floor CPM

User ID / privacy flags

Auction timestamp

CampaignState

Tracks your budget, pacing, and strategy state:

DailyBudget

SpentToday

TotalBudget

TargetCPA / TargetROAS

Allowed geos / sites / segments

Active/inactive

Strategy & Engine

Strategy decides how much to bid

Engine applies:

Campaign state

Budget checks

Strategy logic

Returns a BidDecision

Both prod and sim call the same Engine.EvaluateAuction.

🔧 Where to Plug in ML, Budgets, and Real Logic

Here’s where the "real DSP logic" lives:

1. ML Models

Replace strategy_simple.go with:

gRPC call to Python model server

TensorFlow/Caffe model embedded in Go

Value bidding using CTR × CVR × value

2. Campaign Storage

Replace store_inmemory.go with:

Redis

PostgreSQL

DynamoDB

Bigtable

3. Creative Mapping

Use your own DB:

Creative IDs

Landing pages

Creative attributes

Pre-approved creatives for Google AdX

4. BidResponse Extensions

Build real markup:

adm

nurl / burl

ext.impression_tracking_url

Use ${AUCTION_PRICE} macro for win notifications

5. Full Observability

Extend metrics.go and expose:

Prometheus metrics

Log structured events

Latency histograms

Bid/no-bid reasons