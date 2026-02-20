//go:build windows

package wimgapi

import (
	"sync"

	"golang.org/x/sys/windows"
)

const (
	winMessageApp = 0x8000

	WIMCreateNew    = 1
	WIMCreateAlways = 2
	WIMOpenExisting = 3
	WIMOpenAlways   = 4

	WIMMessageBase     = winMessageApp + 0x1476
	WIMMessageText     = WIMMessageBase + 1  //38007
	WIMMessageProgress = WIMMessageBase + 2  //38008
	WIMMessageProcess  = WIMMessageBase + 3  //38009
	WIMMessageScanning = WIMMessageBase + 4  //38010
	WIMMessageSetRange = WIMMessageBase + 5  //38011 将要 captured/applied 的文件总数
	WIMMessageSetPos   = WIMMessageBase + 6  //38012 已经 captured/applied 的文件数量
	WIMMessageStepIt   = WIMMessageBase + 7  //38013 有一个文件被 captured/applied
	WIMMessageCompress = WIMMessageBase + 8  //38014
	WIMMessageError    = WIMMessageBase + 9  //38015
	WIMMessageAlign    = WIMMessageBase + 10 //38016
	WIMMessageRetry    = WIMMessageBase + 11 //38017
	WIMMessageSplit    = WIMMessageBase + 12 //38018
	WIMMessageFileInfo = WIMMessageBase + 13 //38019
	WIMMessageInfo     = WIMMessageBase + 14 //38020
	WIMMessageWarning  = WIMMessageBase + 15 //38021
	WIMMessageChkProc  = WIMMessageBase + 16 //38022

	WIMMessageDone          = 0xFFFFFFF0
	WIMInvalidCallbackValue = 0xFFFFFFFF
	WIMCallbackAbortResult  = 0xFFFFFFFF
	WIMCallbackSuccess      = 0
)

type OpenOptions struct {
	DesiredAccess       uint32
	CreationDisposition uint32
	FlagsAndAttributes  uint32
	CompressionType     uint32
}

type ApplyOptions struct {
	Flags    uint32
	Progress ProgressFunc
}

type CaptureOptions struct {
	Flags    uint32
	Progress ProgressFunc
}

type ProgressEvent struct {
	MessageID uint32
	WParam    uintptr
	LParam    uintptr
}

type ProgressFunc func(evt ProgressEvent) (cancel bool)

type File struct {
	handle windows.Handle
	mu     sync.Mutex
	closed bool
}

type Image struct {
	handle     windows.Handle
	fileHandle windows.Handle
	mu         sync.Mutex
	closed     bool
}

type ImageInfo struct {
	Index        int
	Name         string
	Description  string
	Flags        string
	Architecture string
}

func normalizeOpenOptions(opts OpenOptions) OpenOptions {
	if opts.DesiredAccess == 0 {
		opts.DesiredAccess = windows.GENERIC_READ
	}
	if opts.CreationDisposition == 0 {
		opts.CreationDisposition = WIMOpenExisting
	}
	return opts
}
