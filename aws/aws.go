package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
	"github.com/coinbase/step/utils/to"
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

func (awsc *AwsClientsStr) getSession() *session.Session {
	if awsc.session != nil {
		return awsc.session
	}
	awsc.session = session.Must(session.NewSession())
	return awsc.session
}

func (awsc *AwsClientsStr) getConfig(region *string, account_id *string, role *string) *aws.Config {
	if account_id == nil || region == nil || role == nil {
		return nil
	}

	if awsc.configs == nil {
		awsc.configs = map[string]*aws.Config{}
	}

	key := fmt.Sprintf("%v--::--%v--::--%v", *region, *account_id, *role)
	config, ok := awsc.configs[key]
	if ok && config != nil {
		return config
	}

	arn := to.RoleArn(account_id, role)
	creds := stscreds.NewCredentials(awsc.session, *arn)
	config = aws.NewConfig().WithCredentials(creds).WithRegion(*region)

	awsc.configs[key] = config
	return config
}

func (awsc *AwsClientsStr) S3Client(region *string, account_id *string, role *string) S3API {
	return s3.New(awsc.getSession(), awsc.getConfig(region, account_id, role))
}

func (awsc *AwsClientsStr) LambdaClient(region *string, account_id *string, role *string) LambdaAPI {
	return lambda.New(awsc.getSession(), awsc.getConfig(region, account_id, role))
}

func (awsc *AwsClientsStr) SFNClient(region *string, account_id *string, role *string) SFNAPI {
	return sfn.New(awsc.getSession(), awsc.getConfig(region, account_id, role))
}
