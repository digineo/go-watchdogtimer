package watchdogtimer

import (
	"os"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	// Default watchdog path
	DefaultWatchdogPath = "/dev/watchdog"

	// Set timeout (in seconds)
	WDIOF_SETTIMEOUT = 0x0080

	// Supports magic close char
	WDIOF_MAGICCLOSE = 0x0100

	// Keep alive ping reply
	WDIOF_KEEPALIVEPING = 0x8000
)

type watchdogInfo struct {
	options         uint32
	firmwareVersion uint32
	identity        [32]byte
}

type timer struct {
	handle *os.File

	info *watchdogInfo
}

func getWatchdogInfo(file *os.File) *watchdogInfo {
	info := watchdogInfo{}

	_, _, ep := syscall.Syscall(
		syscall.SYS_IOCTL, uintptr(file.Fd()),
		unix.WDIOC_GETSUPPORT, uintptr(unsafe.Pointer(&info)))
	if ep != 0 {
		return nil
	}

	return &info
}

func (t *timer) hasFeature(value uint32) bool {
	return t.info != nil && t.info.options&value == value
}

func open(name string) (Timer, error) {
	file, err := os.OpenFile(name, os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &timer{
		handle: file,
		info:   getWatchdogInfo(file),
	}, nil
}

func (t *timer) Close() error {
	return t.handle.Close()
}

func (t *timer) Pat() (err error) {
	if t.hasFeature(WDIOF_KEEPALIVEPING) {
		err = unix.IoctlSetInt(
			int(t.handle.Fd()), unix.WDIOC_KEEPALIVE, 1)
	} else {
		_, err = t.handle.Write([]byte("P"))
	}
	return
}

func (t *timer) Disable() error {
	if !t.hasFeature(WDIOF_MAGICCLOSE) {
		return ErrUnsupported
	}

	_, err := t.handle.Write([]byte("V"))
	return err
}

func (t *timer) SetTimeout(seconds time.Duration) error {
	if !t.hasFeature(WDIOF_SETTIMEOUT) {
		return ErrUnsupported
	}

	return unix.IoctlSetPointerInt(
		int(t.handle.Fd()), unix.WDIOC_SETTIMEOUT,
		int(seconds/time.Second))
}

func (t *timer) GetTimeout() (seconds time.Duration, err error) {
	s, err := unix.IoctlGetInt(int(t.handle.Fd()), unix.WDIOC_GETTIMEOUT)
	if err != nil {
		return
	}

	return time.Duration(s) * time.Second, nil
}

func (t *timer) GetTimeLeft() (seconds time.Duration, err error) {
	s, err := unix.IoctlGetInt(int(t.handle.Fd()), unix.WDIOC_GETTIMELEFT)
	if err != nil {
		return
	}

	return time.Duration(s) * time.Second, nil
}
