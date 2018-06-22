package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"neural-style-products"
	"neural-style-user"
	"os"
	"path"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// NewPictureService create new token sale service
func NewPictureService(sess *mgo.Session) *PictureService {
	return &PictureService{Session: sess}
}

// PictureService only available for certain time
type PictureService struct {
	Session *mgo.Session
}

// PictureData upload data info
type PictureData struct {
	WechatID string `json:"wechatid"`
	Address  string `json:"address"`
	PicData  string `json:"picData"`
}

// UniqueID generate picture file name
func UniqueID() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}

	s := base64.URLEncoding.EncodeToString(b)
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func uploadImage2User(info PictureData) UserService.UserInfo {
	var user UserService.UserInfo
	user.Name = info.WechatID
	user.Address = info.Address
	user.WechatID = info.WechatID

	return user
}

func (svc *PictureService) uploadPicture(owner, picData, picID string) (string, error) {
	pos := strings.Index(picData, ",")
	if len(picData) < 11 || pos < 7 {
		return "", errors.New("Bad picture data")
	}

	imgFormat := picData[11 : pos-7]
	realData := picData[pos+1 : len(picData)]

	baseData, err := base64.StdEncoding.DecodeString(realData)
	if err != nil {
		return "", err
	}

	outfileName := owner + "_" + picID + "." + imgFormat
	outfilePath := path.Join("./data", outfileName)

	if *local {
		outputFile, _ := os.Create(outfilePath)
		defer outputFile.Close()

		outputFile.Write(baseData)
	}

	return outfileName, nil
}

func (svc *PictureService) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer func(begin time.Time) {
		fmt.Println("Upload Picture time: ", time.Since(begin))
	}(time.Now())

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Unsupported Method"))
		return
	}

	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("pictures")

	pictureData := PictureData{}
	json.NewDecoder(req.Body).Decode(&pictureData)

	imageID := UniqueID()
	_, err := svc.uploadPicture(pictureData.WechatID, pictureData.PicData, imageID)

	count, err := c.Find(bson.M{"wechatid": pictureData.WechatID}).Count()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}

	if count < 5 {
		err = c.Insert(pictureData)
		if err != nil {
			res.Write([]byte(err.Error()))
			res.WriteHeader(http.StatusInternalServerError)
		}

		//conver to user
		user := uploadImage2User(pictureData)
		userInfo, err := json.Marshal(user)
		if err != nil {
			res.Write([]byte(err.Error()))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// register as a user for elforce
		resp, err := http.Post(*productSite+"/api/v1/register", "application/json", bytes.NewBuffer(userInfo))
		if err != nil {
			res.Write([]byte(err.Error()))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// login
		resp, err = http.Post(*productSite+"/api/v1/authenticate", "application/json", bytes.NewBuffer(userInfo))
		if err != nil {
			res.Write([]byte(err.Error()))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		var token UserService.UserToken
		err = json.NewDecoder(resp.Body).Decode(&token)

		// upload the picture data to the products storage
		prodClient := &http.Client{}
		prodURL := *productSite + "/api/upload/style"

		uploadProd := ProductService.UploadProduct{
			Owner:   user.Name,
			PicData: pictureData.PicData,
		}

		uploadInfo, err := json.Marshal(uploadProd)
		if err != nil {
			res.Write([]byte(err.Error()))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		uploadReq, err := http.NewRequest("POST", prodURL, bytes.NewBuffer(uploadInfo))
		uploadReq.Header.Set("Authorization", "Bearer "+token.Token)
		resp, err = prodClient.Do(uploadReq)
		if err != nil {
			res.Write([]byte(err.Error()))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
		return
	}

	res.WriteHeader(http.StatusForbidden)
}
