package bifrost

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Lock_GrabRelease(t *testing.T) {
	r := MockRelease()
	awsc := MockAwsClients(r)
	s3c := awsc.S3Client(nil, nil, nil)
	grabbed, err := r.GrabLock(s3c)
	assert.NoError(t, err)
	assert.Equal(t, grabbed, true)

	grabbed, err = r.GrabLock(s3c)
	assert.NoError(t, err)
	assert.Equal(t, grabbed, true)

	assert.NoError(t, r.ReleaseLock(s3c))
	assert.NoError(t, r.ReleaseLock(s3c))
}
