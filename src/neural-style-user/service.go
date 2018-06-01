package UserService

import (
	"errors"
	"neural-style-util"
	"os"
	"time"
	"path"
	"strings"
	"strconv"
	"encoding/base64"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	// SecretKey define the private key for generate the JWT token
	SecretKey = os.Getenv("TOKEN_KEY")
)

// UserInfo define the basic user information
type UserInfo struct {
	ID                string   `json:"id"`
	Name              string   `json:"username"`
	Password          string   `json:"password"`
	Phone             string   `json:"phone"`
	Email             string   `json:"email"`
	Portrait          string   `json:"headPortraitUrl"`
	ConcernedProducts []string `json:"concernedProducts"`
	ConcernedUsers    []string `json:"concernedUsers"`
}

// UserToken define the authorization information
type UserToken struct {
	ID    		string `json:"id"`
	Name  		string `json:"username"`
	Token 		string `json:"token"`
	Portrait    string `json:"headPortraitUrl"`
}

// Service define the basic login interface
type Service interface {
	Register(userData UserInfo) (string, error)
	Login(loginData UserInfo) (UserToken, error)
	GetUserInfo(userName string) (UserInfo, error)
	UpdateUserInfo(userData UserInfo) (string, error)
}

// UserService for user login service
type UserService struct {
	Host    string
	Port    string
	Session *mgo.Session
	Logger  log.Logger
}

// NewUserSVC create a new user service
func NewUserSVC(host, port string, logger log.Logger, session *mgo.Session) *UserService {
	return &UserService{Host: host, Port: port, Logger: logger, Session: session}
}

// Register create a new user
func (svc *UserService) Register(userData UserInfo) (string, error) {
	session := svc.Session.Copy()
	defer session.Close()

	// if the user name exists
	var currentUser UserInfo
	c := session.DB("store").C("users")
	err := c.Find(bson.M{"name": userData.Name}).One(&currentUser)
	if err == nil {
		return "", errors.New("User with this name already exists")
	}

	userData.ID = NSUtil.UniqueID()
	result := "Success"
	err = c.Insert(userData)
	if err != nil {
		result = "fail"
		return result, errors.New("Server is busy. Please try later.")
	}

	level.Debug(svc.Logger).Log("API", "Register", "info", userData)
	return result, err
}

// CreateToken create time-limited token
func CreateToken(userName string, log log.Logger) string {
	claims := make(jwt.MapClaims)
	claims["username"] = userName
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() //72小时有效期，过期需要重新登录获取token
	claims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		level.Error(log).Log("API", "CeateToken", "info", "Failed to sign with token", "err", err.Error())
		return ""
	}

	return tokenString
}

// Login login the style transfer platform
func (svc *UserService) Login(loginData UserInfo) (UserToken, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("users")

	var user UserInfo
	err := c.Find(bson.M{"name": loginData.Name}).One(&user)
	if err != nil {
		return UserToken{}, errors.New("This user name is wrong")
	}

	if user.Password != loginData.Password {
		return UserToken{}, errors.New("This password is wrong")
	}

	var userToken UserToken
	userToken.Name = user.Name
	userToken.ID = user.ID
	userToken.Token = CreateToken(loginData.Name, svc.Logger)
	userToken.Portrait = user.Portrait
	return userToken, err
}

func (svc *UserService) GetUserInfo(userName string) (UserInfo, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("users")

	var user UserInfo
	err := c.Find(bson.M{"name": userName}).One(&user)
	if err != nil {
		return user, errors.New("can't get info for the user " + userName)
	}

	return user, nil
}

func (svc *UserService) UpdateUserInfo(userData UserInfo) (string, error) {
	pos := strings.Index(userData.Portrait, "http")
	if pos != 0 {
		imageID := NSUtil.UniqueID()
		newImageURL, err := svc.uploadPicture(userData.Portrait, imageID, "portraits")
		if err != nil {
			return "", errors.New("Server is busy. Please try it later.")
		}
		userData.Portrait = newImageURL
	}

	if userData.Password == "" {
		user, err := svc.GetUserInfo(userData.Name)
		if err != nil {
			return "", errors.New("Server is busy. Please try it later.")
		}
		userData.Password = user.Password
	}

	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("users")

	updateData, err := bson.Marshal(&userData)
	if err != nil {
		return "", errors.New("Server is busy. Please try it later.")
	}
	mData := bson.M{}
	err = bson.Unmarshal(updateData, mData)
	if err != nil {
		return "", errors.New("Server is busy. Please try it later.")
	}

	err = c.Update(bson.M{"name": userData.Name}, bson.M{"$set": mData})
	if err != nil {
		return "", errors.New("Server is busy. Please try it later.")
	}

	return userData.Portrait, nil
}

func (svc *UserService) uploadPicture(picData, picID, picFolder string) (string, error) {
	pos := strings.Index(picData, ",")
	if len(picData) < 11 || pos < 7 {
		level.Debug(svc.Logger).Log("API", "UploadPicture", "info", "Bad Picture Data",
			"DataLength", strconv.Itoa(len(picData)), "sepeatePos", strconv.Itoa(pos))
		return "", errors.New("Bad picture data")
	}

	imgFormat := picData[11 : pos-7]
	realData := picData[pos+1 : len(picData)]

	baseData, err := base64.StdEncoding.DecodeString(realData)
	if err != nil {
		return "", err
	}

	// if the folder exists
	currentFolder := path.Join("./data", picFolder)
	_, err = os.Stat(currentFolder)
	if os.IsNotExist(err) {
		err = os.MkdirAll(currentFolder, 0777)
		if err != nil {
			return "", err
		} 
	}

	outfileName := picID + "." + imgFormat
	// Local FrontEnd Dev version
	outfilePath := path.Join(currentFolder, outfileName)

	outputFile, _ := os.Create(outfilePath)
	defer outputFile.Close()

	outputFile.Write(baseData)

	newImageURL := "http://localhost:8000/" + picFolder + "/" + outfileName
	level.Debug(svc.Logger).Log("Picture URL", newImageURL)
	return newImageURL, nil
}