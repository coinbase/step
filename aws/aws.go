package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
)

type S3API s3iface.S3API
type LambdaAPI lambdaiface.LambdaAPI
type SFNAPI sfniface.SFNAPI

type AwsClients interface {
	S3Client() S3API
	LambdaClient() LambdaAPI
	SFNClient() SFNAPI
}

type AwsClientsStr struct {
	s3Client     S3API
	lambdaClient LambdaAPI
	sfnClient    SFNAPI
}

func (awsc *AwsClientsStr) S3Client() S3API {
	return awsc.s3Client
}

func (awsc *AwsClientsStr) LambdaClient() LambdaAPI {
	return awsc.lambdaClient
}

func (awsc *AwsClientsStr) SFNClient() SFNAPI {
	return awsc.sfnClient
}

func CreateAwsClients() AwsClients {
	sess := session.Must(session.NewSession())

	return &AwsClientsStr{
		s3.New(sess),
		lambda.New(sess),
		sfn.New(sess),
	}
}
