package main

import (
	"time"

	"github.com/pmoura-dev/skeleton"
)

func main() {

	app := skeleton.New(skeleton.Config{
		ServiceName:     "helloworld",
		InternalAddr:    "localhost:8081",
		ShutdownTimeout: 5 * time.Second,
	})

	_ = app.Run()
}
