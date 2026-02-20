package syscallx

import "golang.org/x/sys/windows"

func IsInvalidHandle(h windows.Handle) bool {
	return h == 0 || h == windows.InvalidHandle
}
