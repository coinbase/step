package bifrost

import (
	"fmt"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Lock_GrabRootLock(t *testing.T) {
	r := MockRelease()

	r2 := MockRelease()
	r2.UUID = to.Strp("NOTUUID")

	awsc := MockAwsClients(r)
	dc := awsc.DynamoDBClient(nil, nil, nil)

	assert.NoError(t, r.GrabRootLock(dc))
	assert.NoError(t, r.GrabRootLock(dc))
	fmt.Println(r2.GrabRootLock(dc))
	assert.Error(t, r2.GrabRootLock(dc))

	assert.NoError(t, r.UnlockRoot(dc))
	assert.NoError(t, r.UnlockRoot(dc))
	assert.NoError(t, r2.GrabRootLock(dc))
}
