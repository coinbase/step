package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
)

////////////
// Interfaces
////////////

type S3API s3iface.S3API
type LambdaAPI lambdaiface.LambdaAPI
type SFNAPI sfniface.SFNAPI

type AwsClients interface {
	S3Client(region *string, account_id *string, role *string) S3API
	LambdaClient(region *string, account_id *string, role *string) LambdaAPI
	SFNClient(region *string, account_id *string, role *string) SFNAPI
}

////////////
// AWS Clients
////////////

type AwsClientsStr struct {
	session *session.Session
	configs map[string]*aws.Config

	s3Client     S3API
	lambdaClient LambdaAPI
	sfnClient    SFNAPI
}

func (awsc *AwsClientsStr) GetSession() *session.Session {
	return awsc.session
}

func (awsc *AwsClientsStr) SetSession(sess *session.Session) {
	awsc.session = sess
}

func (awsc *AwsClientsStr) GetConfig(key string) *aws.Config {
	if awsc.configs == nil {
		return nil
	}

	config, ok := awsc.configs[key]
	if ok && config != nil {
		return config
	}

	return nil
}

func (awsc *AwsClientsStr) SetConfig(key string, config *aws.Config) {
	if awsc.configs == nil {
		awsc.configs = map[string]*aws.Config{}
	}
	awsc.configs[key] = config
}

func (awsc *AwsClientsStr) S3Client(region *string, account_id *string, role *string) S3API {
	return s3.New(Session(awsc), Config(awsc, region, account_id, role))
}

func (awsc *AwsClientsStr) LambdaClient(region *string, account_id *string, role *string) LambdaAPI {
	return lambda.New(Session(awsc), Config(awsc, region, account_id, role))
}

func (awsc *AwsClientsStr) SFNClient(region *string, account_id *string, role *string) SFNAPI {
	return sfn.New(Session(awsc), Config(awsc, region, account_id, role))
}
