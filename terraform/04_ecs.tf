# Production cluster
resource "aws_ecs_cluster" "cluster" {
  name = var.project_name
}

locals {
  container_vars = {
    region                = var.region
    image                 = aws_ecr_repository.ecr_repo.repository_url
    log_group             = aws_cloudwatch_log_group.backend.name
    rds_db_name           = var.rds_db_name
    rds_username          = var.rds_username
    rds_password          = var.rds_password
    rds_hostname          = aws_db_instance.db.address
    port                  = var.api_port
    AWS_ACCESS_KEY_ID     = var.aws_access_key_id
    AWS_SECRET_ACCESS_KEY = var.aws_secret_access_key
    AWS_REGION            = var.region
    auth0_audience        = var.auth0_audience
    auth0_domain          = var.auth0_domain
  }
}

# Backend web task definition and service
resource "aws_ecs_task_definition" "api_task" {
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 512
  memory                   = 1024
  family                   = "${var.project_name}-api"


  container_definitions = templatefile(
    "templates/backend-container.json.tpl",
    merge(
      local.container_vars,
      {
        name       = "${var.project_name}-api"
        command    = ["./main", "-mode=api"]
        log_stream = aws_cloudwatch_log_stream.backend_api.name
      }
    )
  )
  execution_role_arn = aws_iam_role.task_execution.arn
  task_role_arn      = aws_iam_role.task_role.arn
}


resource "aws_ecs_task_definition" "worker_task" {
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 512
  memory                   = 1024
  family                   = "${var.project_name}-worker"


  container_definitions = templatefile(
    "templates/backend-container.json.tpl",
    merge(
      local.container_vars,
      {
        name       = "${var.project_name}-worker"
        command    = ["./main", "-mode=worker"]
        log_stream = aws_cloudwatch_log_stream.backend_worker.name
      }
    )
  )
  execution_role_arn = aws_iam_role.task_execution.arn
  task_role_arn      = aws_iam_role.task_role.arn
}


resource "aws_ecs_task_definition" "migrator_task" {
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 512
  memory                   = 1024
  family                   = "${var.project_name}-migrator"


  container_definitions = templatefile(
    "templates/backend-container.json.tpl",
    merge(
      local.container_vars,
      {
        name       = "${var.project_name}-migrator"
        command    = ["./main", "-mode=migrator"]
        log_stream = aws_cloudwatch_log_stream.backend_migrator.name
      }
    )
  )
  execution_role_arn = aws_iam_role.task_execution.arn
  task_role_arn      = aws_iam_role.task_role.arn
}

resource "aws_ecs_service" "backend_api" {
  name                               = "backend-api"
  cluster                            = aws_ecs_cluster.cluster.id
  task_definition                    = aws_ecs_task_definition.api_task.arn
  desired_count                      = 1
  deployment_minimum_healthy_percent = 50
  deployment_maximum_percent         = 200
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  enable_execute_command             = true

  load_balancer {
    target_group_arn = aws_lb_target_group.target_group.arn
    container_name   = "${var.project_name}-api"
    container_port   = var.api_port
  }

  network_configuration {
    security_groups  = [aws_security_group.ecs_security_group.id]
    subnets          = [aws_subnet.public1.id, aws_subnet.public2.id]
    assign_public_ip = true
  }
}

resource "aws_ecs_service" "backend_worker" {
  name                               = "backend-worker"
  cluster                            = aws_ecs_cluster.cluster.id
  task_definition                    = aws_ecs_task_definition.worker_task.arn
  desired_count                      = 1
  deployment_minimum_healthy_percent = 50
  deployment_maximum_percent         = 200
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  enable_execute_command             = true

  network_configuration {
    security_groups  = [aws_security_group.ecs_security_group.id]
    subnets          = [aws_subnet.public1.id, aws_subnet.public2.id]
    assign_public_ip = true
  }
}

# Security Group
resource "aws_security_group" "ecs_security_group" {
  name   = "prod-ecs-backend"
  vpc_id = aws_vpc.vpc.id

  ingress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    security_groups = [aws_security_group.elb.id]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# IAM roles and policies
resource "aws_iam_role" "task_role" {
  name = "backend-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        },
        Effect = "Allow",
        Sid    = ""
      }
    ]
  })

  inline_policy {
    name = "backend-task-ssmmessages"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "ssmmessages:CreateControlChannel",
            "ssmmessages:CreateDataChannel",
            "ssmmessages:OpenControlChannel",
            "ssmmessages:OpenDataChannel",
          ]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }
}

resource "aws_iam_role" "task_execution" {
  name = "ecs-task-execution"

  assume_role_policy = jsonencode(
    {
      Version = "2012-10-17",
      Statement = [
        {
          Action = "sts:AssumeRole",
          Principal = {
            Service = "ecs-tasks.amazonaws.com"
          },
          Effect = "Allow",
          Sid    = ""
        }
      ]
    }
  )
}

resource "aws_iam_role_policy_attachment" "ecs-task-execution-role-policy-attachment" {
  role       = aws_iam_role.task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Cloudwatch Logs
resource "aws_cloudwatch_log_group" "backend" {
  name              = var.project_name
  retention_in_days = var.ecs_backend_retention_days
}

resource "aws_cloudwatch_log_stream" "backend_api" {
  name           = "backend-api"
  log_group_name = aws_cloudwatch_log_group.backend.name
}


resource "aws_cloudwatch_log_stream" "backend_worker" {
  name           = "backend-worker"
  log_group_name = aws_cloudwatch_log_group.backend.name
}


resource "aws_cloudwatch_log_stream" "backend_migrator" {
  name           = "backend-migrator"
  log_group_name = aws_cloudwatch_log_group.backend.name
}
