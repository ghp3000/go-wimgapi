//go:build windows

package wimgapi

import (
	"encoding/xml"
	"strings"
	"unsafe"
)

type imageInfoXML struct {
	Index       int    `xml:"INDEX,attr"`
	Name        string `xml:"NAME"`
	Description string `xml:"DESCRIPTION"`
	Flags       string `xml:"FLAGS"`
	Windows     struct {
		Arch string `xml:"ARCH"`
	} `xml:"WINDOWS"`
}

type imageListXML struct {
	Images []imageInfoXML `xml:"IMAGE"`
}

func (i *Image) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.closed {
		return nil
	}

	r1, _, callErr := procWIMCloseHandle.Call(uintptr(i.handle))
	if r1 == 0 {
		code := codeFromCallErr(callErr)
		if code == 0 {
			code = lastErrorCode()
		}
		return winError("WIMCloseHandle", code)
	}
	i.closed = true
	return nil
}

func (i *Image) Info() (ImageInfo, error) {
	var p uintptr
	var size uint32
	r1, _, callErr := procWIMGetImageInformation.Call(
		uintptr(i.handle),
		uintptr(unsafe.Pointer(&p)),
		uintptr(unsafe.Pointer(&size)),
	)
	if r1 == 0 {
		code := codeFromCallErr(callErr)
		if code == 0 {
			code = lastErrorCode()
		}
		return ImageInfo{}, winError("WIMGetImageInformation", code)
	}
	defer tryFreeMemory(p)

	raw := BytesFromPointer(p, size)
	xmlText := strings.TrimSpace(DecodeUTF16Bytes(raw))
	if xmlText == "" {
		return ImageInfo{}, nil
	}

	var imgNode imageInfoXML
	if err := xml.Unmarshal([]byte(xmlText), &imgNode); err != nil {
		return ImageInfo{}, err
	}
	if imgNode.Index == 0 && imgNode.Name == "" {
		var list imageListXML
		if err := xml.Unmarshal([]byte(xmlText), &list); err != nil {
			return ImageInfo{}, err
		}
		if len(list.Images) > 0 {
			imgNode = list.Images[0]
		}
	}

	return ImageInfo{
		Index:        imgNode.Index,
		Name:         imgNode.Name,
		Description:  imgNode.Description,
		Flags:        imgNode.Flags,
		Architecture: imgNode.Windows.Arch,
	}, nil
}
