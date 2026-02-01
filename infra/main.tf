locals {
  firehose_stream_name = "${var.project_name}-google-bidreq-firehose"
  s3_bucket_name       = "${var.project_name}-bidreq-logs"
}

# S3 bucket to store logs
resource "aws_s3_bucket" "bidreq_logs" {
  bucket = local.s3_bucket_name
}

# New-style lifecycle configuration
resource "aws_s3_bucket_lifecycle_configuration" "bidreq_logs" {
  bucket = aws_s3_bucket.bidreq_logs.id

  rule {
    id     = "expire-old-logs"
    status = "Enabled"

    expiration {
      days = 30
    }

    filter {
      prefix = "bid_requests/"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "bidreq_logs" {
  bucket                  = aws_s3_bucket.bidreq_logs.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# IAM role that Firehose will assume
resource "aws_iam_role" "firehose_role" {
  name = "${var.project_name}-firehose-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "firehose.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

# IAM policy for Firehose to write to S3 and log to CloudWatch
resource "aws_iam_role_policy" "firehose_policy" {
  name = "${var.project_name}-firehose-policy"
  role = aws_iam_role.firehose_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowS3"
        Effect = "Allow"
        Action = [
          "s3:AbortMultipartUpload",
          "s3:GetBucketLocation",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
          "s3:PutObject"
        ]
        Resource = [
          aws_s3_bucket.bidreq_logs.arn,
          "${aws_s3_bucket.bidreq_logs.arn}/*"
        ]
      },
      {
        Sid    = "AllowLogs"
        Effect = "Allow"
        Action = [
          "logs:PutLogEvents"
        ]
        Resource = "*"
      }
    ]
  })
}

# Firehose delivery stream (direct-put)
resource "aws_kinesis_firehose_delivery_stream" "google_bidreq" {
  name        = local.firehose_stream_name
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose_role.arn
    bucket_arn = aws_s3_bucket.bidreq_logs.arn

    # Main successful delivery prefix
    prefix = "bid_requests/dt=!{timestamp:yyyy-MM-dd}/hr=!{timestamp:HH}/"

    # REQUIRED when prefix has expressions
    error_output_prefix = "bid_requests_error/!{firehose:error-output-type}/dt=!{timestamp:yyyy-MM-dd}/"

    buffering_size     = 5   # MiB
    buffering_interval = 300 # seconds (5 minutes)
    compression_format = "GZIP"

    cloudwatch_logging_options {
      enabled         = true
      log_group_name  = "/aws/kinesisfirehose/${local.firehose_stream_name}"
      log_stream_name = "S3Delivery"
    }
  }
}