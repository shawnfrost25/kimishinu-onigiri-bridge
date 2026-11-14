resource "aws_iam_role" "backend_secret_fetch" {
  name = "kimishinu-${var.env}-backend-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sts:AssumeRole"
        ]
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}
