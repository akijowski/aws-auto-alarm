<!-- BEGIN_TF_DOCS -->
## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_event_rule.this](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_rule) | resource |
| [aws_cloudwatch_event_target.sqs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_target) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_allowed_services"></a> [allowed\_services](#input\_allowed\_services) | Allowed AWS services this rule will process | `set(string)` | n/a | yes |
| <a name="input_rule_name"></a> [rule\_name](#input\_rule\_name) | The EventBridge rule name | `string` | n/a | yes |
| <a name="input_target_sqs_arn"></a> [target\_sqs\_arn](#input\_target\_sqs\_arn) | The ARN for the SQS target queue | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_rule_arn"></a> [rule\_arn](#output\_rule\_arn) | n/a |
<!-- END_TF_DOCS -->