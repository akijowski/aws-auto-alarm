variable "target_sqs_arn" {
  description = "The ARN for the SQS target queue"
  type        = string
}

variable "rule_name" {
  description = "The EventBridge rule name"
  type        = string
}

variable "allowed_services" {
  description = "Allowed AWS services this rule will process"
  type        = set(string)
}

data "aws_caller_identity" "current" {}

resource "aws_cloudwatch_event_rule" "this" {
  name        = var.rule_name
  description = "Forward tag-change events to an SQS queue for processing"
  # event_bus_name leave blank for the default bus

  # https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-create-pattern-operators.html
  event_pattern = jsonencode({
    account     = [data.aws_caller_identity.current.account_id]
    source      = ["aws.tag"]
    detail-type = ["Tag Change on Resource"]
    detail = {
      service = var.allowed_services
      # AWS_AUTO_ALARM_ENABLED is present in either changed keys or existing tags
      "$or" = [
        {
          "changed-tag-keys" = ["AWS_AUTO_ALARM_ENABLED"]
          }, {
          "tags" = {
            "AWS_AUTO_ALARM_ENABLED" = [{ "exists" = true }]
          }
        }
      ]
    }
  })
}

resource "aws_cloudwatch_event_target" "sqs" {
  arn       = var.target_sqs_arn
  rule      = aws_cloudwatch_event_rule.this.name
  target_id = "TagsToSQS"
}

output "rule_arn" {
  value = aws_cloudwatch_event_rule.this.arn
}
