resource "aws_s3_bucket" "media" {
  bucket = var.media_s3
}


resource "aws_iam_policy" "s3_full_access" {
  name        = "${var.project_name}-s3-full-access"
  description = "Full access to the S3 bucket for ECS tasks"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "s3:ListBucket"
        ],
        Resource = [
          aws_s3_bucket.media.arn
        ]
      },
      {
        Effect = "Allow",
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject"
        ],
        Resource = [
          "${aws_s3_bucket.media.arn}/*"
        ]
      }
    ]
  })
}


resource "aws_iam_role_policy_attachment" "task_role_s3_policy_attachment" {
  role       = aws_iam_role.task_role.name
  policy_arn = aws_iam_policy.s3_full_access.arn
}

resource "aws_iam_role_policy_attachment" "task_execution_role_s3_policy_attachment" {
  role       = aws_iam_role.task_execution.name
  policy_arn = aws_iam_policy.s3_full_access.arn
}
