package skeleton

import (
	"net/http"
)

// healthzHandler implements a simple health check endpoint.
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
