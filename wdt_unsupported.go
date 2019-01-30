// +build !linux

package watchdogtimer

func openWatchdogTimer(name string) (*linuxWatchdogTimer, error) {
	return nil, ErrUnsupported
}
