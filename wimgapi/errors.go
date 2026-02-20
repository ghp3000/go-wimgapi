//go:build windows

package wimgapi

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sys/windows"
)

type Error struct {
	Op   string
	Code uint32
	Msg  string
}

func (e *Error) Error() string {
	if e.Msg == "" {
		return fmt.Sprintf("%s failed with code=%d", e.Op, e.Code)
	}
	return fmt.Sprintf("%s failed with code=%d: %s", e.Op, e.Code, e.Msg)
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

var (
	ErrAccessDenied      = &Error{Code: uint32(windows.ERROR_ACCESS_DENIED)}
	ErrInvalidParameter  = &Error{Code: uint32(windows.ERROR_INVALID_PARAMETER)}
	ErrPathNotFound      = &Error{Code: uint32(windows.ERROR_PATH_NOT_FOUND)}
	ErrFileNotFound      = &Error{Code: uint32(windows.ERROR_FILE_NOT_FOUND)}
	ErrImageIndexInvalid = errors.New("wimgapi: image index must be >= 1")
)

func winError(op string, code uint32) error {
	return &Error{
		Op:   op,
		Code: code,
		Msg:  formatMessage(code),
	}
}

func codeFromCallErr(err error) uint32 {
	if err == nil {
		return 0
	}
	if errno, ok := err.(windows.Errno); ok {
		return uint32(errno)
	}
	return 0
}

func lastErrorCode() uint32 {
	err := windows.GetLastError()
	if err == nil {
		return 0
	}
	if errno, ok := err.(windows.Errno); ok {
		return uint32(errno)
	}
	return 0
}

func formatMessage(code uint32) string {
	if code == 0 {
		return ""
	}

	flags := uint32(windows.FORMAT_MESSAGE_FROM_SYSTEM | windows.FORMAT_MESSAGE_IGNORE_INSERTS)
	buf := make([]uint16, 512)
	n, err := windows.FormatMessage(flags, 0, code, 0, buf, nil)
	if err != nil || n == 0 {
		return ""
	}
	return strings.TrimSpace(windows.UTF16ToString(buf[:n]))
}
