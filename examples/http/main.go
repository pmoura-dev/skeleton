package main

import (
	"context"
	"net/http"
	"time"

	"github.com/pmoura-dev/skeleton"
	"github.com/pmoura-dev/skeleton/stdtasks"
)

type HTTPServerTask struct {
	server *http.Server
}

func newHTTPServerTask() *HTTPServerTask {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{\"status\": \"alive\"}"))
	})

	return &HTTPServerTask{
		server: &http.Server{
			Addr:    "localhost:8080",
			Handler: mux,
		},
	}
}

func (t *HTTPServerTask) Start() error {
	if err := t.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (t *HTTPServerTask) Shutdown(ctx context.Context) error {
	return t.server.Shutdown(ctx)
}

func main() {

	httpTask := newHTTPServerTask()

	app := skeleton.New(skeleton.Config{
		AppName:         "http-example",
		Version:         "0.0.1",
		ShutdownTimeout: 10 * time.Second,
	})

	app.Register("http-api", httpTask)
	app.Register("http-internal", stdtasks.NewInternalHTTPTask(":8081"))

	app.Run()
}
