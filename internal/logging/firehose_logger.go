// internal/logging/firehose_logger.go
package logging

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
)

type FirehoseLogger struct {
	client     *firehose.Client
	streamName string
	async      bool
}

func NewFirehoseLogger(ctx context.Context) *FirehoseLogger {
	streamName := os.Getenv("FIREHOSE_STREAM_NAME")
	if streamName == "" {
		log.Println("FirehoseLogger: FIREHOSE_STREAM_NAME not set; logger disabled")
		return nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("FirehoseLogger: failed to load AWS config: %v; logger disabled", err)
		return nil
	}

	client := firehose.NewFromConfig(cfg)

	return &FirehoseLogger{
		client:     client,
		streamName: streamName,
		async:      true,
	}
}

func (fl *FirehoseLogger) Log(ctx context.Context, partitionKey string, payload any) {
	if fl == nil {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("FirehoseLogger: json marshal error: %v", err)
		return
	}

	input := &firehose.PutRecordInput{
		DeliveryStreamName: &fl.streamName,
		Record: &types.Record{
			Data: data,
		},
	}

	if fl.async {
		go fl.putRecord(context.Background(), input)
	} else {
		fl.putRecord(ctx, input)
	}
}

func (fl *FirehoseLogger) putRecord(ctx context.Context, input *firehose.PutRecordInput) {
	_, err := fl.client.PutRecord(ctx, input)
	if err != nil {
		log.Printf("FirehoseLogger: PutRecord error: %v", err)
	}
}