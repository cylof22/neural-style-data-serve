package main

import (
	"time"

	"github.com/go-kit/kit/log"
)

type loggingService struct {
	logger       log.Logger
	storeService Service
}

// NewLoggingService generate the log service for cloud storage service
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (svc *loggingService) Save(userID, imgName string, imgData []byte) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "Save", "user", userID,
			"image", imgName, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.storeService.Save(userID, imgName, imgData)
}

func (svc *loggingService) Find(userID, imgName string) (url string, err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "Find", "user", userID,
			"image", imgName, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.storeService.Find(userID, imgName)
}
