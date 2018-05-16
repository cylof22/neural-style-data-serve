package ProductService

import (
	"time"

	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"
)

type loggingService struct {
	logger      log.Logger
	dataService Service
}

// NewLoggingService returns a new instance of a products logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (svc *loggingService) UploadContentFile(productData Product) (prod Product, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "UploadContentFile", "user", prod.Owner,
			"image", prod.URL, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.UploadContentFile(productData)
}

func (svc *loggingService) UploadStyleFile(productData UploadProduct) (prod Product, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "UploadStyleFile", "user", prod.Owner,
			"image", prod.URL, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.UploadStyleFile(productData)
}

func (svc *loggingService) UploadStyleFiles(products BatchProducts) (info string, err error) {
	// Todo: more detail log for the batch upload
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "UploadStyleFiles", "user", products.Owner,
			"took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.UploadStyleFiles(products)
}

func (svc *loggingService) GetProducts() (prods []Product, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetProducts", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetProducts()
}

func (svc *loggingService) GetProductsByUser(userID string) (prods []Product, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetProductsByUser", "owner", userID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetProductsByUser(userID)
}

func (svc *loggingService) GetProductsByTags(tags []string) (prods []Product, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetProductsByTags", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetProductsByTags(tags)
}

func (svc *loggingService) GetProductsByID(id string) (prod Product, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetProductsByID", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetProductsByID(id)
}

func (svc *loggingService) GetReviewsByProductID(id string) (reviews []Review, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetReviewsByProductID", "id", id,
			"took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetReviewsByProductID(id)
}

func (svc *loggingService) GetArtists() (artists []Artist, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetArtists", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetArtists()
}

func (svc *loggingService) GetHotestArtists() (artists []Artist, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetHotestArtists", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetHotestArtists()
}

func (svc *loggingService) GetImage(userID, imageID string) (data []byte, info string, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetImage", "user", userID, "image", imageID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetImage(userID, imageID)
}

func (svc *loggingService) DeleteProduct(productID string) (err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "DeleteProduct", "productID", productID, "took", time.Since(begin), "err", err)
	}(time.Now())
	return svc.dataService.DeleteProduct(productID)
}

func (svc *loggingService) UpdateProduct(productID string, productData UploadProduct) (err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "UpdateProduct", "productID", productID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.UpdateProduct(productID, productData)
}

func (svc *loggingService) Search(keyvals map[string]interface{}) (prods []Product, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "Search", "info", "How to add keyvals", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.Search(keyvals)
}
