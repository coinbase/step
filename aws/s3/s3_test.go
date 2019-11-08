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

func Test_Put_With_Type_Success(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	key := to.Strp("/path")
	content := []byte("<html></html>")
	contentType := to.Strp("text/html")
	err := PutWithType(s3c, bucket, key, &content, contentType)
	assert.NoError(t, err)

	object, out, err := GetObject(s3c, bucket, key)
	assert.NoError(t, err)
	assert.Equal(t, "<html></html>", string(*out))
	assert.Equal(t, "text/html", string(*object.ContentType))
}

func Test_Put_With_Cache_Control_Success(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	key := to.Strp("/path")
	content := []byte("asdji")
	cacheControl := to.Strp("public, max-age=31556926")
	err := PutWithCacheControl(s3c, bucket, key, &content, cacheControl)
	assert.NoError(t, err)

	object, out, err := GetObject(s3c, bucket, key)
	assert.NoError(t, err)
	assert.Equal(t, "asdji", string(*out))
	assert.Equal(t, "public, max-age=31556926", string(*object.CacheControl))
}

func Test_Put_With_Type_And_Cache_Control_Success(t *testing.T) {
	s3c := &mocks.MockS3Client{}
	bucket := to.Strp("bucket")
	key := to.Strp("/path")
	content := []byte("<html></html>")
	contentType := to.Strp("text/html")
	cacheControl := to.Strp("public, max-age=31556926")
	err := PutWithTypeAndCacheControl(s3c, bucket, key, &content, contentType, cacheControl)
	assert.NoError(t, err)

	object, out, err := GetObject(s3c, bucket, key)
	assert.NoError(t, err)
	assert.Equal(t, "<html></html>", string(*out))
	assert.Equal(t, "text/html", string(*object.ContentType))
	assert.Equal(t, "public, max-age=31556926", string(*object.CacheControl))
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
