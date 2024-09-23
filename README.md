# Geo Tracker Backend

![Build Status](https://github.com/Hodik/geo-tracker-be/actions/workflows/deploy.yml/badge.svg?branch=master)

## Overview

Geo Tracker Backend is a robust backend service for tracking geographical locations. It leverages AWS services for deployment and scaling, and uses Go for the backend logic. The project includes a comprehensive CI/CD pipeline for automated deployments.

## Table of Contents

- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Getting Started

### Prerequisites

- Go 1.22 or later
- Docker
- AWS CLI
- Terraform

### Installation

1. **Clone the repository**:

   ```sh
   git clone https://github.com/your-repo/geo-tracker-be.git
   cd geo-tracker-be
   ```

2. **Install Go dependencies**:

   ```sh
   go mod download
   ```

3. **Build the Go application**:

   ```sh
   go build -o main .
   ```

4. **Set up environment variables**:
   Create a `.env` file in the root directory and add the necessary environment variables.

### Running Locally

1. **Run the application**:

   ```sh
   ./main
   ```

2. **Run with Docker**:
   ```sh
   docker-compose up --build
   ```

## Configuration

### Go Configuration

The `Config` struct in `models/config.go` defines the configuration parameters for the application:

```go
type Config struct {
	PollInterval uint8  `gorm:"default:30" json:"poll_interval"`
	Dummy        string `gorm:"unique;default:'singleton'" json:"-"`
}
```

### Terraform Configuration

The Terraform scripts in the `terraform` directory manage the AWS infrastructure. Key files include:

- `02_network.tf`: Defines the VPC, subnets, and route tables.
- `03_load_balancer.tf`: Configures the application load balancer.
- `04_ecs.tf`: Sets up the ECS cluster, task definitions, and services.
- `05_rds.tf`: Manages the RDS instance and security groups.

### Environment Variables

The `variables.tf` file defines the necessary environment variables for Terraform:

```terraform
variable "region" {
  description = "The AWS region to create resources in."
  default     = "us-west-1"
}

variable "project_name" {
  description = "Project name to use in resource names"
  default     = "geo-tracker-be"
}
```

## Deployment

### CI/CD Pipeline

The project includes a GitHub Actions workflow for CI/CD, defined in `.github/workflows/deploy.yml`:

```yaml
name: Deploy CI/CD

on:
  push:
    branches:
      - main

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - name: Checkout
        uses: actions/checkout@v2

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Setup AWS CLI
        uses: aws-actions/configure-aws-credentials@v1
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ vars.AWS_REGION }}

      - name: Login to aws ecr
        run: aws ecr get-login-password | docker login --username AWS --password-stdin ${{ vars.ECR_IMAGE }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: ${{ vars.ECR_IMAGE }}

      - name: Deploy to ECS
        run: ./scripts/deploy.sh
```

### Manual Deployment

1. **Initialize Terraform**:

   ```sh
   terraform init
   ```

2. **Apply Terraform configuration**:

   ```sh
   terraform apply
   ```

3. **Run the deployment script**:
   ```sh
   ./scripts/deploy.sh
   ```

## Usage

### API Endpoints

The backend provides several API endpoints for interacting with the geographical tracking service. These endpoints are secured using JWT authentication.

### Middleware

- **DB Middleware**: Injects the database instance into the Gin context.
- **JWT Middleware**: Validates JWT tokens for secure endpoints.

This project was inspired by various open-source projects and aims to provide a comprehensive solution for geographical tracking.
