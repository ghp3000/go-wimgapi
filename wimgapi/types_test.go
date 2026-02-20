//go:build windows

package wimgapi

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestNormalizeOpenOptionsDefaults(t *testing.T) {
	got := normalizeOpenOptions(OpenOptions{})
	if got.DesiredAccess != windows.GENERIC_READ {
		t.Fatalf("DesiredAccess default = %d, want %d", got.DesiredAccess, windows.GENERIC_READ)
	}
	if got.CreationDisposition != WIMOpenExisting {
		t.Fatalf("CreationDisposition default = %d, want %d", got.CreationDisposition, WIMOpenExisting)
	}
}

func TestNormalizeOpenOptionsKeepValues(t *testing.T) {
	in := OpenOptions{
		DesiredAccess:       0x1111,
		CreationDisposition: 0x2222,
		FlagsAndAttributes:  0x3333,
		CompressionType:     0x4444,
	}
	got := normalizeOpenOptions(in)
	if got != in {
		t.Fatalf("normalizeOpenOptions changed explicit fields: got=%+v want=%+v", got, in)
	}
}
