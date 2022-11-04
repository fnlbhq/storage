package storage

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"strings"
	"time"
)

func AmazonS3(c Credentials) Provider {
	// Define the parameters for the session you want to create.
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(c.Key, c.Secret, ""), // Specifies your credentials.
		Endpoint:    aws.String(c.Endpoint),                                // Find your endpoint in the control panel, under Settings. Prepend "https://".
		Region:      aws.String(c.Region),                                  // Must be "us-east-1" when creating new Spaces. Otherwise, use the region in your endpoint, such as "nyc3".
	}
	// Step 3: The new session validates your request and directs it to your
	// Space's specified endpoint using the AWS SDK.
	svc, err := session.NewSession(s3Config)
	if err != nil {
		panic(err)
	}
	return amazonS3{client: s3.New(svc), defaultACL: "private"}
}

type amazonS3 struct {
	client     *s3.S3
	defaultACL string
}

func (a amazonS3) GetBucket(name string) (Bucket, error) {
	b := &bucket{
		name:       name,
		client:     a.client,
		defaultACL: a.defaultACL,
	}
	// create the Bucket
	if err := b.createBucket(); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeBucketAlreadyExists {
				return b, nil
			} else {
				return nil, err
			}
		}
	}
	return b, nil
}

func (a amazonS3) DeleteBucket(name string) error {
	params := &s3.DeleteBucketInput{Bucket: aws.String(name)}
	_, err := a.client.DeleteBucket(params)
	return err
}

// Bucket

type bucket struct {
	name       string
	defaultACL string
	client     *s3.S3
}

func (b bucket) Name() string {
	return b.name
}

//func (b bucket) exists() bool {
//	svc := s3.New(session.New(&b.client.Config))
//	input := &s3.HeadBucketInput{
//		Bucket: aws.String(b.name),
//	}
//	result, err := svc.HeadBucket(input)
//	if err != nil {
//		if aerr, ok := err.(awserr.Error); ok {
//			switch aerr.Code() {
//			case s3.ErrCodeNoSuchBucket:
//				return false
//			}
//		}
//	}
//	fmt.Println(result)
//	return true
//}

func (b bucket) Save(object *Object) error {
	_, err := b.client.PutObject(&s3.PutObjectInput{
		Bucket:   aws.String(b.name),
		Key:      aws.String(object.Key),
		Body:     strings.NewReader(string(object.Payload)),
		ACL:      aws.String("private"),
		Metadata: object.MetaData,
	})
	return err
}

func (b bucket) Get(key string) (*Object, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	}
	result, err := b.client.GetObject(input)
	if err != nil {
		return nil, err
	}
	bytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	return &Object{
		Bucket:  b.name,
		Key:     key,
		Payload: bytes,
	}, nil
}

func (b bucket) Delete(key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	}
	_, err := b.client.DeleteObject(input)
	return err
}

func (b bucket) createBucket() (err error) {
	params := &s3.CreateBucketInput{Bucket: aws.String(b.name)}
	_, err = b.client.CreateBucket(params)
	return
}

func (b bucket) deleteBucket() (err error) {
	params := &s3.DeleteBucketInput{Bucket: aws.String(b.name)}
	_, err = b.client.DeleteBucket(params)
	return
}

func (b bucket) Keys() ([]string, error) {
	// Get the list of items
	resp, err := b.client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: &b.name})
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, item := range resp.Contents {
		//fmt.Println("Name:          ", *item.Key)
		//fmt.Println("Last modified: ", *item.LastModified)
		//fmt.Println("Size:          ", *item.Size)
		//fmt.Println("Storage class: ", *item.StorageClass)
		//fmt.Println("")
		keys = append(keys, *item.Key)
	}
	return keys, nil
}

func (b bucket) Move(originKey, destinationKey string) error {
	var obj *Object
	var err error
	if obj, err = b.Get(originKey); err != nil {
		return err
	}
	obj.Key = destinationKey
	if err = b.Save(obj); err != nil {
		return err
	}
	return b.Delete(originKey)
}

func (b bucket) DownloadURL(key string) (string, error) {
	req, _ := b.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	return req.Presign(5 * time.Minute)
}

func (b bucket) DefaultACL() string {
	return b.defaultACL
}
