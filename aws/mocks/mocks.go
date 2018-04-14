package mocks

import "github.com/coinbase/step/aws"

type MockAwsClientsStr struct {
	S3     *MockS3Client
	Lambda *MockLambdaClient
	SFN    *MockSFNClient
}

func (awsc *MockAwsClientsStr) S3Client(*string, *string, *string) aws.S3API {
	return awsc.S3
}

func (awsc *MockAwsClientsStr) LambdaClient(*string, *string, *string) aws.LambdaAPI {
	return awsc.Lambda
}

func (awsc *MockAwsClientsStr) SFNClient(*string, *string, *string) aws.SFNAPI {
	return awsc.SFN
}

func MockAwsClients() *MockAwsClientsStr {
	return &MockAwsClientsStr{
		&MockS3Client{},
		&MockLambdaClient{},
		&MockSFNClient{},
	}
}
