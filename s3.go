package gsession

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type S3Storage struct {
	S3        s3iface.S3API
	Bucket    string
	Prefix    string
	Marshaler Marshaler
}

func (s *S3Storage) Save(ctx context.Context, id string, values map[interface{}]interface{}) error {
	data, err := s.Marshaler.Marshal(values)
	if err != nil {
		return fmt.Errorf("Marshaler.Marshal: %v", err)
	}

	r := &s3.PutObjectInput{
		Bucket:        aws.String(s.Bucket),
		Key:           aws.String(s.Prefix + id),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   aws.String(s.Marshaler.ContentType()),
	}
	if _, err := s.S3.PutObjectWithContext(ctx, r); err != nil {
		return fmt.Errorf("s3.PutObjectWithContext: %v", err)
	}

	return nil
}

func (s *S3Storage) Load(ctx context.Context, id string) (map[interface{}]interface{}, error) {
	r := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Prefix + id),
	}
	obj, err := s.S3.GetObjectWithContext(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("s3.GetObjectWithContext: %v", err)
	}
	defer obj.Body.Close()

	body, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %v", err)
	}

	values, err := s.Marshaler.Unmarshal(body)
	if err != nil {
		return nil, fmt.Errorf("Marshaler.Unmarshal: %v", err)
	}

	return values, nil
}

func (s *S3Storage) Delete(ctx context.Context, id string) error {
	r := &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Prefix + id),
	}
	if _, err := s.S3.DeleteObjectWithContext(ctx, r); err != nil {
		return fmt.Errorf("s3.DeleteObjectWithContext: %v", err)
	}
	return nil
}
