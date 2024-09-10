#!/bin/bash

set -e

echo "Updating api..."

# Updating web service
aws ecs update-service --cluster geo-tracker-be --service backend-api --force-new-deployment  --query "service.serviceName"  --output json
aws ecs update-service --cluster geo-tracker-be --service backend-worker --force-new-deployment  --query "service.serviceName"  --output json

echo "Done!"
