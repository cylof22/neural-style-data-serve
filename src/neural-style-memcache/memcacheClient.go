package ImageCache

import (
	"errors"

	"github.com/bradfitz/gomemcache/memcache"
)

// Service store the marked cache file in memcached and get the corresponding marked file
type Service struct {
	MemcachedURL    []string
	MemcachedClient *memcache.Client
}

// Init initialize the memcached service
func (svc *Service) Init() {
	//create a handle
	svc.MemcachedClient = memcache.New(svc.MemcachedURL...)
	if svc.MemcachedClient == nil {
		// Todo: add memcache initialize error
	}
}

// AddImage add an image file to the memcached
func (svc *Service) AddImage(key string, img []byte) error {
	imgItem := memcache.Item{Key: key, Value: img}
	return svc.MemcachedClient.Add(&imgItem)
}

// GetImage get an image file from the memcached
func (svc *Service) GetImage(key string) ([]byte, error) {
	//get key's value
	it, err := svc.MemcachedClient.Get("foo")
	if err != nil {
		return nil, err
	}

	if string(it.Key) != key {
		return nil, errors.New("Unknown Error in memcached for " + key)
	}

	return it.Value, nil
}
