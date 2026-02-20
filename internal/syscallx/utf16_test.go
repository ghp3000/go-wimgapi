package syscallx

import "testing"

func TestDecodeUTF16Bytes(t *testing.T) {
	// "AB" in UTF-16LE + NUL
	data := []byte{0x41, 0x00, 0x42, 0x00, 0x00, 0x00}
	if got := DecodeUTF16Bytes(data); got != "AB" {
		t.Fatalf("DecodeUTF16Bytes() = %q, want %q", got, "AB")
	}
}

func TestDecodeUTF16BytesWithBOM(t *testing.T) {
	// BOM + "X"
	data := []byte{0xFF, 0xFE, 0x58, 0x00, 0x00, 0x00}
	if got := DecodeUTF16Bytes(data); got != "X" {
		t.Fatalf("DecodeUTF16Bytes() = %q, want %q", got, "X")
	}
}
