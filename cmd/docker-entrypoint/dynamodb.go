package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sirupsen/logrus"
)

// AccountEntry represents an account on which to run the Lambda function.
type AccountEntry struct {
	EntryID            string   `json:"EntryID"`
	AccountID          string   `json:"AccountID"`
	AllowPorts         []int32  `json:"AllowPorts"`
	Regions            []string `json:"Regions"`
	RoleName           string   `json:"RoleName"`
	SessionName        string   `json:"SessionName"`
	DurationSeconds    int32    `json:"DurationSeconds"`
	ExternalID         string   `json:"ExternalID"`
	TagName            string   `json:"TagName"`
	ExclusiveTagValues []string `json:"ExclusiveTagValues"`
	ManagedTagValues   []string `json:"ManagedTagValues"`
	Description        string   `json:"Description"`
}

// getAccounts returns the accounts and corresponding configurations for running the Lambda function.
func getAccounts(client *dynamodb.Client, tableName string) ([]AccountEntry, error) {
	logger := log.WithFields(logrus.Fields{
		"dynamoTable": tableName,
	})
	var accounts []AccountEntry

	// get the list of accounts and configurations
	params := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	result, err := client.Scan(context.TODO(), params)
	if err != nil {
		logger.Errorf("Failed to get accounts from DynamoDB: %s", err)
		return accounts, err
	}
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &accounts); err != nil {
		logger.Errorf("Failed to parse list of accounts returned from DynamoDB: %s", err)
		return accounts, err
	}
	logger.WithFields(logrus.Fields{
		"accounts": accounts,
	}).Debug("Account configurations retrieved from DynamoDB")
	return accounts, nil
}

// newDynamoClient creates a new DynamoDB client in the given region.
func newDynamoClient(region string) (*dynamodb.Client, error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.WithFields(logrus.Fields{
			"region": region,
		}).Errorf("Failed to load DynamoDB client config: %s", err)
		return nil, err
	}
	return dynamodb.NewFromConfig(awsConfig), nil
}

// validateAccount checks the account to make sure values are valid and sets defaults where necessary.
func validateAccount(account AccountEntry) (AccountEntry, error) {
	acct := account
	if acct.AccountID == "" {
		msg := "Account ID cannot be empty"
		log.Error(msg)
		return acct, fmt.Errorf(msg)
	}

	if len(acct.AllowPorts) == 0 {
		acct.AllowPorts = []int32{80, 443}
	}

	if acct.Description == "" {
		acct.Description = "Managed by UpdateSonarCloudSecurityGroups function (DO NOT MODIFY)"
	}

	if acct.DurationSeconds == 0 {
		acct.DurationSeconds = 1800 // 30 minutes
	}

	if len(acct.Regions) == 0 {
		msg := "You must specify at least 1 region for an account"
		log.Error(msg)
		return acct, fmt.Errorf(msg)
	}

	if acct.RoleName == "" {
		acct.RoleName = "STS-UpdateSonarCloudSecurityGroupsRole"
	}

	if acct.SessionName == "" {
		acct.SessionName = "UpdateSonarCloudSecurityGroupsFunction"
	}

	if acct.TagName == "" {
		acct.TagName = "fn.imperva.com/UpdateSonarCloudSecurityGroups/state"
	}

	if len(acct.ManagedTagValues) == 0 {
		acct.ManagedTagValues = []string{"managed"}
	}

	if len(acct.ExclusiveTagValues) == 0 {
		acct.ExclusiveTagValues = []string{"exclusive"}
	}
	return acct, nil
}
