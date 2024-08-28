# AWS Auto Alarm

The question is simple: is it possible to watch resource changes in AWS and create a standard set of alarms?

## Services

The following AWS services are supported:

- [ ] DynamoDB Table
- [ ] EventBridge Rule
- [ ] SQS

## Step Functions

TODO: This approach has been put on hold in favor of using a big ol' Lambda function to do the message processing via SQS.

## Lambda Function

A Lambda function is used to consume EventBridge events off of an SQS queue.
Each queue message will correspond to a tag resource event, which will trigger the creation or deletion of a CloudWatch Alarm.

### Configure sample input

I don't have anything too fancy right now.
Here is how you can modify the sample lambda input:

1. Modify the `sample_cloudwatch_event.json` file to include the desired input.
2. Run the following command to "stringify" the json using `jq`:

```bash
jq -cM < ./samples/lambda/sample_cloudwatch_event.json | pbcopy
```

3. Paste the output into the `./samples/lambda/sample_lambda_input.json` file. Use the "body" field.
