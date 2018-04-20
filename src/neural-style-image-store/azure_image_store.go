package ImageStoreService

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
)

// In Global the StorageURL is .blob.core.windows.net
// In China the StorageURL is .blob.core.chinacloudapi.cn

// AzureImageStore store the image on Azure storage
type AzureImageStore struct {
	StorageAccount string
	StorageKey     string
	StorageURL     string
}

// Save image on azure storage
func (svc *AzureImageStore) Save(img Image) error {
	// Create a default request pipeline using your storage account name and account key.
	credential := azblob.NewSharedKeyCredential(svc.StorageAccount, svc.StorageKey)
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// From the Azure portal, get your storage account blob service URL endpoint.
	blobURL := "https://" + "%s" + svc.StorageURL + "/%s"
	blobURL = fmt.Sprintf(blobURL, svc.StorageAccount, img.UserID)
	URL, _ := url.Parse(blobURL)

	// create container for the user
	containerURL := azblob.NewContainerURL(*URL, p)

	// Create the container
	ctx := context.Background() // This example uses a never-expiring context
	_, err := containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)

	// add the image file as a blob to the container
	imgBlobURL := containerURL.NewBlockBlobURL(filepath.Ext(img.Location))
	file, err := os.Open(img.ParentPath + img.Location)

	// The high-level API UploadFileToBlockBlob function uploads blocks in parallel for optimal performance, and
	// can handle large files as well.
	// This function calls PutBlock/PutBlockList for files larger 256 MBs, and calls PutBlob for any file smaller
	_, err = azblob.UploadFileToBlockBlob(ctx, file, imgBlobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16})

	if err != nil {
		return err
	}

	return nil
}

// Find the selected image from id
func (svc *AzureImageStore) Find(userID, fileName string) (string, error) {
	credential := azblob.NewSharedKeyCredential(svc.StorageAccount, svc.StorageKey)

	// return the blob url for the end user
	// Set the desired SAS signature values and sign them with the shared key credentials to get the SAS query parameters.
	sasQueryParams := azblob.BlobSASSignatureValues{
		Protocol:      azblob.SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(48 * time.Hour), // 48-hours before expiration
		Permissions:   azblob.AccountSASPermissions{Read: true}.String(),
		ContainerName: userID,
		BlobName:      fileName,
	}.NewSASQueryParameters(credential)

	qp := sasQueryParams.Encode()
	if len(qp) == 0 {
		return "", errors.New(fileName + "doesn't exist")
	}

	publicblobURL := "https://%s" + svc.StorageURL + "?%s"
	publicblobURL = fmt.Sprintf(publicblobURL, svc.StorageAccount, qp)

	return publicblobURL, nil
}

// FindAllByUser return all the image for a selected user
func (svc *AzureImageStore) FindAllByUser(userID string) ([]string, error) {
	var blobsURL []string

	credential := azblob.NewSharedKeyCredential(svc.StorageAccount, svc.StorageKey)
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	blobURL := "https://" + "%s" + svc.StorageURL + "/%s"
	blobURL = fmt.Sprintf(blobURL, svc.StorageAccount, userID)
	URL, _ := url.Parse(blobURL)

	containerURL := azblob.NewContainerURL(*URL, p)

	for marker := (azblob.Marker{}); marker.NotDone(); {
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := containerURL.ListBlobs(ctx, marker, azblob.ListBlobsOptions{})

		// ListBlobs returns the start of the next segment; you MUST use this to get
		// the next segment (after processing the current result segment).
		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Blobs.Blob {
			sasQueryParams := azblob.BlobSASSignatureValues{
				Protocol:      azblob.SASProtocolHTTPS,
				ExpiryTime:    time.Now().UTC().Add(48 * time.Hour), // 48-hours before expiration
				Permissions:   azblob.AccountSASPermissions{Read: true}.String(),
				ContainerName: userID,
				BlobName:      blobInfo.Name,
			}.NewSASQueryParameters(credential)

			qp := sasQueryParams.Encode()
			if len(qp) == 0 {
				return "", errors.New(fileName + "doesn't exist")
			}

			publicblobURL := "https://%s" + svc.StorageURL + "?%s"
			publicblobURL = fmt.Sprintf(publicblobURL, svc.StorageAccount, qp)

			append(blobsURL, publicblobURL)
		}
	}
	
	if len(blobsURL) != 0 {
		return blobsURL, nil
	}
	else {
		return nil, errors.New("Unknow user")
	}
}
