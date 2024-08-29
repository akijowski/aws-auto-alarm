variable "queue_name" {
  description = "Name of the SQS queue"
  type        = string
}

variable "event_rule_name" {
  description = "Name of the CloudWatch Event Rule that will deliver messages"
  type        = string
}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "this" {
  statement {
    sid = "AllowEventsSendMessages"

    effect = "Allow"
    actions = [
      "sqs:SendMessage"
    ]
    principals {
      identifiers = ["events.amazonaws.com"]
      type        = "Service"
    }
    resources = [
      aws_sqs_queue.this.arn
    ]
    condition {
      test     = "ArnEquals"
      values   = ["arn:aws:events:*:${data.aws_caller_identity.current.account_id}:rule/${var.event_rule_name}"]
      variable = "aws:SourceArn"
    }
  }
}

resource "aws_sqs_queue" "this" {
  name = var.queue_name
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 3
  })
}

resource "aws_sqs_queue_policy" "this" {
  policy    = data.aws_iam_policy_document.this.json
  queue_url = aws_sqs_queue.this.url
}

resource "aws_sqs_queue" "dlq" {
  name = "${var.queue_name}-dlq"
}

resource "aws_sqs_queue_redrive_allow_policy" "dlq" {
  queue_url = aws_sqs_queue.dlq.id

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.this.arn]
  })
}

output "url" {
  value = aws_sqs_queue.this.url
}

output "arn" {
  value = aws_sqs_queue.this.arn
}
