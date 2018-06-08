package main

import (
	"encoding/json"
	"net/http"

	mgo "gopkg.in/mgo.v2"
)

// NewTokenPreSale create new token sale service
func NewTokenPreSale(sess *mgo.Session) *TokenSaleService {
	return &TokenSaleService{Session: sess}
}

// TokenSaleService only available for certain time
type TokenSaleService struct {
	Session *mgo.Session
}

// TokenSaleInfo define the basic information of the token sale info
type TokenSaleInfo struct {
	Address    string `json:"address"`
	WechatID   string `json:"wechatid"`
	TelegramID string `json:"telegramid"`
	Mail       string `json:"mail"`
	Phone      string `json:"phone"`
	Name       string `json:"name"`
}

func (svc *TokenSaleService) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	info := TokenSaleInfo{}
	json.NewDecoder(req.Body).Decode(&info)

	if len(info.Address) == 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("tokens")

	err := c.Insert(info)
	if err != nil {
		res.Write([]byte(err.Error()))
		res.WriteHeader(http.StatusBadRequest)
	}

	res.WriteHeader(http.StatusOK)
}
