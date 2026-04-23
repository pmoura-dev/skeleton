package skeleton

import "time"

type Config struct {
	AppName string
	Version string

	ShutdownTimeout time.Duration
}
