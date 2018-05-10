package UserService

import (
	"errors"
	"neural-style-util"
	"os"
	"time"

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
	ConcernedProducts []string `json:"concernedProducts"`
	ConcernedUsers    []string `json:"concernedUsers"`
}

// UserToken define the authorization information
type UserToken struct {
	ID    string `json:"id"`
	Name  string `json:"username"`
	Token string `json:"token"`
}

// Service define the basic login interface
type Service interface {
	Register(userData UserInfo) (string, error)
	Login(loginData UserInfo) (UserToken, error)
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
		return result, errors.New("Failed to add a new user")
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
	return userToken, err
}
