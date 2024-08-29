variable "lambda_name" {
  description = "Name of the lambda function"
  type        = string
}

variable "abs_path_to_archive_file" {
  description = "Absolute path to the lambda zip archive"
  type        = string
  validation {
    condition     = fileexists(var.abs_path_to_archive_file)
    error_message = "The file at the specified path does not exist. Did you package the Lambda?"
  }
}

variable "lambda_role_arn" {
  description = "ARN of the role to be attached to the lambda function"
  type        = string
}

variable "sqs_queue_arn" {
  description = "ARN of the SQS queue to be attached to the lambda function"
  type        = string
}

resource "aws_lambda_function" "this" {
  function_name = var.lambda_name
  description   = "Tweek Week 2024 project"
  role          = var.lambda_role_arn

  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]

  timeout = 30
  publish = true

  filename         = var.abs_path_to_archive_file
  source_code_hash = filebase64sha256(var.abs_path_to_archive_file)

  environment {
    variables = {
      "AWS_AUTO_ALARM_LOG_LEVEL" = "info"
    }
  }
}

resource "aws_lambda_alias" "this" {
  function_name    = aws_lambda_function.this.arn
  function_version = aws_lambda_function.this.version
  name             = "live"
}

resource "aws_lambda_event_source_mapping" "sqs" {
  function_name    = aws_lambda_alias.this.arn
  event_source_arn = var.sqs_queue_arn
  enabled          = true
}

output "arn" {
  value = aws_lambda_function.this.qualified_arn
}
