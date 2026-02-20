//go:build windows

package wimgapi

import "golang.org/x/sys/windows"

var (
	modWimgapi = windows.NewLazySystemDLL("wimgapi.dll")

	procWIMCreateFile                = modWimgapi.NewProc("WIMCreateFile")
	procWIMCloseHandle               = modWimgapi.NewProc("WIMCloseHandle")
	procWIMGetImageCount             = modWimgapi.NewProc("WIMGetImageCount")
	procWIMLoadImage                 = modWimgapi.NewProc("WIMLoadImage")
	procWIMCaptureImage              = modWimgapi.NewProc("WIMCaptureImage")
	procWIMSetTemporaryPath          = modWimgapi.NewProc("WIMSetTemporaryPath")
	procWIMGetImageInformation       = modWimgapi.NewProc("WIMGetImageInformation")
	procWIMFreeMemory                = modWimgapi.NewProc("WIMFreeMemory")
	procWIMApplyImage                = modWimgapi.NewProc("WIMApplyImage")
	procWIMRegisterMessageCallback   = modWimgapi.NewProc("WIMRegisterMessageCallback")
	procWIMUnregisterMessageCallback = modWimgapi.NewProc("WIMUnregisterMessageCallback")
)

func ensureLoaded() error {
	return modWimgapi.Load()
}

func tryFreeMemory(ptr uintptr) {
	if ptr == 0 {
		return
	}
	if err := procWIMFreeMemory.Find(); err != nil {
		// Some WIMGAPI versions do not export WIMFreeMemory.
		return
	}
	_, _, _ = procWIMFreeMemory.Call(ptr)
}
