package main

import (
	"errors"
	"fmt"
	"mime"
	"path/filepath"

	"github.com/bradfitz/gomemcache/memcache"
)

// CacheService store the marked cache file in memcached and get the corresponding marked file
type CacheService struct {
	MemcachedClient *memcache.Client
}

// NewCacheService generate a new storage service
func NewCacheService(cacheServer ...string) *CacheService {
	//create a handle
	client := memcache.New(cacheServer...)
	if client == nil {
		// Todo: add log for memcache initialize error
		fmt.Println("Fail to connect to the memcache server")
	}

	return &CacheService{MemcachedClient: client}
}

// AddImage add an image file to the memcached
func (svc *CacheService) AddImage(key string, img []byte) error {
	fmt.Println("Start memcache save")
	imgItem := memcache.Item{Key: key, Value: img}
	return svc.MemcachedClient.Add(&imgItem)
}

// GetImage get an image file from the memcached
func (svc *CacheService) GetImage(key string) ([]byte, string, error) {
	//get key's value
	it, err := svc.MemcachedClient.Get(key)
	if err != nil {
		return nil, "", err
	}

	mimeType := mime.TypeByExtension(filepath.Ext(key))
	if string(it.Key) != key {
		return nil, "", errors.New("Unknown Error in memcached for " + key)
	}

	return it.Value, mimeType, nil
}
