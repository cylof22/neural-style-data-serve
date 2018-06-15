package main

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type loggingService struct {
	logger        log.Logger
	socialService Service
}

func newLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (svc *loggingService) GetReviewsByProductID(id string) (reviews []Review, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetReviewsByProductID", "id", id,
			"took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.socialService.GetReviewsByProductID(id)
}

func (svc *loggingService) AddReviewByProductID(review Review) (err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "AddReviewByProductID", "user", review.User,
			"took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.socialService.AddReviewByProductID(review)
}
