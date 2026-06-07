package httpx

import (
	"net/http"

	"github.com/MrEthical07/projectbook-backend/internal/core/params"
)

// URLParam returns route parameter value by key.
//
// This helper keeps routing code decoupled from concrete router implementation.
func URLParam(r *http.Request, key string) string {
	return params.URLParam(r, key)
}
