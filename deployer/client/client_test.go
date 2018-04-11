package client

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/coinbase/step/aws/mocks"
	"github.com/coinbase/step/machine"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_Client_NewRelease(t *testing.T) {
	release := NewRelease(
		to.Strp(machine.EmptyStateMachine),
		to.Strp("lambda_name"),
		to.Strp("step"),
		to.Strp("bucket"),
		to.Strp("region"),
		to.Strp("account_id"),
		to.Strp("lambda_sha"),
	)

	release.ProjectName = to.Strp("project")
	release.ConfigName = to.Strp("config")

	assert.NoError(t, release.ValidateClientAttributes())
}

func Test_Client_PrepareRelease(t *testing.T) {
	awsc := mocks.MockAwsClients()
	awsc.Lambda.ListTagsResp = &lambda.ListTagsOutput{
		Tags: map[string]*string{"ProjectName": to.Strp("ProjectStep"), "ConfigName": to.Strp("Configy")},
	}

	release, err := PrepareRelease(
		awsc,
		to.Strp(machine.EmptyStateMachine),
		to.Strp("lambda_name"),
		to.Strp("step"),
		to.Strp("bucket"),
		to.Strp("../../resources/empty_lambda.zip"), // Location to empty zip file
	)

	assert.NoError(t, err)
	assert.Equal(t, "ProjectStep", *release.ProjectName)
	assert.Equal(t, "Configy", *release.ConfigName)
	assert.NoError(t, release.ValidateClientAttributes())
}

func Test_Client_PrepareReleaseBundle(t *testing.T) {
	awsc := mocks.MockAwsClients()
	awsc.Lambda.ListTagsResp = &lambda.ListTagsOutput{
		Tags: map[string]*string{"ProjectName": to.Strp("ProjectStep"), "ConfigName": to.Strp("Configy")},
	}

	release, err := PrepareReleaseBundle(
		awsc,
		to.Strp(machine.EmptyStateMachine),
		to.Strp("lambda_name"),
		to.Strp("step"),
		to.Strp("bucket"),
		to.Strp("../../resources/empty_lambda.zip"), // Location to empty zip file
	)

	assert.NoError(t, err)
	assert.Equal(t, "ProjectStep", *release.ProjectName)
	assert.Equal(t, "Configy", *release.ConfigName)
	assert.NoError(t, release.ValidateClientAttributes())
}
