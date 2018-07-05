// s3 tools
package s3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/utils/to"
)

/////////
// ERRORS
/////////

type NotFoundError struct {
	bucket *string
	path   *string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Not Found %v %v", *e.bucket, *e.path)
}

func s3Error(bucket *string, path *string, err error) error {
	if err == nil {
		return nil
	}

	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case s3.ErrCodeNoSuchKey:
			return &NotFoundError{bucket, path}
		}
	}

	return err
}

/////////
// Wrappers
/////////

// Get downloads content from S3
func Get(s3c aws.S3API, bucket *string, path *string) (*[]byte, error) {
	return get(s3c, &s3.GetObjectInput{
		Bucket: bucket,
		Key:    path,
	})
}

func get(s3c aws.S3API, input *s3.GetObjectInput) (*[]byte, error) {
	results, err := s3c.GetObject(input)

	if err != nil {
		return nil, s3Error(input.Bucket, input.Key, err)
	}

	defer results.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, results.Body); err != nil {
		return nil, err
	}

	b := buf.Bytes()
	return &b, nil
}

// Put uploads content to s3
func Put(s3c aws.S3API, bucket *string, path *string, content *string) error {
	if content == nil {
		return fmt.Errorf("Put content is nil")
	}

	return put(s3c, &s3.PutObjectInput{
		Bucket: bucket,
		Key:    path,
		Body:   strings.NewReader(*content),
		ACL:    to.Strp("private"),
	})
}

func PutSecure(s3c aws.S3API, bucket *string, path *string, content *string, kmsKeyId *string) error {
	if content == nil {
		return fmt.Errorf("Put content is nil")
	}

	if kmsKeyId == nil {
		return fmt.Errorf("KMSKeyID content is nil")
	}

	return put(s3c, &s3.PutObjectInput{
		Bucket:               bucket,
		Key:                  path,
		Body:                 strings.NewReader(*content),
		ACL:                  to.Strp("private"),
		ServerSideEncryption: to.Strp("aws:kms"),
		SSEKMSKeyId:          kmsKeyId,
	})
}

func put(s3c aws.S3API, input *s3.PutObjectInput) error {
	_, err := s3c.PutObject(input)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes contents from S3
func Delete(s3c aws.S3API, bucket *string, path *string) error {
	_, err := s3c.DeleteObject(&s3.DeleteObjectInput{
		Bucket: bucket,
		Key:    path,
	})

	if err != nil {
		return err
	}

	return nil
}

/////////
// Struct Helpers
/////////

// GetStruct returns a Struct from S3
func GetStruct(s3c aws.S3API, bucket *string, path *string, str interface{}) error {
	raw, err := Get(s3c, bucket, path)
	if err != nil {
		return err
	}

	if raw == nil {
		return nil
	}

	err = json.Unmarshal(*raw, str)

	if err != nil {
		return err
	}

	return nil
}

// PutStruct Uploads a Struct to S3
func PutStruct(s3c aws.S3API, bucket *string, path *string, str interface{}) error {
	outputJSON, err := json.Marshal(str)

	if err != nil {
		return err
	}

	return Put(s3c, bucket, path, to.Strp(string(outputJSON)))
}

/////////
// Validation Helpers
/////////

// GetLastModified returns that last modified time of a file
func GetLastModified(s3c aws.S3API, bucket *string, path *string) (*time.Time, error) {
	results, err := s3c.GetObject(&s3.GetObjectInput{
		Bucket: bucket,
		Key:    path,
	})

	if err != nil {
		return nil, s3Error(bucket, path, err)
	}

	return results.LastModified, nil
}

/////////
// File Helpers
/////////

// PutFile uploads a file to S3
func PutFile(s3c aws.S3API, file_path *string, bucket *string, s3_file_path *string) error {
	return putFile(s3c, file_path, &s3.PutObjectInput{
		Bucket: bucket,
		Key:    s3_file_path,
		ACL:    to.Strp("private"),
	})
}

func PutSecureFile(s3c aws.S3API, file_path *string, bucket *string, s3_file_path *string, kmsKeyId *string) error {
	if kmsKeyId == nil {
		return fmt.Errorf("KMSKeyID content is nil")
	}

	return putFile(s3c, file_path, &s3.PutObjectInput{
		Bucket:               bucket,
		Key:                  s3_file_path,
		ACL:                  to.Strp("private"),
		ServerSideEncryption: to.Strp("aws:kms"),
		SSEKMSKeyId:          kmsKeyId,
	})
}

func putFile(s3c aws.S3API, file_path *string, input *s3.PutObjectInput) error {
	// Open the file
	bts, err := ioutil.ReadFile(*file_path)
	if err != nil {
		return err
	}

	input.SetBody(bytes.NewReader(bts))

	return put(s3c, input)
}

/////////
// SHA256 Helpers
/////////

// GetSHA256 returns a hex string of the SHA256 of the value of a key in S3
func GetSHA256(s3c aws.S3API, bucket *string, path *string) (string, error) {
	bytes, err := Get(s3c, bucket, path)
	if err != nil {
		return "", err
	}
	return to.SHA256AByte(bytes), nil
}
