variable "bucket_name" {
  description = "bucket_name"
  type = string
}

resource "aws_s3_bucket" "logs" {
  bucket = var.bucket_name
  acl    = "private"

  lifecycle_rule {
    id      = "log"
    enabled = true
    expiration {
      days = 30
    }
  }
}

resource "aws_iam_user" "writer" {
  name = "${var.bucket_name}_writer"
}

resource "aws_iam_user_policy" "policy" {
  name = "${var.bucket_name}.fill"
  user = aws_iam_user.writer.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:PutObject"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:s3:::${var.bucket_name}/*"
    }
  ]
}
EOF
}
