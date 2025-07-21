#!/bin/bash

# Configuration
PROJECT_ID="your-gcp-project-id"  # Replace with your actual project ID
SERVICE_NAME="cpimp-scanner"
REGION="us-central1"
IMAGE_NAME="gcr.io/${PROJECT_ID}/${SERVICE_NAME}"

echo "🔨 Building Docker image..."
docker build -t ${IMAGE_NAME} .

echo "📤 Pushing image to Google Container Registry..."
docker push ${IMAGE_NAME}

echo "🚀 Deploying to Cloud Run..."
gcloud run deploy ${SERVICE_NAME} \
  --image ${IMAGE_NAME} \
  --platform managed \
  --region ${REGION} \
  --allow-unauthenticated \
  --memory 2Gi \
  --cpu 2 \
  --timeout 3600 \
  --concurrency 1 \
  --max-instances 1 \
  --set-env-vars="SCAN_MODE=production" \
  --project ${PROJECT_ID}

echo "✅ Deployment complete!"
echo "📊 View logs with: gcloud logs tail --follow --service=${SERVICE_NAME}"
echo "🔍 Monitor at: https://console.cloud.google.com/run/detail/${REGION}/${SERVICE_NAME}/logs?project=${PROJECT_ID}" 