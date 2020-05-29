package dynamodb

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/stretchr/testify/assert"
)

type MockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	putItemCallback    func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	deleteItemCallback func(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
}

func (c *MockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return c.putItemCallback(input)
}

func (c *MockDynamoDBClient) DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return c.deleteItemCallback(input)
}

func TestLock(t *testing.T) {
	t.Run("lock failure", func(t *testing.T) {
		client := &MockDynamoDBClient{}

		client.putItemCallback = func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
			return nil, awserr.New(dynamodb.ErrCodeConditionalCheckFailedException, "The conditional request failed.", errors.New("fake error"))
		}

		grabbed, err := GrabLock(client, "tableName", "lockPath", "uuid")
		assert.NoError(t, err)
		assert.False(t, grabbed)
	})

	t.Run("lock acquired successfully", func(t *testing.T) {
		client := &MockDynamoDBClient{}

		client.putItemCallback = func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
			assert.Equal(t, "tableName", *input.TableName)
			assert.Equal(t, "lockPath", *input.Item[columnKey].S)
			assert.Equal(t, "uuid", *input.Item[columnId].S)
			assert.Equal(t, "(attribute_not_exists (#0)) OR (#1 = :0)", *input.ConditionExpression)

			assert.Equal(t, "key", *input.ExpressionAttributeNames["#0"])
			assert.Equal(t, "id", *input.ExpressionAttributeNames["#1"])
			assert.Equal(t, "uuid", *input.ExpressionAttributeValues[":0"].S)

			return &dynamodb.PutItemOutput{}, nil
		}

		grabbed, err := GrabLock(client, "tableName", "lockPath", "uuid")
		assert.NoError(t, err)
		assert.True(t, grabbed)
	})
}

func TestUnlock(t *testing.T) {
	t.Run("unlock failure", func(t *testing.T) {
		client := &MockDynamoDBClient{}
		client.deleteItemCallback = func(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
			return nil, awserr.New(dynamodb.ErrCodeConditionalCheckFailedException, "The conditional request failed.", errors.New("fake error"))
		}

		err := ReleaseLock(client, "tableName", "lockPath", "uuid")
		assert.Error(t, err)
	})

	t.Run("unlock released", func(t *testing.T) {
		client := &MockDynamoDBClient{}

		client.deleteItemCallback = func(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
			assert.Equal(t, "tableName", *input.TableName)

			assert.Equal(t, "id", *input.ExpressionAttributeNames["#0"])
			assert.Equal(t, "uuid", *input.ExpressionAttributeValues[":0"].S)
			assert.Equal(t, "lockPath", *input.Key[columnKey].S)

			return &dynamodb.DeleteItemOutput{}, nil
		}

		err := ReleaseLock(client, "tableName", "lockPath", "uuid")
		assert.NoError(t, err)
	})
}
