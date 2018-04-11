package deployer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coinbase/step/aws/mocks"
	"github.com/coinbase/step/utils/to"
)

func Test_Release_Is_Valid(t *testing.T) {
	r := MockRelease()
	r.SetDefaults(to.Strp("asd"), to.Strp("asd"))
	assert.NoError(t, r.ValidateAttributes())
}

func Test_Release_DeployStepFunction(t *testing.T) {
	sfnClient := &mocks.MockSFNClient{}
	r := MockRelease()

	err := r.DeployStepFunction(sfnClient)
	assert.NoError(t, err)
}

func Test_Release_DeployLambda(t *testing.T) {
	lambdaClient := &mocks.MockLambdaClient{}
	r := MockRelease()

	err := r.DeployLambda(lambdaClient)
	assert.NoError(t, err)

}
