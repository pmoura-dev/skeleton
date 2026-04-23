package stdtasks

import (
	"context"
	"net/http"
)

type InternalHTTPTask struct {
	server *http.Server
}

func NewInternalHTTPTask(addr string) *InternalHTTPTask {
	t := &InternalHTTPTask{}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", t.healthzHandler)

	t.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return t
}

func (t *InternalHTTPTask) Start() error {
	if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (t *InternalHTTPTask) Shutdown(ctx context.Context) error {
	return t.server.Shutdown(ctx)
}

func (t *InternalHTTPTask) healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
