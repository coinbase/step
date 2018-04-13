package client

import (
	"testing"
	"time"

	"github.com/coinbase/step/aws/mocks"
	"github.com/coinbase/step/deployer"
	"github.com/coinbase/step/machine"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_Client_PrepareReleaseBundle(t *testing.T) {
	awsc := mocks.MockAwsClients()
	release := &deployer.Release{
		ReleaseId:        to.TimeUUID("release-"),
		CreatedAt:        to.Timep(time.Now()),
		ProjectName:      to.Strp("project"),
		ConfigName:       to.Strp("project"),
		LambdaName:       to.Strp("project"),
		StepFnName:       to.Strp("project"),
		Bucket:           to.Strp("project"),
		StateMachineJSON: to.Strp(machine.EmptyStateMachine),
		AwsRegion:        to.Strp("project"),
		AwsAccountID:     to.Strp("project"),
	}

	err := PrepareReleaseBundle(
		awsc,
		release,
		to.Strp("../../resources/empty_lambda.zip"), // Location to empty zip file
	)

	assert.NoError(t, err)
}
