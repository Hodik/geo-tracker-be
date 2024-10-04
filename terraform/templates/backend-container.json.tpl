[
  {
    "name": "${name}",
    "image": "${image}",
    "essential": true,
    "links": [],
    "portMappings": [
      {
        "containerPort": 8000,
        "hostPort": 8000,
        "protocol": "tcp"
      }
    ],
    "command": ${jsonencode(command)},
    "environment": [
      {
        "name": "DB_STRING",
        "value": "postgresql://${rds_username}:${rds_password}@${rds_hostname}:5432/${rds_db_name}"
      },
      {"name": "PORT", "value": "${port}"},
      {"name": "AUTH0_DOMAIN", "value": "${auth0_domain}"},
      {"name": "AUTH0_AUDIENCE", "value": "${auth0_audience}"},
      {"name": "AWS_ACCESS_KEY_ID", "value": "${AWS_ACCESS_KEY_ID}"},
      {"name": "AWS_SECRET_ACCESS_KEY", "value": "${AWS_SECRET_ACCESS_KEY}"},
      {"name": "AWS_REGION", "value": "${AWS_REGION}"}
    ],
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-group": "${log_group}",
        "awslogs-region": "${region}",
        "awslogs-stream-prefix": "${log_stream}"
      }
    }
  }
]