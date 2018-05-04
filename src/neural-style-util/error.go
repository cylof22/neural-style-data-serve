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
