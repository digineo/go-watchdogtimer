// +build !linux

package watchdogtimer

func openWatchdogTimer(name string) (*timer, error) {
	return nil, ErrUnsupported
}
