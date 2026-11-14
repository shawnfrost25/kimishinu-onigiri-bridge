resource "aws_iam_role" "backend_secret_fetch" {
  name = "test_policy"
  assume_role_policy = jsondecode({
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
