package bifrost

import (
	"testing"
	"time"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_IsHalt_ReleaseTimeout(t *testing.T) {
	r := MockRelease()
	awsc := MockAwsClients(r)

	assert.NoError(t, r.IsHalt(awsc.S3))
	r.Timeout = to.Intp(0)
	assert.Error(t, r.IsHalt(awsc.S3))

	// 10 second halt
	r.StartedAt = to.Timep(time.Now().Add(-1 * (9 * time.Second)))
	r.Timeout = to.Intp(10)
	assert.NoError(t, r.IsHalt(awsc.S3))
	r.StartedAt = to.Timep(time.Now().Add(-1 * (11 * time.Second)))
	assert.Error(t, r.IsHalt(awsc.S3))
}

func Test_IsHalt_HaltKey(t *testing.T) {
	r := MockRelease()
	awsc := MockAwsClients(r)

	assert.NoError(t, r.IsHalt(awsc.S3))
	assert.NoError(t, r.Halt(awsc.S3, to.Strp("error")))
	assert.Error(t, r.IsHalt(awsc.S3))

	// If the Halt key is older than 5 mins ignore it
	awsc.S3.GetObjectResp[*r.HaltPath()].Resp.LastModified = to.Timep(time.Now().Add(-1 * (10 * time.Minute)))
	assert.NoError(t, r.IsHalt(awsc.S3))
}
