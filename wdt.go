package watchdogtimer

import (
	"errors"
	"io"
	"time"
)

var (
	// ErrUnsupported is returned if the given operation is not supported.
	ErrUnsupported = errors.New("disable not supported")
)

type WatchdogTimer interface {
	io.Closer

	Pat() error

	Disable() error

	SetTimeout(seconds time.Duration) error

	GetTimeout() (seconds time.Duration, err error)

	GetTimeLeft() (seconds time.Duration, err error)
}

// Open the named platform specific Watchdog timer.
func Open(name string) (WatchdogTimer, error) {
	return openWatchdogTimer(name)
}
