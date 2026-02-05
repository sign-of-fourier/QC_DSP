#!/bin/bash
#chmod +x source_env.sh
INFRA_DIR="./infra"

export FIREHOSE_STREAM_NAME=$(terraform -chdir=$INFRA_DIR output -raw firehose_stream_name)
export S3_BUCKET_NAME=$(terraform -chdir=$INFRA_DIR output -raw s3_bucket_name)
export AWS_REGION=$(terraform -chdir=$INFRA_DIR output -raw aws_region)
export BIDDER_PORT=$(terraform -chdir=$INFRA_DIR output -raw bidder_port)
export BIDDER_SEAT_ID=$(terraform -chdir=$INFRA_DIR output -raw bidder_seat_id)

echo "Loaded env from Terraform:"
echo "FIREHOSE_STREAM_NAME=$FIREHOSE_STREAM_NAME"
echo "S3_BUCKET_NAME=$S3_BUCKET_NAME"
echo "AWS_REGION=$AWS_REGION"
echo "BIDDER_PORT=$BIDDER_PORT"
echo "BIDDER_SEAT_ID=$BIDDER_SEAT_ID"