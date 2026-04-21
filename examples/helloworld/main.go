package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pmoura-dev/skeleton"
)

type happyTask struct {
	intervalDuration time.Duration

	ticker *time.Ticker
}

func (t *happyTask) Start() error {
	fmt.Printf("Starting ticker with interval: %s\n", t.intervalDuration)

	t.ticker = time.NewTicker(t.intervalDuration)

	for range t.ticker.C {
		fmt.Println("Hello, world!")
	}

	return nil
}

func (t *happyTask) Shutdown(ctx context.Context) error {
	t.ticker.Stop()

	fmt.Println("Ticker stopped")

	return nil
}

type faultyTask struct {
	errTimeout time.Duration
}

func (t *faultyTask) Start() error {
	timer := time.NewTimer(t.errTimeout)
	<-timer.C

	return nil
}

func (t *faultyTask) Shutdown(ctx context.Context) error {
	time.Sleep(10 * time.Second)
	return errors.ErrUnsupported
}

func main() {

	app := skeleton.New(skeleton.Config{
		AppName:         "hello_world",
		ShutdownTimeout: 3 * time.Second,
	})

	_ = app.Register("happy", &happyTask{
		intervalDuration: 1 * time.Second,
	})

	_ = app.Register("faulty", &faultyTask{
		errTimeout: 3 * time.Second,
	})

	app.Run()
}
