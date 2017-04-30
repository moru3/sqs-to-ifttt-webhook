sqs-to-ifttt-webhook

# Function
Regularly acquire messages notified to SQS and periodically post data to IFTTT's Webhook URL.

- https://ifttt.com/maker_webhooks

# setting.yml
Place setting.yml in the same hierarchy.

```yml
awsRegion: [AWS Region]
queueURL: [SQS endpoint URL]
iftttURL: [IFTTT's webhook URL]
interval: [loop interval (seconds)]
```
