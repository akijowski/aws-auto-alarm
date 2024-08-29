terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}

variable "sqs_queue_arn" {
  description = "ARN of the SQS the lambda will read from"
  type        = string
}

variable "role_name" {
  description = "Name of the role"
  type        = string
}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "lambda_access" {
  statement {
    effect = "Allow"

    principals {
      identifiers = ["lambda.amazonaws.com"]
      type = "Service"
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "lambda" {
  statement {
    sid = "SQS"

    effect = "Allow"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:Get*",
      "sqs:List*"
    ]
    resources = [var.sqs_queue_arn]
  }

  statement {
    sid = "ManageAlarms"

    effect = "Allow"
    actions = [
      "cloudwatch:ListTagsForResource",
      "cloudwatch:DescribeAlarms",
      "cloudwatch:DeleteAlarms",
      "cloudwatch:DisableAlarmActions",
      "cloudwatch:EnableAlarmActions",
      "cloudwatch:PutCompositeAlarm",
      "cloudwatch:PutMetricAlarm",
      #       "cloudwatch:SetAlarmState",
      "cloudwatch:TagResource",
      "cloudwatch:UntagResource"
    ]

    resources = ["arn:aws:cloudwatch:*:${data.aws_caller_identity.current.account_id}:alarm:*"]
  }
}

resource "aws_iam_role" "lambda" {
  name               = var.role_name
  description        = "Tweek Week 2024 project"
  assume_role_policy = data.aws_iam_policy_document.lambda_access.json
}

resource "aws_iam_role_policy" "lambda" {
  name = "permissions"
  policy = data.aws_iam_policy_document.lambda.json
  role   = aws_iam_role.lambda.name
}

resource "aws_iam_role_policy_attachment" "basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.lambda.name
}

output "lambda_role_arn" {
  value = aws_iam_role.lambda.arn
}
