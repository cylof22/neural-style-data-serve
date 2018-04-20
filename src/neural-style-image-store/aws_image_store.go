package ImageStoreService

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// AwsImageStore store the image on aws-s3
type AwsImageStore struct {
	storageFolder string
}

// Save image on aws-s3
func (svc *AwsImageStore) Save(img *Image) error {
	// Initialize a session that the SDK uses to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// create or find a existing bucket
	s3svc := s3.New(sess)
	_, err := s3svc.CreateBucket(&s3.CreateBucketInput{Bucket: aws.String(img.UserID)})

	// Todo: How to check the bucket existence
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				// Todo: bucket name needs to be redefined
				return errors.New("user id duplicated, please use a new unique id")
			case s3.ErrCodeObjectAlreadyInActiveTierError:
				return errors.New("S3 Authorization error")
			}
		}
	} else {
		err = s3svc.WaitUntilBucketExists(&s3.HeadBucketInput{
			Bucket: aws.String(img.UserID)})
		if err != nil {
			waitErr := fmt.Sprintf("Error occurred while waiting for bucket to be created, %v", img.UserID)
			return errors.New(waitErr)
		}

		// Everything we post to the S3 bucket should be marked 'private'
		acl := "private"
		params := &s3.PutBucketAclInput{
			Bucket: &(img.UserID),
			ACL:    &acl,
		}

		// Set bucket ACL
		_, err := s3svc.PutBucketAcl(params)
		if err != nil {
			return err
		}
	}

	f, err := os.Open(img.ParentPath + img.Location)
	if err != nil {
		return errors.New("Failed to open the file")
	}
	defer f.Close()

	uploader := s3manager.NewUploader(sess)

	// Upload the file's body to S3 bucket as an object with the key being the
	// same as the filename.
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(img.UserID),

		// Can also use the `filepath` standard library package to modify the
		// filename as need for an S3 object key. Such as turning absolute path
		// to a relative path.
		Key: aws.String(filepath.Ext(img.Location)),

		// The file to be uploaded. io.ReadSeeker is preferred as the Uploader
		// will be able to optimize memory when uploading large content. io.Reader
		// is supported, but will require buffering of the reader's bytes for
		// each part.
		Body: f,
	})
	if err != nil {
		// Print the error and exit.
		uploadErr := fmt.Sprintf("Unable to upload %q to %q, %v", "filename", img.UserID, err)
		return errors.New(uploadErr)
	}

	return nil
}

// Find the selected image from id
func (svc *AwsImageStore) Find(userID, fileName string) (string, error) {
	return "", nil
}

// FindAllByUser return all the image for a selected user
func (svc *AwsImageStore) FindAllByUser(userID string) ([]string, error) {

	return nil, nil
}
