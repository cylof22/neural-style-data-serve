package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func decodeNSCacheSaveRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	key := vars["key"]

	imgData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return NSCacheSaveRequest{Key: key, Data: imgData}, nil
}

func encodeNSCacheSaveResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	saveRes := response.(NSCacheSaveResponse)
	if saveRes.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(saveRes)
}

func decodeNSCacheGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	key := vars["key"]

	return NSCacheGetRequest{Key: key}, nil
}

func encodeNSCachedGetResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	getRes := response.(NSCacheGetResponse)

	if getRes.Error != nil {
		// Todo: add error log
		w.WriteHeader(http.StatusInternalServerError)
	}

	fmt.Println(getRes.Type)
	w.Header().Set("Content-Type", getRes.Type)
	w.Header().Set("Accept-Ranges", "bytes")
	imgSize := len(getRes.Data)
	w.Header().Set("Content-Length", strconv.FormatInt(int64(imgSize), 10))

	length, err := w.Write(getRes.Data)

	if length != len(getRes.Data) {
		return errors.New("Empty image")
	}

	return err
}
