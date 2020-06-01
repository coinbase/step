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
	locker := NewInMemoryLocker()

	t.Run("root lock acquired", func(t *testing.T) {
		assert.NoError(t, r.GrabRootLock(s3c, locker, "lambdaname"))
	})

	t.Run("same root lock acquired", func(t *testing.T) {
		assert.NoError(t, r.GrabRootLock(s3c, locker, "lambdaname"))

		locks := locker.GetLockByNamespace("lambdaname")
		// We are re-using an existing lock
		assert.Equal(t, len(locks), 1)
		assert.Equal(t, locks[0].lockPath, "account/project/config/lock")
	})

	t.Run("conflict when acquiring root lock with different uuid", func(t *testing.T) {
		assert.Error(t, r2.GrabRootLock(s3c, locker, "lambdaname"))
		// There should be no changes in the existing locks
		assert.Equal(t, len(locker.GetLockByNamespace("lambdaname")), 1)
	})

	t.Run("root lock released", func(t *testing.T) {
		assert.NoError(t, r.UnlockRoot(s3c, locker, "lambdaname"))
		assert.Equal(t, len(locker.GetLockByNamespace("lambdaname")), 0)
	})

	t.Run("same root lock released", func(t *testing.T) {
		assert.NoError(t, r.UnlockRoot(s3c, locker, "lambdaname"))
		assert.Equal(t, len(locker.GetLockByNamespace("lambdaname")), 0)
	})

	t.Run("root lock with same uuid acquired", func(t *testing.T) {
		assert.NoError(t, r2.GrabRootLock(s3c, locker, "lambdaname"))
		locks := locker.GetLockByNamespace("lambdaname")
		assert.Equal(t, len(locks), 1)
		assert.Equal(t, locks[0].lockPath, "account/project/config/lock")
		assert.Equal(t, locks[0].uuid, "NOTUUID")
	})
}
