//go:build windows

package wimgapi

import (
	"fmt"
	"sync"
)

type DecodedProgressEvent struct {
	MessageID uint32
	Name      string
	WParam    uintptr
	LParam    uintptr
	Current   uint64
	Total     uint64
	Percent   float64
	ErrorCode uint32
	Summary   string
	Noisy     bool
}

type ProgressDecoder struct {
	mu      sync.Mutex
	total   uint64
	current uint64
}

func NewProgressDecoder() *ProgressDecoder {
	return &ProgressDecoder{}
}

func (d *ProgressDecoder) Decode(evt ProgressEvent) DecodedProgressEvent {
	d.mu.Lock()
	defer d.mu.Unlock()

	out := DecodedProgressEvent{
		MessageID: evt.MessageID,
		Name:      messageName(evt.MessageID),
		WParam:    evt.WParam,
		LParam:    evt.LParam,
		Current:   d.current,
		Total:     d.total,
	}

	switch evt.MessageID {
	case WIMMessageSetRange:
		// In many environments, LParam carries total range.
		if evt.LParam != 0 {
			d.total = uint64(evt.LParam)
		} else {
			d.total = uint64(evt.WParam)
		}
		if d.current > d.total {
			d.current = d.total
		}
	case WIMMessageSetPos:
		// WParam often carries current position; LParam may carry total at end.
		if evt.WParam != 0 {
			d.current = uint64(evt.WParam)
		}
		if evt.LParam != 0 {
			lp := uint64(evt.LParam)
			if d.total == 0 || lp > d.total {
				d.total = lp
			}
			// Observed pattern: completion SetPos with LParam == total.
			if lp == d.total {
				d.current = d.total
			}
		}
	case WIMMessageStepIt:
		d.current++
	case WIMMessageError, WIMMessageWarning:
		out.ErrorCode = uint32(evt.WParam)
	}

	out.Current = d.current
	out.Total = d.total
	if d.total > 0 {
		out.Percent = (float64(d.current) / float64(d.total)) * 100
	}
	out.Summary = summarize(out)
	out.Noisy = isNoisy(out)
	return out
}

func messageName(id uint32) string {
	switch id {
	case WIMMessageText:
		return "TEXT"
	case WIMMessageProgress:
		return "PROGRESS"
	case WIMMessageProcess:
		return "PROCESS"
	case WIMMessageScanning:
		return "SCANNING"
	case WIMMessageSetRange:
		return "SET_RANGE"
	case WIMMessageSetPos:
		return "SET_POS"
	case WIMMessageStepIt:
		return "STEP_IT"
	case WIMMessageCompress:
		return "COMPRESS"
	case WIMMessageError:
		return "ERROR"
	case WIMMessageAlign:
		return "ALIGN"
	case WIMMessageRetry:
		return "RETRY"
	case WIMMessageSplit:
		return "SPLIT"
	case WIMMessageInfo:
		return "INFO"
	case WIMMessageWarning:
		return "WARNING"
	case WIMMessageFileInfo:
		return "FILE_INFO"
	case WIMMessageChkProc:
		return "CHK_PROCESS"
	case WIMMessageDone:
		return "DONE"
	default:
		return fmt.Sprintf("UNKNOWN(0x%X)", id)
	}
}

func summarize(evt DecodedProgressEvent) string {
	switch evt.MessageID {
	case WIMMessageSetRange, WIMMessageSetPos, WIMMessageStepIt, WIMMessageProgress:
		if evt.Total > 0 {
			return fmt.Sprintf("%s: %d/%d (%.1f%%)", evt.Name, evt.Current, evt.Total, evt.Percent)
		}
		return fmt.Sprintf("%s: current=%d", evt.Name, evt.Current)
	case WIMMessageError:
		return fmt.Sprintf("ERROR: code=%d", evt.ErrorCode)
	case WIMMessageWarning:
		return fmt.Sprintf("WARNING: code=%d", evt.ErrorCode)
	case WIMMessageDone:
		return "DONE"
	default:
		return fmt.Sprintf("%s: wParam=%d lParam=%d", evt.Name, evt.WParam, evt.LParam)
	}
}

func isNoisy(evt DecodedProgressEvent) bool {
	switch evt.MessageID {
	case WIMMessageProcess, WIMMessageFileInfo, WIMMessageChkProc:
		return true
	default:
		return len(evt.Name) >= 7 && evt.Name[:7] == "UNKNOWN"
	}
}
