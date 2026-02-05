# RTB Bidder Simulation Pipeline

### End-to-End Synthetic Data Generation + Logging Infrastructure

This repository implements a complete RTB (Real-Time Bidding) simulation environment for testing and data collection:

* **Go-based OpenRTB bidder** (local, no bidding logic yet)
* **Protobuf decoding** for realistic RTB requests
* **Kinesis Firehose → S3** log ingestion pipeline
* **Synthetic OpenRTB traffic generator** (replaces real exchange)
* **Terraform-managed AWS infrastructure**
* Clear path toward production deployment with bidding logic

This system is currently focused on **synthetic data simulation** to test the complete logging pipeline from request generation to S3 storage.

---

## ⭐ High-Level Overview

### End-to-End RTB Simulation + Logging Pipeline

Your system is composed of four major stages, each responsible for part of the overall lifecycle:

```
[1. Synthetic Loadgen]  →  [2. Go Bidder]  →  [3. Firehose]  →  [4. S3 Data Lake]
   (Fake Exchange)         (Local/No-Bid)      (Streaming)       (ML Training)
```

---

## 1. Synthetic Traffic Generator (Loadgen)

### Purpose
Pretend to be Google and generate realistic RTB requests.

The synthetic generator (`cmd/loadgen`) plays the role of an exchange or Google Authorized Buyers.

### What it does

* Constructs valid `BidRequest` protobuf messages with:
  * `imp` (impressions)
  * `device` (OS, type, user agent)
  * `geo` (country, city)
  * `site` (domain, page)
  * `user` (ID)
  * floors, sizes, etc.
* Encodes via `proto.Marshal`
* POSTs to your bidder endpoint:
  ```
  POST http://localhost:8080/openrtb
  ```
* At controlled rate (QPS) and volume (N requests)

### Benefits

* ✅ Safe, cost-free, realistic data feed
* ✅ Stress-test your bidder logic
* ✅ Realistic distributions (domains, OS, country, floors)
* ✅ No reliance on a Google RTB seat yet

**Think of it as synthetic OpenRTB traffic that behaves like a miniature exchange.**

---

## 2. Your Go Bidder (`/openrtb`)

### Purpose
Receive bid requests, decode them, log them, and always return 204 (no-bid).

### Current Implementation

The bidder is **hosted locally** without bidding logic. This is purely for testing the data pipeline.

### What it does

1. Accepts HTTP POST requests on `/openrtb`
2. Reads binary protobuf body
3. `proto.Unmarshal` → `pb.BidRequest`
4. **Always returns 204 No Content** (no bidding logic)
5. Writes normalized logs via Firehose logger:
   * `auction_id`
   * `timestamp`
   * `device`
   * `geo`
   * `floor`
   * `size`
   * `user_id`
   * `site_domain`

### Benefits

* ✅ Ready-to-deploy RTB bidder interface
* ✅ Fully correct protobuf decoding pipeline
* ✅ Safe "no-bid" behavior while learning/testing
* ✅ Real normalized data for downstream ML/simulation

**This is the exact front end your real bidder will use.**

### Future: EC2 Deployment with Bidding Logic

In the future, the bidder will be:
* Hosted on **EC2** (not local)
* Include **real bidding logic** (bid price calculation, strategy)
* Return **BidResponse** messages
* Scale horizontally with load balancing

---

## 3. Firehose Logger (Go → AWS Kinesis Firehose)

### Purpose
Reliably stream logs from your bidder to S3.

### How it works

1. **JSON marshalling** of your normalized log struct
2. Sends to Firehose via AWS SDK (buffered mode)
3. Firehose:
   * Buffers by size/time
   * Compresses (GZIP)
   * Writes to S3 organized by:
     ```
     dt=YYYY-MM-DD/
     hr=HH/
     ```

### Benefits

* ✅ Zero data loss
* ✅ Automatic retry/backoff
* ✅ Cost-efficient streaming ingestion
* ✅ Perfect for high-throughput RTB workloads
* ✅ Ideal training data format for ML systems

**This is your core data ingestion pipeline.**

---

## 4. S3 Storage Lake (Raw RTB Logs Over Time)

### Purpose
Store all impression-level features for analysis, ML training, and simulation.

### What gets stored

Once Firehose flushes, your S3 bucket contains data like:

```
s3://qc-dsp-bidreq-logs/bid_requests/dt=2026-02-04/hr=18/part-0001.json.gz
```

Each file is:
* **Gzipped NDJSON** (newline-delimited JSON)
* **Partitioned by time** (date/hour)
* **Indexable** by Glue/Athena
* **Machine learning ready**

### Use Cases

#### ML / RL / Bandit Training
* Learn CTR models
* Floor optimization
* Scene distribution modeling
* Value estimation

#### Simulation / Replay Environment
Replays historical distribution of:
* Geo
* Device
* Floor
* Inventory
* Traffic patterns

Tests bidding policies **offline before going live**:
* ✅ Enables safe experimentation (no spend)
* ✅ Validates strategy changes
* ✅ Optimizes before production

**This is your source of truth for everything downstream.**

---

## 🎯 Current System Focus

This version is for **synthetic data simulation** to test sinking logs to S3:

* ✅ **Loadgen** generates fake but realistic RTB traffic
* ✅ **Go bidder** runs locally without bidding logic
* ✅ **Firehose** streams data reliably
* ✅ **S3** stores partitioned logs for analysis

### What's NOT included yet:
* ❌ Real bidding logic (always returns 204)
* ❌ EC2 deployment (bidder runs locally)
* ❌ Real exchange integration
* ❌ BidResponse generation

---

## 🚀 Getting Started

### Prerequisites

```bash
# Install Go
brew install go

# Install Protobuf Compiler
brew install protobuf

# Install AWS CLI
brew install awscli
aws configure

# Install Terraform
brew tap hashicorp/tap
brew install hashicorp/tap/terraform
```

---

### Step 1: Build the Bidder 

```bash
go mod tidy  
go build ./...
```

Set environment variables and Run locally:

```bash
source ./source_env.sh
go run ./cmd/bidder
```

Check health:

```bash
curl -v http://localhost:8080/health
```

---

### Step 2: Run the Synthetic Loadgen

```bash

# Optional: override defaults
export BIDDER_URL=http://localhost:8080/openrtb
export LOADGEN_NUM_REQUESTS=50

go run ./cmd/loadgen
```

Example load test @ 200 QPS for 10 seconds:

```bash
./loadgen \
  -endpoint=http://localhost:8080/openrtb \
  -qps=200 \
  -duration=10s
```

Example high-volume test @ 2000 QPS:

```bash
./loadgen \
  -endpoint=http://localhost:8080/openrtb \
  -qps=2000 \
  -duration=30s
```

---

### Step 3: Deploy AWS Infrastructure (Terraform)

```bash
cd infra/terraform
terraform init
terraform plan
terraform apply
```

This provisions:
* Firehose delivery stream
* S3 bucket with partitioning
* IAM roles/policies
* CloudWatch log groups

---

### Step 4: Verify Data in S3

After running loadgen, check your S3 bucket:

```bash
aws s3 ls s3://$S3_BUCKET_NAME/bid_requests/ --recursive
```

Expected output:
```
2026-02-04 18:23:45  123456  bid_requests/dt=2026-02-04/hr=18/part-0001.json.gz
2026-02-04 18:28:12  234567  bid_requests/dt=2026-02-04/hr=18/part-0002.json.gz
```

Download and inspect:

```bash
aws s3 cp s3://qc-dsp-bidreq-logs/bid_requests/dt=2026-02-04/hr=18/part-0001.json.gz - | gunzip
```

---

## 📊 Architecture Diagram

```
┌─────────────────────┐
│ Synthetic Loadgen   │  Generates realistic BidRequest protobufs
│ (cmd/loadgen)       │  at controlled QPS
└──────────┬──────────┘
           │ HTTP POST (protobuf binary)
           ▼
┌─────────────────────┐
│ Go Bidder (Local)   │  Decodes protobuf
│ /openrtb endpoint   │  Returns 204 (no-bid)
│                     │  Extracts features
└──────────┬──────────┘
           │ Normalized JSON logs
           ▼
┌─────────────────────┐
│ Kinesis Firehose    │  Buffers, compresses, retries
│                     │  Partitions by dt/hr
└──────────┬──────────┘
           │ Batched, gzipped writes
           ▼
┌─────────────────────┐
│ S3 Data Lake        │  Long-term storage
│ /bid_requests/      │  Queryable via Athena
│  dt=YYYY-MM-DD/     │  ML training ready
│   hr=HH/            │
└─────────────────────┘
           │
           ▼
┌─────────────────────┐
│ Future Use Cases    │
│ • ML Training       │
│ • Simulation        │
│ • Analytics         │
│ • Optimization      │
└─────────────────────┘
```

---

## 🔮 Future Roadmap

### Phase 1: Current (Testing Pipeline)
- [x] Synthetic traffic generator
- [x] Local bidder (no-bid only)
- [x] Firehose → S3 pipeline
- [x] Terraform infrastructure

### Phase 2: Production Bidder
- [ ] Deploy bidder to EC2
- [ ] Implement bidding logic
- [ ] Generate BidResponse messages
- [ ] Add pacing + budget controls
- [ ] Load balancing (ALB/NLB)
- [ ] Auto-scaling groups

### Phase 3: Real Exchange Integration
- [ ] Google Authorized Buyers integration
- [ ] Real-time bidding
- [ ] Win/loss notifications
- [ ] Billing reconciliation

### Phase 4: ML + Optimization
- [ ] CTR/CVR prediction models
- [ ] Floor price optimization
- [ ] Value estimation
- [ ] Replay-based simulation environment

---

## ✅ Summary

You now have:

* ✅ A **synthetic OpenRTB exchange** that generates realistic traffic
* ✅ A **Go bidder** that correctly decodes protobuf and logs data
* ✅ A **Firehose → S3 pipeline** for reliable data ingestion
* ✅ **Terraform infrastructure** for reproducible deployments
* ✅ A foundation for **real bidding logic and EC2 deployment**

This system validates your entire logging pipeline end-to-end using synthetic data, preparing you for production deployment.