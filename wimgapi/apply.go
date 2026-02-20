//go:build windows

package wimgapi

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

func (i *Image) Apply(target string, opts ApplyOptions) error {
	targetPtr, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}

	var callbackUserData uintptr
	if opts.Progress != nil {
		callbackUserData = newCallbackState(opts.Progress)
		defer deleteCallbackState(callbackUserData)

		registerHandle := i.handle
		if i.fileHandle != 0 {
			registerHandle = i.fileHandle
		}
		registeredHandle := registerHandle

		r1, _, callErr := procWIMRegisterMessageCallback.Call(uintptr(registerHandle), callbackProc, callbackUserData)
		if uint32(r1) == WIMInvalidCallbackValue && registerHandle != i.handle {
			// Compatibility fallback for implementations expecting image handle.
			r1, _, callErr = procWIMRegisterMessageCallback.Call(uintptr(i.handle), callbackProc, callbackUserData)
			registeredHandle = i.handle
		}
		if uint32(r1) == WIMInvalidCallbackValue {
			code := codeFromCallErr(callErr)
			if code == 0 {
				code = lastErrorCode()
			}
			return winError("WIMRegisterMessageCallback", code)
		}

		defer procWIMUnregisterMessageCallback.Call(
			uintptr(registeredHandle),
			callbackProc,
		)
	}

	r1, _, callErr := procWIMApplyImage.Call(
		uintptr(i.handle),
		uintptr(unsafe.Pointer(targetPtr)),
		uintptr(opts.Flags),
	)
	if r1 == 0 {
		code := codeFromCallErr(callErr)
		if code == 0 {
			code = lastErrorCode()
		}
		return winError("WIMApplyImage", code)
	}
	return nil
}
