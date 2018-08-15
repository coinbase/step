package bifrost

import (
	"testing"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_Lock_GrabRelease(t *testing.T) {
	r := MockRelease()

	r2 := MockRelease()
	r2.UUID = to.Strp("NOTUUID")

	awsc := MockAwsClients(r)
	s3c := awsc.S3Client(nil, nil, nil)

	assert.NoError(t, r.GrabLock(s3c))
	assert.NoError(t, r.GrabLock(s3c))
	assert.Error(t, r2.GrabLock(s3c))

	assert.NoError(t, r.ReleaseLock(s3c))
	assert.NoError(t, r.ReleaseLock(s3c))
	assert.NoError(t, r2.GrabLock(s3c))
}
