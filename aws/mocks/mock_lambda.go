package mocks

import (
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
)

type MockLambdaClient struct {
	lambdaiface.LambdaAPI
	UpdateFunctionCodeResp  *lambda.FunctionConfiguration
	UpdateFunctionCodeError error
	ListTagsResp            *lambda.ListTagsOutput
}

func (m *MockLambdaClient) init() {
	if m.UpdateFunctionCodeResp == nil {
		m.UpdateFunctionCodeResp = &lambda.FunctionConfiguration{}
	}
}

func (m *MockLambdaClient) UpdateFunctionCode(in *lambda.UpdateFunctionCodeInput) (*lambda.FunctionConfiguration, error) {
	m.init()
	return m.UpdateFunctionCodeResp, m.UpdateFunctionCodeError
}

func (m *MockLambdaClient) ListTags(in *lambda.ListTagsInput) (*lambda.ListTagsOutput, error) {
	m.init()
	return m.ListTagsResp, nil
}
