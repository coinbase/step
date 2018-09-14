package s3

import (
	"testing"

	"github.com/coinbase/step/aws/mocks"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_Get_Success(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	_, err := Get(s3c, to.Strp("bucket"), to.Strp("/path"))
	assert.Error(t, err)
	assert.IsType(t, &NotFoundError{}, err)

	s3c.AddGetObject("/path", "asd", nil)
	out, err := Get(s3c, to.Strp("bucket"), to.Strp("/path"))
	assert.NoError(t, err)
	assert.Equal(t, "asd", string(*out))
}

func Test_Put_Success(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	key := to.Strp("/path")
	err := PutStr(s3c, bucket, key, to.Strp("asdji"))
	assert.NoError(t, err)

	out, err := Get(s3c, bucket, key)
	assert.NoError(t, err)
	assert.Equal(t, "asdji", string(*out))
}

func Test_Delete_Success(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	key := to.Strp("/path")
	err := PutStr(s3c, bucket, key, to.Strp("asdji"))
	assert.NoError(t, err)

	out, err := Get(s3c, bucket, key)
	assert.NoError(t, err)
	assert.Equal(t, "asdji", string(*out))

	err = Delete(s3c, bucket, key)
	assert.NoError(t, err)

	_, err = Get(s3c, bucket, key)
	assert.Error(t, err)
	assert.IsType(t, &NotFoundError{}, err)
}

func Test_GetStruct_Success(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	s3c.AddGetObject("/path", `{"name": "asd"}`, nil)
	str := struct {
		Name string
	}{}

	err := GetStruct(s3c, to.Strp("bucket"), to.Strp("/path"), &str)
	assert.NoError(t, err)
	assert.Equal(t, "asd", str.Name)
}

func Test_PutStruct_Success(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	key := to.Strp("/path")
	err := PutStruct(s3c, bucket, key, struct {
		Name string
	}{"asd"})
	assert.NoError(t, err)

	str := struct {
		Name string
	}{}

	err = GetStruct(s3c, to.Strp("bucket"), to.Strp("/path"), &str)
	assert.NoError(t, err)
	assert.Equal(t, "asd", str.Name)
}
