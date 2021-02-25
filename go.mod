module gitlab.com/impervainc/tech-marketing/iac/lambda/go-vanity-server

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.2.0
	github.com/aws/aws-sdk-go-v2/config v1.1.1
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.0.2
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.1.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.1.1
	github.com/josh-hogle/logrus-cloudwatch-hook v0.1.1
	github.com/sirupsen/logrus v1.8.0
)
