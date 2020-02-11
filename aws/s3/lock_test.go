package s3

import (
	"fmt"
	"testing"

	"github.com/coinbase/step/aws/mocks"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_GrabLock_Success(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	grabbed, err := GrabLock(s3c, bucket, path, "UUID")

	assert.NoError(t, err)
	assert.True(t, grabbed)
}

func Test_GrabUserLock_Success(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	grabbed, err := GrabUserLock(s3c, bucket, path)

	assert.NoError(t, err)
	assert.True(t, grabbed)
}

func Test_GrabLock_Success_Already_Has_Lock(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	s3c.AddGetObject(*path, `{"uuid": "UUID"}`, nil)
	grabbed, err := GrabLock(s3c, bucket, path, "UUID")

	assert.NoError(t, err)
	assert.True(t, grabbed)
}

func Test_GrabLock_Failure_Already_Locked(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	s3c.AddGetObject(*path, `{"uuid": "NOT_UUID"}`, nil)
	grabbed, err := GrabLock(s3c, bucket, path, "UUID")

	assert.NoError(t, err)
	assert.False(t, grabbed)
}
func Test_GrabUserLock_Failure_Already_Locked(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	s3c.AddGetObject(*path, `{"user": "test", "lock_reason": "testing"}`, nil)
	grabbed, err := GrabUserLock(s3c, bucket, path)

	assert.NoError(t, err)
	assert.False(t, grabbed)
}

func Test_GrabLock_Failure_S3_Get_Error(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	s3c.AddGetObject(*path, `{"uuid": "NOT_UUID"}`, fmt.Errorf("ERRRR"))
	grabbed, err := GrabLock(s3c, bucket, path, "UUID")

	assert.Error(t, err)
	assert.False(t, grabbed)
}

func Test_GrabUserLock_Failure_S3_Get_Error(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	s3c.AddGetObject(*path, `{"user": "test", "lock_reason": "hello"}`, fmt.Errorf("ERRRR"))
	grabbed, err := GrabUserLock(s3c, bucket, path)

	assert.Error(t, err)
	assert.False(t, grabbed)
}

func Test_GrabLock_Failure_S3_Upload_Error(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	s3c.AddPutObject(*path, fmt.Errorf("ERRRR"))
	grabbed, err := GrabLock(s3c, bucket, path, "UUID")

	assert.Error(t, err)
	assert.True(t, grabbed)
}

func Test_GrabUserLock_Failure_S3_Upload_Error(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	s3c.AddPutObject(*path, fmt.Errorf("ERRRR"))
	grabbed, err := GrabUserLock(s3c, bucket, path)

	assert.Error(t, err)
	assert.True(t, grabbed)
}

func Test_ReleaseLock_Success_No_Object(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	err := ReleaseLock(s3c, bucket, path, "UUID")

	assert.NoError(t, err)
}
func Test_ReleaseUserLock_Success_No_Object(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	err := ReleaseUserLock(s3c, bucket, path)

	assert.NoError(t, err)
}

func Test_ReleaseLock_Success_Correct_Lock(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	s3c.AddGetObject(*path, `{"uuid": "UUID"}`, nil)
	err := ReleaseLock(s3c, bucket, path, "UUID")

	assert.NoError(t, err)
}

func Test_ReleaseUserLock_Success_Correct_Lock(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	s3c.AddGetObject(*path, `{"user": "test", "lock_reason": "hello"}`, nil)
	err := ReleaseUserLock(s3c, bucket, path)

	assert.NoError(t, err)
}

func Test_ReleaseLock_Failure_AnotherReleasesLock(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	path := to.Strp("path")

	s3c.AddGetObject(*path, `{"uuid": "NOT_UUID"}`, nil)
	err := ReleaseLock(s3c, bucket, path, "UUID")

	assert.Error(t, err)
}
