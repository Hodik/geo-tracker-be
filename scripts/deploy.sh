#!/bin/bash

set -e


# Collect ECS_GROUP_ID and PRIVATE_SUBNET_ID for running migrations
echo "Collecting data..."
ECS_GROUP_ID=$(aws ec2 describe-security-groups --filters Name=group-name,Values=prod-ecs-backend --query "SecurityGroups[*][GroupId]" --output text)
PRIVATE_SUBNET_ID=$(aws ec2 describe-subnets  --filters "Name=tag:Name,Values=geo-tracker-be-1" --query "Subnets[*][SubnetId]"  --output text)

echo "Running migration task..."
# Construct NETWORK_CONFIGURATON to run migtaion task 
NETWORK_CONFIGURATON="{\"awsvpcConfiguration\": {\"subnets\": [\"${PRIVATE_SUBNET_ID}\"], \"securityGroups\": [\"${ECS_GROUP_ID}\"],\"assignPublicIp\": \"ENABLED\"}}"
# Start migration task

echo ${NETWORK_CONFIGURATON}
MIGRATION_TASK_ARN=$(aws ecs run-task --cluster geo-tracker-be --task-definition geo-tracker-be-migrator --count 1 --launch-type FARGATE --network-configuration "${NETWORK_CONFIGURATON}" --query 'tasks[*][taskArn]' --output text)
echo "Task ${MIGRATION_TASK_ARN} running..."
# Wait migration task to complete
aws ecs wait tasks-stopped --cluster geo-tracker-be --tasks "${MIGRATION_TASK_ARN}"


echo "Updating api..."

# Updating web service
aws ecs update-service --cluster geo-tracker-be --service backend-api --force-new-deployment  --query "service.serviceName"  --output json
aws ecs update-service --cluster geo-tracker-be --service backend-worker --force-new-deployment  --query "service.serviceName"  --output json

echo "Done!"
