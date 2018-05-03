package main

import mgo "gopkg.in/mgo.v2"

// StorageService define the basic storage service
type StorageService struct {
	dbSession *mgo.Session
}

// NewStorageService generate a new storage service
func NewStorageService(session *mgo.Session) *StorageService {
	return &StorageService{dbSession: session}
}

// Save store the target image file to cloud storage
func (svc *StorageService) Save(userID, imgName string, imgData []byte) error {
	img := Image{
		UserID:    userID,
		ImageName: imgName,
		ImageData: imgData,
	}

	resultChannel := make(chan UploadResult)

	imgJob := ImageJob{
		UploadImage:   img,
		ResultChannel: resultChannel,
	}
	JobQueue <- imgJob

	resultInfo := <-resultChannel
	if resultInfo.UploadError != nil {
		return resultInfo.UploadError
	}

	return nil
}

// Find return the public access url for downloading the image file during a limited time
func (svc *StorageService) Find(userID, imgName string) (string, error) {

	return "", nil
}
