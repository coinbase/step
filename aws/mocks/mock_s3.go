package mocks

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/coinbase/step/utils/to"
)

// S3Client
type GetObjectResponse struct {
	Resp  *s3.GetObjectOutput
	Body  string
	Error error
}

type PutObjectResponse struct {
	Resp  *s3.PutObjectOutput
	Error error
}

type DeleteObjectResponse struct {
	Resp  *s3.DeleteObjectOutput
	Error error
}

type GetBucketTaggingResponse struct {
	Resp  *s3.GetBucketTaggingOutput
	Error error
}

type MockS3Client struct {
	s3iface.S3API

	GetObjectResp map[string]*GetObjectResponse

	PutObjectResp map[string]*PutObjectResponse

	DeleteObjectResp map[string]*DeleteObjectResponse

	GetBucketTaggingResp map[string]*GetBucketTaggingResponse
}

func (m *MockS3Client) init() {
	if m.GetObjectResp == nil {
		m.GetObjectResp = map[string]*GetObjectResponse{}
	}

	if m.PutObjectResp == nil {
		m.PutObjectResp = map[string]*PutObjectResponse{}
	}

	if m.DeleteObjectResp == nil {
		m.DeleteObjectResp = map[string]*DeleteObjectResponse{}
	}

	if m.GetBucketTaggingResp == nil {
		m.GetBucketTaggingResp = map[string]*GetBucketTaggingResponse{}
	}
}

func MakeS3Body(ret string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(ret))
}

func makeS3Resp(ret string, contentType *string) *s3.GetObjectOutput {
	return &s3.GetObjectOutput{
		Body:         MakeS3Body(ret),
		ContentType:  contentType,
		LastModified: to.Timep(time.Now()),
	}
}

func AWSS3NotFoundError() error {
	return awserr.New(s3.ErrCodeNoSuchKey, "not found", nil)
}

func (m *MockS3Client) addGetObjectWithContentType(key string, body string, contentType *string, err error) {
	m.init()
	m.GetObjectResp[key] = &GetObjectResponse{
		Resp:  makeS3Resp(body, contentType),
		Body:  body,
		Error: err,
	}
}

func (m *MockS3Client) AddGetObject(key string, body string, err error) {
	m.addGetObjectWithContentType(key, body, nil, err)
}

func (m *MockS3Client) AddPutObject(key string, err error) {
	m.init()
	m.PutObjectResp[key] = &PutObjectResponse{
		Resp:  &s3.PutObjectOutput{},
		Error: err,
	}
}

func (m *MockS3Client) SetBucketTags(bucket string, tags map[string]string, err error) {
	m.init()
	tagSet := []*s3.Tag{}

	for tk, tv := range tags {
		tagSet = append(tagSet, &s3.Tag{Key: to.Strp(tk), Value: to.Strp(tv)})
	}

	m.GetBucketTaggingResp[bucket] = &GetBucketTaggingResponse{
		Resp: &s3.GetBucketTaggingOutput{
			TagSet: tagSet,
		},
		Error: err,
	}
}

func (m *MockS3Client) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	m.init()
	resp := m.GetObjectResp[*in.Key]

	if resp == nil {
		return nil, AWSS3NotFoundError()
	}

	resp.Resp.Body = MakeS3Body(resp.Body)
	return resp.Resp, resp.Error
}

func (m *MockS3Client) ListObjects(in *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
	return nil, nil
}

func (m *MockS3Client) PutObject(in *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	m.init()

	resp := m.PutObjectResp[*in.Key]
	// Simulates adding the object
	buf := new(bytes.Buffer)
	buf.ReadFrom(in.Body)
	m.addGetObjectWithContentType(*in.Key, buf.String(), in.ContentType, nil)

	if resp == nil {
		return &s3.PutObjectOutput{}, nil
	}
	return resp.Resp, resp.Error
}

func (m *MockS3Client) GetBucketTagging(in *s3.GetBucketTaggingInput) (*s3.GetBucketTaggingOutput, error) {
	m.init()
	resp := m.GetBucketTaggingResp[*in.Bucket]
	if resp == nil {
		return nil, fmt.Errorf("Unkown Bucket, should mock the tags")
	}
	return resp.Resp, resp.Error
}

func (m *MockS3Client) DeleteObject(in *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	m.init()

	resp := m.DeleteObjectResp[*in.Key]

	delete(m.GetObjectResp, *in.Key)

	if resp == nil {
		return &s3.DeleteObjectOutput{}, nil
	}
	return resp.Resp, resp.Error
}
