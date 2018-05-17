package NSUtil

import (
	"encoding/base64"
	"strings"
	"path"
	"os"
)
// upload picture file
func UploadPicture(owner, picData, picID, picFolder string, isLocalDev bool) (string, error) {
	pos := strings.Index(picData, ",")
	imgFormat := picData[11 : pos-7]
	realData := picData[pos+1 : len(picData)]

	baseData, err := base64.StdEncoding.DecodeString(realData)
	if err != nil {
		return "", err
	}

	outfileName := picID + "." + imgFormat
	// Local FrontEnd Dev version
	if isLocalDev {
		outfilePath := path.Join("./data", picFolder, outfileName)

		outputFile, _ := os.Create(outfilePath)
		defer outputFile.Close()

		outputFile.Write(baseData)

		newImageURL := "http://localhost:8000/" + picFolder + "/" + outfileName
		return newImageURL, nil
	}

	return "", nil
}