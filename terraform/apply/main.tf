variable "project_name" {
  description = "The name of the project"
  type        = string
}

variable "aws_region" {
  description = "The AWS region"
  type        = string
  validation {
    condition     = contains(["us-east-1", "us-east-2", "us-west-2"], var.aws_region)
    error_message = "The variable aws_region must be a valid region: one of us-east-1, us-east-2, or us-west-2."
  }
}

module "sqs" {
  source = "../modules/sqs"

  queue_name      = var.project_name
  event_rule_name = "${var.project_name}-sqs"
}

module "iam_role" {
  source = "../modules/iam_role"

  role_name     = var.project_name
  sqs_queue_arn = module.sqs.arn

  providers = {
    aws = aws.global
  }
}

module "lambda" {
  source = "../modules/lambda"

  lambda_name              = var.project_name
  abs_path_to_archive_file = abspath("${path.module}/../../out/bootstrap.zip")
  lambda_role_arn          = module.iam_role.lambda_role_arn
  sqs_queue_arn            = module.sqs.arn

}

module "logging" {
  source = "../modules/logs"

  lambda_name = var.project_name
}

module "eventbridge" {
  source = "../modules/eventbridge"

  rule_name        = "${var.project_name}-sqs"
  target_sqs_arn   = module.sqs.arn
  allowed_services = toset(["sqs"])
}

output "sqs_queue_arn" {
  value = module.sqs.arn
}

output "iam_role_arn" {
  value = module.iam_role.lambda_role_arn
}

output "lambda_arn" {
  value = module.lambda.arn
}
