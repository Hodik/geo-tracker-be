resource "aws_ecr_repository" "ecr_repo" {
  name                 = "${var.project_name}-backend"
  image_tag_mutability = "MUTABLE"
}
