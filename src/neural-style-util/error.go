package NSUtil

import (
	"context"
	"encoding/json"
	"net/http"
)

// EncodeError write the error information to the response
func EncodeError(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("context-type", "application/json,charset=utf8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

// NSError define the basic error information for the response
type NSError struct {
	Info string
	Code int
}

// NewError generate error with string
func NewError(info string) NSError {
	return NSError{Info: info, Code: http.StatusInternalServerError}
}

// NewErrorWithStatus generate error with infor and http status code
func NewErrorWithStatus(status int, info string) NSError {
	return NSError{Code: status, Info: info}
}

// Error implement the error interface
func (err NSError) Error() string {
	return err.Info
}

// StatusCode implement the StatusCoder of gokit
func (err NSError) StatusCode() int {
	return err.Code
}
