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

type linuxWatchdogTimer struct {
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

func (w *linuxWatchdogTimer) hasFeature(value uint32) bool {
	return w.info != nil && w.info.options&value == value
}

func openWatchdogTimer(name string) (*linuxWatchdogTimer, error) {
	file, err := os.OpenFile(name, os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &linuxWatchdogTimer{
		handle: file,
		info:   getWatchdogInfo(file),
	}, nil
}

func (w *linuxWatchdogTimer) Close() error {
	return w.handle.Close()
}

func (w *linuxWatchdogTimer) Pat() (err error) {
	if w.hasFeature(WDIOF_KEEPALIVEPING) {
		err = unix.IoctlSetInt(
			int(w.handle.Fd()), unix.WDIOC_KEEPALIVE, 1)
	} else {
		_, err = w.handle.Write([]byte("P"))
	}
	return
}

func (w *linuxWatchdogTimer) Disable() error {
	if !w.hasFeature(WDIOF_MAGICCLOSE) {
		return ErrUnsupported
	}

	_, err := w.handle.Write([]byte("V"))
	return err
}

func (w *linuxWatchdogTimer) SetTimeout(seconds time.Duration) error {
	if !w.hasFeature(WDIOF_SETTIMEOUT) {
		return ErrUnsupported
	}

	return unix.IoctlSetPointerInt(
		int(w.handle.Fd()), unix.WDIOC_SETTIMEOUT,
		int(seconds/time.Second))
}

func (w *linuxWatchdogTimer) GetTimeout() (seconds time.Duration, err error) {
	s, err := unix.IoctlGetInt(int(w.handle.Fd()), unix.WDIOC_GETTIMEOUT)
	if err != nil {
		return
	}

	return time.Duration(s) * time.Second, nil
}

func (w *linuxWatchdogTimer) GetTimeLeft() (seconds time.Duration, err error) {
	s, err := unix.IoctlGetInt(int(w.handle.Fd()), unix.WDIOC_GETTIMELEFT)
	if err != nil {
		return
	}

	return time.Duration(s) * time.Second, nil
}
