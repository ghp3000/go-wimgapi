package syscallx

import "unsafe"

func BytesFromPointer(ptr uintptr, size uint32) []byte {
	if ptr == 0 || size == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size)
}
