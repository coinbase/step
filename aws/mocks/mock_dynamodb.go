package mocks

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type MockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI

	PutItemInputs    []*dynamodb.PutItemInput
	DeleteItemInputs []*dynamodb.DeleteItemInput

	PutItemError    error
	DeleteItemError error
}

func (m *MockDynamoDBClient) init() {
	if m.PutItemInputs == nil {
		m.PutItemInputs = []*dynamodb.PutItemInput{}
	}

	if m.DeleteItemInputs == nil {
		m.DeleteItemInputs = []*dynamodb.DeleteItemInput{}
	}
}

func (m *MockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	if m.PutItemError != nil {
		return nil, m.PutItemError
	}

	m.PutItemInputs = append(m.PutItemInputs, input)
	return &dynamodb.PutItemOutput{}, nil
}

func (m *MockDynamoDBClient) DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	if m.DeleteItemError != nil {
		return nil, m.DeleteItemError
	}

	m.DeleteItemInputs = append(m.DeleteItemInputs, input)
	return &dynamodb.DeleteItemOutput{}, nil
}
