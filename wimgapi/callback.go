//go:build windows

package wimgapi

import (
	"sync"
	"sync/atomic"
	"syscall"
)

type callbackState struct {
	fn ProgressFunc
}

var (
	callbackProc = syscall.NewCallback(wimMessageCallback)
	callbackSeq  atomic.Uintptr
	callbacks    sync.Map // map[uintptr]*callbackState
)

func newCallbackState(fn ProgressFunc) uintptr {
	id := callbackSeq.Add(1)
	callbacks.Store(id, &callbackState{fn: fn})
	return id
}

func deleteCallbackState(id uintptr) {
	callbacks.Delete(id)
}

func wimMessageCallback(messageID, wParam, lParam, userData uintptr) uintptr {
	v, ok := callbacks.Load(userData)
	if !ok {
		return WIMCallbackSuccess
	}

	state, ok := v.(*callbackState)
	if !ok || state.fn == nil {
		return WIMCallbackSuccess
	}

	cancel := state.fn(ProgressEvent{
		MessageID: uint32(messageID),
		WParam:    wParam,
		LParam:    lParam,
	})
	if cancel {
		return WIMCallbackAbortResult
	}
	return WIMCallbackSuccess
}
