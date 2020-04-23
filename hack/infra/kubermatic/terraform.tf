terraform {
  required_version = ">= 0.11"

  backend "s3" {
    encrypt = true
    bucket = "tfstate-kubecarrier.io"
    region = "eu-central-1"
    key = "kubermatic"
  }
}

provider "aws" {
  region = "eu-central-1"
}

variable "AWS_ACCOUNT_ID" {
  type = string
}
