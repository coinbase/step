package bifrost

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/coinbase/step/aws/mocks"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func MockRelease() *Release {
	return &Release{
		AwsRegion:     to.Strp("region"),
		AwsAccountID:  to.Strp("account"),
		ReleaseID:     to.TimeUUID("release-"),
		CreatedAt:     to.Timep(time.Now()),
		ProjectName:   to.Strp("project"),
		ConfigName:    to.Strp("config"),
		Bucket:        to.Strp("bucket"),
		LockTableName: to.Strp("lock"),
	}
}

func MockAwsClients(r *Release) *mocks.MockClients {
	awsc := mocks.MockAwsClients()

	raw, _ := json.Marshal(r)

	awsc.S3.AddGetObject(fmt.Sprintf("%v/%v/%v/%v/release", *r.AwsAccountID, *r.ProjectName, *r.ConfigName, *r.ReleaseID), string(raw), nil)
	r.ReleaseSHA256 = to.SHA256Struct(&r)

	r.SetDefaults(r.AwsRegion, r.AwsAccountID, "")

	return awsc
}

func TestReleasePaths(t *testing.T) {
	release := MockRelease()
	release.ReleaseID = to.Strp("id")

	assert.Equal(t, "account/project", *release.ProjectDir())
	assert.Equal(t, "account/project/config", *release.RootDir())
	assert.Equal(t, "account/project/config/id", *release.ReleaseDir())
	assert.Equal(t, "account/project/config/id/release", *release.ReleasePath())
	assert.Equal(t, "account/project/config/id/log", *release.LogPath())
	assert.Equal(t, "account/project/config/lock", *release.RootLockPath())
	assert.Equal(t, "account/project/config/id/lock", *release.ReleaseLockPath())
	assert.Equal(t, "account/project/_shared", *release.SharedProjectDir())
}

func Test_Bifrost_Release_Is_Valid(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)

	assert.NoError(t, release.Validate(awsc.S3Client(nil, nil, nil), &Release{}))
}
