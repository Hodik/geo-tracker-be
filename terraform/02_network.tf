# Production VPC
resource "aws_vpc" "vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

# Public subnets
resource "aws_subnet" "public1" {
  cidr_block        = "10.0.1.0/24"
  vpc_id            = aws_vpc.vpc.id
  availability_zone = var.availability_zones[0]
  tags = {
    Name = "geo-tracker-be"
  }
}

resource "aws_subnet" "public2" {
  cidr_block        = "10.0.2.0/24"
  vpc_id            = aws_vpc.vpc.id
  availability_zone = var.availability_zones[1]
  tags = {
    Name = "geo-tracker-be"
  }
}

# Route tables and association with the subnets
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.vpc.id
}

resource "aws_route_table_association" "public1" {
  route_table_id = aws_route_table.public.id
  subnet_id      = aws_subnet.public1.id
}

resource "aws_route_table_association" "public2" {
  route_table_id = aws_route_table.public.id
  subnet_id      = aws_subnet.public2.id
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id
}

resource "aws_route" "igw" {
  route_table_id         = aws_route_table.public.id
  gateway_id             = aws_internet_gateway.igw.id
  destination_cidr_block = "0.0.0.0/0"
}
