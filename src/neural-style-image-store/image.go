package main

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Image define the basic information for the uploaded image
type Image struct {
	UserID    string
	ImageName string
	ImageData []byte
	Location  string
}

// UploadResult define the basic inforation after the image is uploaded to the cloud storage
// The StorageAccount and Storage Key is special for the Azure cloud storage, No such conception for AWS S3 now.
// Todo: How to the store the two value safely in the database
type UploadResult struct {
	UserID         string
	Name           string
	Location       string
	StorageAccount string
	StorageKey     string
	UploadError    error
}

var mimeExtensions = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
}

// CreateImageFromURL get the image from the url
func (img *Image) CreateImageFromURL(imageURL string) (string, error) {
	response, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", errors.New("Bad image URL")
	}

	defer response.Body.Close()

	mimeType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		return "", errors.New("Unsupported image type" + mimeType)
	}

	// Get an extension for the file
	_, valid := mimeExtensions[mimeType]
	if !valid {
		return "", errors.New("Unsupported image type" + mimeType)
	}

	// Get a name from the URL
	imgName := filepath.Base(imageURL)
	img.Location = imgName

	imgPath := img.Location
	savedFile, err := os.Create(imgPath)

	if err != nil {
		return "", err
	}
	defer savedFile.Close()

	_, err = io.Copy(savedFile, response.Body)
	if err != nil {
		return "", err
	}

	return imgPath, nil
}

// CreateFromFile create the local image on sever from a upload file
func (img *Image) CreateFromFile(file multipart.File, headers *multipart.FileHeader) (string, error) {
	imgName := headers.Filename
	img.Location = imgName

	savedPath := img.Location
	savedFile, err := os.Create(savedPath)
	if err != nil {
		return "", err
	}
	defer savedFile.Close()

	_, err = io.Copy(savedFile, file)
	if err != nil {
		return "", err
	}

	return savedPath, nil
}
