#!/bin/sh
set -e

# MinIO server details (replace with your actual endpoint and credentials if different)
MINIO_ENDPOINT=${MINIO_ENDPOINT:-"http://minio:9000"}
MINIO_ACCESS_KEY=${MINIO_ROOT_USER:-"minioadmin"}
MINIO_SECRET_KEY=${MINIO_ROOT_PASSWORD:-"minioadmin"}
MC_ALIAS="local"

# Wait for MinIO server to be ready
echo "Waiting for MinIO server at $MINIO_ENDPOINT..."
until mc alias set $MC_ALIAS "$MINIO_ENDPOINT" "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY"; do
    echo "MinIO server not ready yet, retrying in 5 seconds..."
    sleep 5
done
echo "MinIO server is ready."

# Buckets to create
BUCKETS="media icons avatars"

# Policy template (allows public GetObject)
POLICY_TEMPLATE='{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {"AWS": ["*"]},
            "Action": ["s3:GetBucketLocation"],
            "Resource": ["arn:aws:s3:::%s"]
        },
        {
            "Effect": "Allow",
            "Principal": {"AWS": ["*"]},
            "Action": ["s3:GetObject"],
            "Resource": ["arn:aws:s3:::%s/*"]
        }
    ]
}'

# Create buckets and apply policies
for BUCKET in $BUCKETS; do
    BUCKET_PATH="$MC_ALIAS/$BUCKET"
    echo "Checking bucket: $BUCKET_PATH"

    if ! mc ls "$BUCKET_PATH" > /dev/null 2>&1; then
        echo "Creating bucket: $BUCKET_PATH"
        mc mb "$BUCKET_PATH"
    else
        echo "Bucket $BUCKET_PATH already exists."
    fi

    echo "Applying policy to bucket: $BUCKET_PATH"
    POLICY_JSON=$(printf "$POLICY_TEMPLATE" "$BUCKET" "$BUCKET")
    echo "$POLICY_JSON" | mc policy set-json - "$BUCKET_PATH"
    echo "Policy applied to $BUCKET_PATH."

    # Configure webhook notification for this bucket
    WEBHOOK_TARGET_URL="http://api:3000/api/v1/webhook/storage/events" # Corrected URL
    WEBHOOK_QUEUE_ARN="arn:minio:sqs::${BUCKET}:webhook" # Use bucket name in ARN for uniqueness

    echo "Configuring webhook for bucket: $BUCKET_PATH to target $WEBHOOK_TARGET_URL"

    # Add event notification configuration
    # Send events for object creation (PUT, POST, COPY, CompleteMultipartUpload) and deletion
    # NOTE: Current API handler only processes 'media' bucket events. Sending for others might be noisy.
    # Removed suffix filter to trigger for all file types
    mc event add "$BUCKET_PATH" "$WEBHOOK_QUEUE_ARN" --event "s3:ObjectCreated:*,s3:ObjectRemoved:*" --webhook "$WEBHOOK_TARGET_URL" || echo "WARN: Failed to add webhook event listener for $BUCKET_PATH. Continuing..."

    # Add a check to see if the configuration was successful (optional but good practice)
    echo "Verifying webhook configuration for $BUCKET_PATH..."
    mc event list "$BUCKET_PATH" "$WEBHOOK_QUEUE_ARN"

done

echo "MinIO initialization complete."

# Keep the container running if needed for debugging or further steps
# If this script runs as a one-off job, this line might not be necessary.
# tail -f /dev/null 