# Application Load Balancer for production
resource "aws_lb" "elb" {
  name               = "prod"
  load_balancer_type = "application"
  internal           = false
  security_groups    = [aws_security_group.elb.id]
  subnets            = [aws_subnet.public1.id, aws_subnet.public2.id]
}

# Target group for backend web application
resource "aws_lb_target_group" "target_group" {
  name        = var.project_name
  port        = 80
  protocol    = "HTTP"
  vpc_id      = aws_vpc.vpc.id
  target_type = "ip"

  health_check {
    path                = "/ping/"
    port                = "traffic-port"
    healthy_threshold   = 5
    unhealthy_threshold = 2
    timeout             = 29
    interval            = 30
    matcher             = "200"
  }
}

# Target listener for http:80
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.elb.id
  port              = "80"
  protocol          = "HTTP"
  depends_on        = [aws_lb_target_group.target_group]

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.target_group.arn
  }
}

# Target listener for https:443
#resource "aws_alb_listener" "prod_https" {
#  load_balancer_arn = aws_lb.elb.id
#  port              = "443"
#  protocol          = "HTTPS"
#  ssl_policy        = "ELBSecurityPolicy-2016-08"
#  depends_on        = [aws_lb_target_group.target_group]

# default_action {
#    type             = "forward"
#    target_group_arn = aws_lb_target_group.target_group.arn
#  }
#}

# Allow traffic from 80 and 443 ports only
resource "aws_security_group" "elb" {
  name        = "prod-lb"
  description = "Controls access to the ALB"
  vpc_id      = aws_vpc.vpc.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
