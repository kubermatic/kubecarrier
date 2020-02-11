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
        "s3:GetLifecycleConfiguration",
        "s3:GetBucketTagging",
        "s3:GetInventoryConfiguration",
        "s3:GetObjectVersionTagging",
        "s3:ListBucketVersions",
        "s3:GetBucketLogging",
        "s3:ListBucket",
        "s3:GetAccelerateConfiguration",
        "s3:GetBucketPolicy",
        "s3:GetObjectVersionTorrent",
        "s3:GetObjectAcl",
        "s3:GetEncryptionConfiguration",
        "s3:GetBucketObjectLockConfiguration",
        "s3:GetBucketRequestPayment",
        "s3:GetAccessPointPolicyStatus",
        "s3:GetObjectVersionAcl",
        "s3:GetObjectTagging",
        "s3:GetMetricsConfiguration",
        "s3:GetBucketPublicAccessBlock",
        "s3:GetBucketPolicyStatus",
        "s3:ListBucketMultipartUploads",
        "s3:GetObjectRetention",
        "s3:GetBucketWebsite",
        "s3:GetBucketVersioning",
        "s3:GetBucketAcl",
        "s3:GetObjectLegalHold",
        "s3:GetBucketNotification",
        "s3:GetReplicationConfiguration",
        "s3:ListMultipartUploadParts",
        "s3:PutObject",
        "s3:GetObject",
        "s3:GetObjectTorrent",
        "s3:DescribeJob",
        "s3:GetBucketCORS",
        "s3:GetAnalyticsConfiguration",
        "s3:GetObjectVersionForReplication",
        "s3:GetBucketLocation",
        "s3:GetAccessPointPolicy",
        "s3:GetObjectVersion"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:s3:::${var.bucket_name}"
    }
  ]
}
EOF
}
