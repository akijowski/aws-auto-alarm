terraform {
  required_version = "~> 1.8"

  backend "s3" {}

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.64"
    }
  }
}

locals {
  project_tags = {
    "Tweek2024" = "true"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = local.project_tags
  }
}

provider "aws" {
  region = "us-east-1"
  alias  = "global"

  default_tags {
    tags = local.project_tags
  }
}
