package bifrost

import (
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Lock_GrabRootLock(t *testing.T) {
	r := MockRelease()

	r2 := MockRelease()
	r2.UUID = to.Strp("NOTUUID")

	awsc := MockAwsClients(r)
	s3c := awsc.S3Client(nil, nil, nil)
	dc := awsc.DynamoDBClient(nil, nil, nil)

	assert.NoError(t, r.GrabRootLock(s3c, dc))
	assert.NoError(t, r.GrabRootLock(s3c, dc))
	assert.Error(t, r2.GrabRootLock(s3c, dc))

	assert.NoError(t, r.UnlockRoot(s3c, dc))
	assert.NoError(t, r.UnlockRoot(s3c, dc))
	assert.NoError(t, r2.GrabRootLock(s3c, dc))
}
