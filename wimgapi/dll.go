//go:build windows

package wimgapi

import (
	"encoding/binary"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

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

func IsInvalidHandle(h windows.Handle) bool {
	return h == 0 || h == windows.InvalidHandle
}
func BytesFromPointer(ptr uintptr, size uint32) []byte {
	if ptr == 0 || size == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size)
}

// DecodeUTF16Bytes decodes UTF-16LE bytes (optionally BOM-prefixed) into UTF-8 string.
func DecodeUTF16Bytes(b []byte) string {
	if len(b) < 2 {
		return ""
	}

	start := 0
	if len(b) >= 2 {
		bom := binary.LittleEndian.Uint16(b[:2])
		if bom == 0xFEFF {
			start = 2
		}
	}

	u16 := make([]uint16, 0, (len(b)-start)/2)
	for i := start; i+1 < len(b); i += 2 {
		v := binary.LittleEndian.Uint16(b[i : i+2])
		if v == 0 {
			break
		}
		u16 = append(u16, v)
	}

	return string(utf16.Decode(u16))
}
