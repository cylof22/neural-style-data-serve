package main

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// StorageService define the basic storage service
type StorageService struct {
	dbSession *mgo.Session
}

// StorageInfo define the azure storage account information
type StorageInfo struct {
	Key     string
	Account string
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
	defer close(resultChannel)

	imgJob := ImageJob{
		UploadImage:   img,
		ResultChannel: resultChannel,
	}
	JobQueue <- imgJob

	resultInfo := <-resultChannel
	if resultInfo.UploadError != nil {
		return resultInfo.UploadError
	}

	// update the upload result to the database: {userID + Name : StorageAccount}
	session := svc.dbSession.Copy()
	defer session.Close()

	c := session.DB("store").C("storage")

	info := StorageInfo{
		Key:     resultInfo.UserID + resultInfo.Name,
		Account: resultInfo.StorageAccount,
	}
	return c.Insert(info)
}

// Find return the public access url for downloading the image file during a limited time
func (svc *StorageService) Find(userID, imgName string) (string, error) {
	session := svc.dbSession.Copy()
	defer session.Close()

	key := userID + imgName
	c := session.DB("store").C("storage")

	var info StorageInfo
	// find the StorageAccount for the key: userID + imgName from the database
	err := c.Find(bson.M{"key": key}).One(&info)
	if err != nil {
		return "", err
	}

	// get the shared access url from the azure storage
	url, err := Stores[info.Account].Find(userID, imgName)
	if err != nil {
		return "", err
	}

	return url, err
}
