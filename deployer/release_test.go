package deployer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coinbase/step/aws/mocks"
	"github.com/coinbase/step/utils/to"
)

func Test_Release_DeployStepFunction(t *testing.T) {
	sfnClient := &mocks.MockSFNClient{}
	r := MockRelease()

	err := r.DeployStepFunction(sfnClient)
	assert.NoError(t, err)
}

func Test_Release_DeployLambda(t *testing.T) {
	lambdaClient := &mocks.MockLambdaClient{}
	s3c := &mocks.MockS3Client{}

	r := MockRelease()
	r.Bucket = to.Strp("bucket")
	s3c.AddGetObject(*r.LambdaZipPath(), "", nil)

	err := r.DeployLambda(lambdaClient, s3c)
	assert.NoError(t, err)

}
