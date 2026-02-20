//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"go-wimgapi/wimgapi"
	"golang.org/x/sys/windows"
)

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	switch cmd {
	case "list":
		if len(os.Args) != 3 {
			usage()
			os.Exit(2)
		}
		if err := runList(os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "apply":
		if len(os.Args) != 5 {
			usage()
			os.Exit(2)
		}
		idx, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: index must be an integer")
			os.Exit(2)
		}
		if err := runApply(os.Args[2], idx, os.Args[4]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "capture":
		if len(os.Args) != 4 {
			usage()
			os.Exit(2)
		}
		if err := runCapture(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func runList(wimPath string) error {
	f, err := wimgapi.Open(wimPath, wimgapi.OpenOptions{})
	if err != nil {
		return err
	}
	defer f.Close()

	images, err := f.Images()
	if err != nil {
		return err
	}
	for _, img := range images {
		fmt.Printf("#%d\t%s\t%s\t%s\t%s\n", img.Index, img.Name, img.Description, img.Flags, img.Architecture)
	}
	return nil
}

func runApply(wimPath string, index int, target string) error {
	f, err := wimgapi.Open(wimPath, wimgapi.OpenOptions{})
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := f.LoadImage(index)
	if err != nil {
		return err
	}
	defer img.Close()

	applied := false
	err = img.Apply(target, wimgapi.ApplyOptions{
		Progress: func(evt wimgapi.ProgressEvent) bool {
			if evt.MessageID == wimgapi.WIMMessageDone {
				applied = true
			}
			return false
		},
	})
	if err != nil {
		return err
	}
	if !applied {
		return errors.New("apply finished without done callback")
	}
	return nil
}

func runCapture(sourceDir, wimPath string) error {
	f, err := wimgapi.Open(wimPath, wimgapi.OpenOptions{
		DesiredAccess:       windows.GENERIC_READ | windows.GENERIC_WRITE,
		CreationDisposition: wimgapi.WIMCreateAlways,
	})
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := f.Capture(sourceDir, wimgapi.CaptureOptions{})
	if err != nil {
		return err
	}
	defer img.Close()
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  wimctl list <path-to-wim>")
	fmt.Fprintln(os.Stderr, "  wimctl apply <path-to-wim> <index> <target-dir>")
	fmt.Fprintln(os.Stderr, "  wimctl capture <source-dir> <path-to-wim>")
}
