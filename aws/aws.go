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

type Clients struct {
	session *session.Session
	configs map[string]*aws.Config
}

func (c Clients) Session() *session.Session {
	if c.session != nil {
		return c.session
	}
	// new session
	sess := session.Must(session.NewSession())
	c.session = sess
	return sess
}

func (c Clients) Config(
	region *string,
	account_id *string,
	role *string) *aws.Config {

	config := aws.NewConfig().WithMaxRetries(10)

	if region != nil {
		config = config.WithRegion(*region)
	}

	// return no config for nil inputs
	if account_id == nil || role == nil {
		return config
	}

	// Assume a role
	arn := fmt.Sprintf(
		"arn:aws:iam::%v:role/%v",
		*account_id,
		*role,
	)

	// include region in cache key otherwise concurrency errors
	key := fmt.Sprintf("%v::%v", *region, arn)

	// check for cached config
	if c.configs != nil && c.configs[key] != nil {
		return c.configs[key]
	}

	// new creds
	creds := stscreds.NewCredentials(c.Session(), arn)

	// new config
	config = config.WithCredentials(creds)

	if c.configs == nil {
		c.configs = map[string]*aws.Config{}
	}

	c.configs[key] = config
	return config
}

func (c *Clients) S3Client(
	region *string,
	account_id *string,
	role *string) S3API {
	return s3.New(c.Session(), c.Config(region, account_id, role))
}

func (c *Clients) LambdaClient(region *string, account_id *string, role *string) LambdaAPI {
	return lambda.New(c.Session(), c.Config(region, account_id, role))
}

func (c *Clients) SFNClient(region *string, account_id *string, role *string) SFNAPI {
	return sfn.New(c.Session(), c.Config(region, account_id, role))
}
