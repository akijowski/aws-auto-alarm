# AWS Auto Alarm

The question is simple: is it possible to watch resource changes in AWS and create a standard set of alarms?

The idea comes from this blog article:
https://aws.amazon.com/blogs/mt/monitor-tag-changes-on-aws-resources-with-serverless-workflows-and-amazon-cloudwatch-events/

## Services

The following AWS services are supported:

- [ ] DynamoDB Table
- [ ] EventBridge Rule
- [x] SQS

## Upsert Alarms

During an upsert action, the code will generate alarms based on the provided data.

- The provided ARN is parsed and used to generate alarms based on the service.
- Additional configuration such as `alarmPrefix` and `overrides` will be processed as template data for the alarm.

The alarm data is then sent to Cloudwatch as an upsert operation `PutMetricAlarms`.

## Delete Alarms

The code will currently generate the current alarms based on the ARN and then try to delete them based on the generated names.

**TODO: This is not yet implemented.**

During the delete action, the code will find and delete all alarms that have the following tags:

- `AWS_AUTO_ALARM_MANAGED=true`
- `AWS_AUTO_ALARM_SOURCE_ARN=<provided arn>`

## Step Functions

TODO: This approach has been put on hold in favor of using a big ol' Lambda function to do the message processing via SQS.

## Lambda Function

A Lambda function is used to consume EventBridge events off of an SQS queue.
Each queue message will correspond to a tag resource event, which will trigger the creation or deletion of a CloudWatch Alarm.

The Lambda will process messages from SQS.
Each SQS message will have a body that contains the tag change event details.

```json
{
  "version": "0",
  "id": "bddcf1d6-0251-35a1-aab0-adc1fb47c11c",
  "detail-type": "Tag Change on Resource",
  "source": "aws.tag",
  "account": "123456789012",
  "time": "2018-09-18T20:41:38Z",
  "region": "us-east-1",
  "resources": [
    "arn:aws:ec2:us-east-1:123456789012:instance/i-0000000aaaaaaaaaa"
  ],
  "detail": {
    "changed-tag-keys": [
      "a-new-key",
      "an-updated-key",
      "a-deleted-key"
    ],
    "service": "ec2",
    "resource-type": "instance",
    "version": 3,
    "tags": {
      "a-new-key": "tag-value-on-new-key-just-added",
      "an-updated-key": "tag-value-was-just-changed",
      "an-unchanged-key": "tag-value-still-the-same"
    }
  }
}
```

The Lambda function will upsert or delete CloudWatch alarms based on the tag change event.

An upsert action is determined based on the following:

- `detail.tags` contains a key `AWS_AUTO_ALARM_ENABLED` and the value is `true`.

Tags prefixed with `AWS_AUTO_ALARM_` will be passed to the alarm upsert process as configuration.

A delete action is determined based on the following:

- `detail.changed-tag-keys` contains `AWS_AUTO_ALARM_ENABLED` and `detail.tags` does not contain a key `AWS_AUTO_ALARM_ENABLED`.

### Configure sample input

I don't have anything too fancy right now.
Here is how you can modify the sample lambda input:

1. Modify the `sample_cloudwatch_event.json` file to include the desired input.
2. Run the following command to "stringify" the json using `jq`:

```bash
jq -cM < ./samples/lambda/sample_cloudwatch_event.json | pbcopy
```

3. Paste the output into the `./samples/lambda/sample_lambda_input.json` file. Use the "body" field.

## CLI

The CLI is used to parse a config file and upsert or delete alarms.
Additionally, the CLI can be used to output a sample of data as a "dry-run".

Example config:

```json
{
    "quiet": false,
    "dryRun": true,
    "prettyPrint": true,
    "delete": false,
    "ARN": "arn:aws:sqs:us-east-1:0123456789012:queue/test-queue",
    "alarmActions": [
        "arn:aws:sns:us-east-1:0123456789012:topic/Foo"
    ],
    "okActions": [
        "arn:aws:sns:us-east-1:0123456789012:topic/Bar"
    ],
    "alarmPrefix": "hello"
}
```

## TODO

Outstanding tasks:

- [ ] Replace viper config with https://github.com/knadh/koanf
  - [ ] Remove use in lambda config
  - [ ] Remove use in cli config
- [ ] Better CLI command parsing
- [ ] Better naming of config fields
