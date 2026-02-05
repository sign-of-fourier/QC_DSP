output "firehose_stream_name" {
  value = aws_kinesis_firehose_delivery_stream.google_bidreq.name
}

output "s3_bucket_name" {
  value = aws_s3_bucket.bidreq_logs.bucket
}

output "aws_region" {
  value = var.aws_region
}

output "bidder_port" {
  value = 8080
}

output "bidder_seat_id" {
  value = "test-seat"
}

output "bidder_default_campaign" {
  value = "default"
}