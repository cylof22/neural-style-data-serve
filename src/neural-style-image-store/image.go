package ImageStoreService

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Image define the basic information for the uploaded image
type Image struct {
	ID          string
	UserID      string
	Name        string
	Location    string
	Size        int64
	CreatedAt   time.Time
	Description string
	ParentPath  string
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
	ext, valid := mimeExtensions[mimeType]
	if !valid {
		return "", errors.New("Unsupported image type" + mimeType)
	}

	// Get a name from the URL
	img.Name = filepath.Base(imageURL)
	img.Location = img.ID + ext

	imgPath := img.ParentPath + img.Location
	savedFile, err := os.Create(imgPath)

	if err != nil {
		return "", err
	}
	defer savedFile.Close()

	size, err := io.Copy(savedFile, response.Body)
	if err != nil {
		return "", err
	}

	img.Size = size

	return imgPath, nil
}

// CreateFromFile create the local image on sever from a upload file
func (img *Image) CreateFromFile(file multipart.File, headers *multipart.FileHeader) (string, error) {
	img.Name = headers.Filename
	img.Location = img.ID + filepath.Ext(img.Name)

	savedPath := img.ParentPath + img.Location
	savedFile, err := os.Create(savedPath)
	if err != nil {
		return "", err
	}
	defer savedFile.Close()

	size, err := io.Copy(savedFile, file)
	if err != nil {
		return "", err
	}
	img.Size = size

	return savedPath, nil
}
