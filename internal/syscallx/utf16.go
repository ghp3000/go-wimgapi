package syscallx

import (
	"encoding/binary"
	"unicode/utf16"
)

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
