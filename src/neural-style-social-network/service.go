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

// Followee product followee information
type Followee struct {
	ProductID string `json:"productid"`
	UserID    string `json:"userid"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
}

// Service define basic service interface for social network
type Service interface {
	GetReviewsByProductID(id string) ([]Review, error)
	AddReviewByProductID(review Review) error
	GetFolloweesByProductID(id string) ([]Followee, error)
	AddFolloweesByProductID(use Followee) error
	DeleteFolloweeByID(productID, UserID string) error
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
		level.Error(svc.Logger).Log("API", "GetReviewsByProductID", "info", err.Error())
	}

	level.Debug(svc.Logger).Log("API", "AddReviewByProductID", "user", review.User, "id", review.ProductID, "comments", review.Comment)
	return err
}

// GetFolloweesByProductID get all the followees for a given product id
func (svc *SocialService) GetFolloweesByProductID(id string) ([]Followee, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("followees")

	var followees []Followee
	err := c.Find(bson.M{"productid": id}).All(&followees)
	if err != nil {
		level.Error(svc.Logger).Log("API", "GetFolloweesByProductID", "productid", id, "info", err.Error())
	}

	if len(followees) != 0 {
		return followees, nil
	}

	level.Debug(svc.Logger).Log("API", "GetFolloweesByProductID", "productid", id, "size", len(followees))
	return followees, nil
}

// AddFolloweesByProductID add new followees to a given product id
func (svc *SocialService) AddFolloweesByProductID(info Followee) error {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("followees")

	err := c.Insert(info)

	if err != nil {
		level.Error(svc.Logger).Log("API", "AddFolloweesByProductID", "user", info.Name, "productid", info.ProductID, "info", err.Error())
	}

	level.Debug(svc.Logger).Log("API", "AddFolloweesByProductID", "user", info.UserID, "id", info.ProductID)
	return nil
}

// DeleteFolloweeByID remove the followee information for a given product id
func (svc *SocialService) DeleteFolloweeByID(productID, UserID string) error {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("followees")

	err := c.Remove(bson.M{"productid": productID, "userid": UserID})
	if err != nil {
		level.Error(svc.Logger).Log("API", "DeleteFolloweeByIDProductID", "productid", productID, "userid", UserID, "info", err.Error())
		return err
	}

	level.Debug(svc.Logger).Log("API", "DeleteFolloweeByIDProductID", "productid", productID, "userid", UserID)
	return err
}
