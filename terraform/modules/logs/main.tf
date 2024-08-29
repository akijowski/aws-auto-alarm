variable "lambda_name" {
  description = "The name of the lambda function"
  type        = string
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.lambda_name}"
  retention_in_days = 14
}
