//go:build windows

package wimgapi

import (
	"unsafe"

	"go-wimgapi/internal/syscallx"
	"golang.org/x/sys/windows"
)

func Open(path string, opts OpenOptions) (*File, error) {
	if err := ensureLoaded(); err != nil {
		return nil, err
	}

	opts = normalizeOpenOptions(opts)
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	var creationResult uint32
	r1, _, callErr := procWIMCreateFile.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(opts.DesiredAccess),
		uintptr(opts.CreationDisposition),
		uintptr(opts.FlagsAndAttributes),
		uintptr(opts.CompressionType),
		uintptr(unsafe.Pointer(&creationResult)),
	)

	h := windows.Handle(r1)
	if syscallx.IsInvalidHandle(h) {
		code := codeFromCallErr(callErr)
		if code == 0 {
			code = lastErrorCode()
		}
		return nil, winError("WIMCreateFile", code)
	}

	return &File{handle: h}, nil
}

func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	r1, _, callErr := procWIMCloseHandle.Call(uintptr(f.handle))
	if r1 == 0 {
		code := codeFromCallErr(callErr)
		if code == 0 {
			code = lastErrorCode()
		}
		return winError("WIMCloseHandle", code)
	}
	f.closed = true
	return nil
}

func (f *File) ImageCount() (int, error) {
	r1, _, callErr := procWIMGetImageCount.Call(uintptr(f.handle))
	if r1 == 0 {
		code := codeFromCallErr(callErr)
		if code != 0 {
			return 0, winError("WIMGetImageCount", code)
		}
	}
	return int(r1), nil
}

func (f *File) LoadImage(index int) (*Image, error) {
	if index < 1 {
		return nil, ErrImageIndexInvalid
	}

	r1, _, callErr := procWIMLoadImage.Call(uintptr(f.handle), uintptr(uint32(index)))
	h := windows.Handle(r1)
	if syscallx.IsInvalidHandle(h) {
		code := codeFromCallErr(callErr)
		if code == 0 {
			code = lastErrorCode()
		}
		return nil, winError("WIMLoadImage", code)
	}

	return &Image{handle: h, fileHandle: f.handle}, nil
}

func (f *File) SetTemporaryPath(path string) error {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	r1, _, callErr := procWIMSetTemporaryPath.Call(
		uintptr(f.handle),
		uintptr(unsafe.Pointer(pathPtr)),
	)
	if r1 == 0 {
		code := codeFromCallErr(callErr)
		if code == 0 {
			code = lastErrorCode()
		}
		return winError("WIMSetTemporaryPath", code)
	}
	return nil
}

func (f *File) Capture(path string, opts CaptureOptions) (*Image, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	var callbackUserData uintptr
	if opts.Progress != nil {
		callbackUserData = newCallbackState(opts.Progress)
		defer deleteCallbackState(callbackUserData)

		r1, _, callErr := procWIMRegisterMessageCallback.Call(
			uintptr(f.handle),
			callbackProc,
			callbackUserData,
		)
		if uint32(r1) == WIMInvalidCallbackValue {
			code := codeFromCallErr(callErr)
			if code == 0 {
				code = lastErrorCode()
			}
			return nil, winError("WIMRegisterMessageCallback", code)
		}

		defer procWIMUnregisterMessageCallback.Call(
			uintptr(f.handle),
			callbackProc,
		)
	}

	r1, _, callErr := procWIMCaptureImage.Call(
		uintptr(f.handle),
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(opts.Flags),
	)
	h := windows.Handle(r1)
	if syscallx.IsInvalidHandle(h) {
		code := codeFromCallErr(callErr)
		if code == 0 {
			code = lastErrorCode()
		}
		return nil, winError("WIMCaptureImage", code)
	}
	return &Image{handle: h, fileHandle: f.handle}, nil
}

func (f *File) Images() ([]ImageInfo, error) {
	count, err := f.ImageCount()
	if err != nil {
		return nil, err
	}

	images := make([]ImageInfo, 0, count)
	for i := 1; i <= count; i++ {
		img, err := f.LoadImage(i)
		if err != nil {
			return nil, err
		}
		info, infoErr := img.Info()
		closeErr := img.Close()
		if infoErr != nil {
			return nil, infoErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		images = append(images, info)
	}
	return images, nil
}
