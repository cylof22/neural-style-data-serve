package UserService

import (
	"errors"
	"fmt"
	"neural-style-util"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

// UserService for user login service
type UserService struct {
	Host    string
	Port    string
	Session *mgo.Session
}

// NewUserSVC create a new user service
func NewUserSVC(host, port string, session *mgo.Session) *UserService {
	return &UserService{Host: host, Port: port, Session: session}
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

	fmt.Println("Register data is")
	fmt.Println(userData)
	return result, err
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
	userToken.Token = NSUtil.CreateToken(loginData.Name)
	return userToken, err
}
