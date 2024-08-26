resource "aws_db_subnet_group" "db_subnet_group" {
  name       = "rds-subnet-group"
  subnet_ids = [aws_subnet.public1.id, aws_subnet.public2.id]
}

resource "aws_db_instance" "db" {
  identifier              = var.project_name
  db_name                 = var.rds_db_name
  username                = var.rds_username
  password                = var.rds_password
  port                    = "5432"
  engine                  = "postgres"
  engine_version          = "12.17"
  instance_class          = var.rds_instance_class
  allocated_storage       = "20"
  storage_encrypted       = false
  vpc_security_group_ids  = [aws_security_group.rds_prod.id]
  db_subnet_group_name    = aws_db_subnet_group.db_subnet_group.name
  multi_az                = false
  storage_type            = "gp2"
  publicly_accessible     = true
  backup_retention_period = 5
  skip_final_snapshot     = true
}

resource "aws_security_group" "rds_prod" {
  name   = "rds-security-group"
  vpc_id = aws_vpc.vpc.id

  ingress {
    protocol        = "tcp"
    from_port       = "5432"
    to_port         = "5432"
    security_groups = [aws_security_group.ecs_security_group.id]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}
