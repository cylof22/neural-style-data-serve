package main

import (
	"errors"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Review define the basic elements of the review
type Review struct {
	ID        uint32 `json:"id"`
	ProductID string `json:"productId"`
	Timestamp string `json:"timestamp"`
	User      string `json:"user"`
	Rating    uint8  `json:"rating"`
	Comment   string `json:"comment"`
}

// Service define basic service interface for social network
type Service interface {
	GetReviewsByProductID(id string) ([]Review, error)
	AddReviewByProductID(review Review) error
}

// SocialService define implementation of the social service
type SocialService struct {
	Session *mgo.Session
	Logger  log.Logger
}

func newSocialSVC(logger log.Logger, session *mgo.Session) Service {
	return &SocialService{Session: session, Logger: logger}
}

// GetReviewsByProductID find the
func (svc *SocialService) GetReviewsByProductID(id string) ([]Review, error) {

	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("reviews")

	var reviews []Review
	err := c.Find(bson.M{"productid": id}).All(&reviews)
	if err != nil {
		// Add log information here
		level.Debug(svc.Logger).Log("API", "GetReviewsByProductID", "info", err.Error(), "id", id)
		return reviews, errors.New("Database error")
	}

	if len(reviews) != 0 {
		return reviews, nil
	}

	level.Debug(svc.Logger).Log("API", "GetReviewsByProductID", "info", "get reviews by id ok", "id", id)
	return nil, nil
}

// AddReviewByProductID add review data to the product id
func (svc *SocialService) AddReviewByProductID(review Review) error {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("reviews")

	err := c.Insert(review)

	if err != nil {
		level.Debug(svc.Logger).Log("API", "GetReviewsByProductID", "info", err.Error())
	}

	level.Debug(svc.Logger).Log("API", "AddReviewByProductID", "user", review.User, "id", review.ProductID, "comments", review.Comment)
	return err
}
