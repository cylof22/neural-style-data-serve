package main

import (
	"encoding/json"
	"net/http"

	"neural-style-user"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// NewTokenPreSale create new token sale service
func NewTokenPreSale(sess *mgo.Session) *TokenSaleService {
	return &TokenSaleService{Session: sess}
}

// TokenSaleService only available for certain time
type TokenSaleService struct {
	Session *mgo.Session
}

const (
	artist = iota
	photographer
)

// TokenSaleInfo define the basic information of the token sale info
type TokenSaleInfo struct {
	Address    string `json:"address"`
	WechatID   string `json:"wechatid"`
	TelegramID string `json:"telegramid"`
	Mail       string `json:"mail"`
	Phone      string `json:"phone"`
	Name       string `json:"name"`
	Profession string `json:"profession"`
}

func tokenSale2User(info TokenSaleInfo) UserService.UserInfo {
	var user UserService.UserInfo
	user.Address = info.Address
	user.WechatID = info.WechatID
	user.TelegramID = info.TelegramID
	user.Email = info.Mail
	user.Phone = info.Phone
	user.Name = info.Name

	switch info.Profession {
	case "0":
		user.Profession = artist
	case "1":
		user.Profession = photographer
	}
	// generate unique id
	user.ID = ""

	return user
}

func (svc *TokenSaleService) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Unsupported Method"))
		return
	}

	info := TokenSaleInfo{}
	json.NewDecoder(req.Body).Decode(&info)

	if len(info.Address) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Bad wallet address"))
		return
	}

	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("tokens")

	count, err := c.Find(bson.M{"address": info.Address}).Count()
	if err != nil || count != 0 {
		res.Write([]byte("Duplicated Wallet address"))
		res.WriteHeader(http.StatusBadRequest)
	}

	err = c.Insert(info)
	if err != nil {
		res.Write([]byte(err.Error()))
		res.WriteHeader(http.StatusBadRequest)
	}

	res.WriteHeader(http.StatusOK)
}
