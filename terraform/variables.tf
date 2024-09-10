variable "region" {
  description = "The AWS region to create resources in."
  default     = "us-west-1"
}

variable "project_name" {
  description = "Project name to use in resource names"
  default     = "geo-tracker-be"
}

variable "availability_zones" {
  description = "Availability zones"
  default     = ["us-west-1a", "us-west-1b"]
}

variable "ecs_backend_retention_days" {
  description = "Retention period for backend logs"
  default     = 30
}

variable "api_port" {
  description = "Port for the API service"
  default     = 8000
}

# rds

variable "rds_db_name" {
  description = "RDS database name"
  default     = "geo_tracker"
}

variable "rds_username" {
  description = "RDS database username"
  default     = "geo_tracker"
}

variable "rds_password" {
  description = "postgres password for DB"
}

variable "rds_instance_class" {
  description = "RDS instance type"
  default     = "db.t3.micro"
}

variable "aws_access_key_id" {
  description = "AWS key"
}

variable "aws_secret_access_key" {
  description = "AWS key secret"
}


variable "auth0_domain" {
  description = "Auth0 Domain for auth"
  default     = "dev-q0x1ep2b4jnbc1u8.us.auth0.com"
}


variable "auth0_audience" {
  description = "Auth0 client id for auth"
  default     = "mNOR3o8CwsHc6WZiP6mZGQdUDiNshVXb"
}
