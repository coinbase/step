package dynamodb

import (
	"fmt"
	"time"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"

	stepaws "github.com/coinbase/step/aws"
)

var (
	columnKey  = "key"
	columnId   = "id"
	columnTime = "time"
)

func GrabLock(dc stepaws.DynamoDBAPI, tableName string, lockPath string, uuid string) (bool, error) {
	// Construct a conditional expression such that we only allow a new lock
	// to be created if there is not already one for the same key.
	condExp := expression.Name(columnKey).AttributeNotExists()
	condExp = condExp.Or(expression.Name(columnId).Equal(expression.Value(uuid)))

	expr, err := expression.NewBuilder().WithCondition(condExp).Build()
	if err != nil {
		return false, err
	}

	// Attempt to create a lock
	_, err = dc.PutItem(&dynamodb.PutItemInput{
		TableName:                 awssdk.String(tableName),
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Item: map[string]*dynamodb.AttributeValue{
			columnKey: {
				S: awssdk.String(lockPath),
			},
			columnId: {
				S: awssdk.String(uuid),
			},
			columnTime: {
				S: awssdk.String(time.Now().Format(time.RFC3339)),
			},
		},
	})

	if err != nil {
		awsErr, ok := err.(awserr.Error)
		// A lock already exists for the same key.
		if ok && awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func ReleaseLock(dc stepaws.DynamoDBAPI, tableName string, lockPath string, uuid string) error {
	// Construct a condition expression such that we only allow a lock
	// to be deleted if the key, and the UUID aligns.
	condExp := expression.Name(columnId).Equal(expression.Value(uuid))
	expr, err := expression.NewBuilder().WithCondition(condExp).Build()
	if err != nil {
		return err
	}

	// Attempt to delete lock
	_, err = dc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName:                 awssdk.String(tableName),
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key: map[string]*dynamodb.AttributeValue{
			columnKey: {
				S: awssdk.String(lockPath),
			},
		},
	})

	if err != nil {
		awsErr, ok := err.(awserr.Error)
		// A lock already exists, but with a different UUID.
		if ok && awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
			return fmt.Errorf("Lock was stolen for release with UUID(%v)", uuid)
		}

		return err
	}

	return nil
}
