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
		AwsRegion:    to.Strp("region"),
		AwsAccountID: to.Strp("account"),
		ReleaseID:    to.TimeUUID("release-"),
		CreatedAt:    to.Timep(time.Now()),
		ProjectName:  to.Strp("project"),
		ConfigName:   to.Strp("config"),
		Bucket:       to.Strp("bucket"),
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

func Test_Bifrost_Release_Is_Valid(t *testing.T) {
	release := MockRelease()
	awsc := MockAwsClients(release)

	assert.NoError(t, release.Validate(awsc.S3Client(nil, nil, nil), &Release{}))
}
