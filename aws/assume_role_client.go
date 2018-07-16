package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/coinbase/step/utils/to"
)

type AssumeRoleClient interface {
	GetSession() *session.Session
	SetSession(*session.Session)
	GetConfig(key string) *aws.Config
	SetConfig(key string, config *aws.Config)
}

func Session(arc AssumeRoleClient) *session.Session {
	if arc.GetSession() != nil {
		return arc.GetSession()
	}
	sess := session.Must(session.NewSession())
	arc.SetSession(sess)
	return sess
}

func Config(arc AssumeRoleClient, region *string, account_id *string, role *string) *aws.Config {
	if account_id == nil || region == nil || role == nil {
		return nil
	}

	key := fmt.Sprintf("%v--::--%v--::--%v", *region, *account_id, *role)

	if config := arc.GetConfig(key); config != nil {
		return config
	}

	arn := to.RoleArn(account_id, role)
	creds := stscreds.NewCredentials(Session(arc), *arn)
	config := aws.NewConfig().WithCredentials(creds).WithRegion(*region).WithMaxRetries(10)

	arc.SetConfig(key, config)
	return config
}
