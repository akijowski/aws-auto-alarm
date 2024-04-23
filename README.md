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
