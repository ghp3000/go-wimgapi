//go:build windows

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ghp3000/go-wimgapi/wimgapi"
	"golang.org/x/sys/windows"
)

type treeEntry struct {
	IsDir  bool
	Size   int64
	SHA256 string
}

func main() {
	source := filepath.Clean(filepath.Join("examples", "testdata"))
	if _, err := os.Stat(source); err != nil {
		fail("source testdata not found: %v", err)
	}

	absSource, err := filepath.Abs(source)
	if err != nil {
		fail("resolve source path failed: %v", err)
	}

	tmpRoot, err := os.MkdirTemp("", "wimgapi-roundtrip-*")
	if err != nil {
		fail("create temp root failed: %v", err)
	}
	defer os.RemoveAll(tmpRoot)

	wimPath := filepath.Join(tmpRoot, "roundtrip.wim")
	restoreDir := filepath.Join(tmpRoot, "restore")
	wimTempDir := filepath.Join(tmpRoot, "wim-temp")
	if err := os.MkdirAll(restoreDir, 0o755); err != nil {
		fail("create restore dir failed: %v", err)
	}
	if err := os.MkdirAll(wimTempDir, 0o755); err != nil {
		fail("create wim temp dir failed: %v", err)
	}

	if err := captureDirectory(absSource, wimPath, wimTempDir); err != nil {
		if privilegeErr(err) {
			fail("capture failed: %v (run in elevated Administrator shell)", err)
		}
		fail("capture failed: %v", err)
	}
	fmt.Printf("captured: %s -> %s\n", absSource, wimPath)

	if err := verifyAndApply(wimPath, restoreDir, wimTempDir); err != nil {
		if privilegeErr(err) {
			fail("verify/apply failed: %v (run in elevated Administrator shell)", err)
		}
		fail("verify/apply failed: %v", err)
	}
	fmt.Printf("applied: %s -> %s\n", wimPath, restoreDir)

	if err := compareTrees(absSource, restoreDir); err != nil {
		fail("tree compare failed: %v", err)
	}

	fmt.Println("roundtrip test passed")
}

func captureDirectory(sourceDir, wimPath, tempDir string) error {
	f, err := wimgapi.Open(wimPath, wimgapi.OpenOptions{
		DesiredAccess:       windows.GENERIC_READ | windows.GENERIC_WRITE,
		CreationDisposition: wimgapi.WIMCreateAlways,
	})
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.SetTemporaryPath(tempDir); err != nil {
		return err
	}

	decoder := wimgapi.NewProgressDecoder()
	events := 0

	img, err := f.Capture(sourceDir, wimgapi.CaptureOptions{
		Progress: func(evt wimgapi.ProgressEvent) bool {
			events++
			decoded := decoder.Decode(evt)
			switch decoded.MessageID {
			case wimgapi.WIMMessageSetRange, wimgapi.WIMMessageSetPos, wimgapi.WIMMessageStepIt:
				fmt.Printf("[capture] progress: %d/%d (%.1f%%)\n", decoded.Current, decoded.Total, decoded.Percent)
			case wimgapi.WIMMessageError, wimgapi.WIMMessageWarning:
				fmt.Printf("[capture] %s\n", decoded.Summary)
			}
			return false
		},
	})
	if err != nil {
		return err
	}
	defer img.Close()
	fmt.Printf("[capture] callbacks=%d\n", events)
	return nil
}

func verifyAndApply(wimPath, restoreDir, tempDir string) error {
	f, err := wimgapi.Open(wimPath, wimgapi.OpenOptions{})
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.SetTemporaryPath(tempDir); err != nil {
		return err
	}

	count, err := f.ImageCount()
	if err != nil {
		return err
	}
	if count < 1 {
		return fmt.Errorf("unexpected image count: %d", count)
	}

	images, err := f.Images()
	if err != nil {
		return err
	}
	if len(images) < 1 {
		return fmt.Errorf("images() returned empty")
	}

	img, err := f.LoadImage(1)
	if err != nil {
		return err
	}
	defer img.Close()

	info, err := img.Info()
	if err != nil {
		return err
	}
	fmt.Printf("image #1: index=%d name=%q flags=%q arch=%q\n", info.Index, info.Name, info.Flags, info.Architecture)

	var events int
	decoder := wimgapi.NewProgressDecoder()
	var processedItems int
	var progressEvents int
	var nonProgressEvents int
	doneSeen := false
	err = img.Apply(restoreDir, wimgapi.ApplyOptions{
		Progress: func(evt wimgapi.ProgressEvent) bool {
			events++
			decoded := decoder.Decode(evt)

			switch evt.MessageID {
			case wimgapi.WIMMessageProcess, wimgapi.WIMMessageFileInfo, wimgapi.WIMMessageChkProc:
				processedItems++
			}
			switch decoded.MessageID {
			case wimgapi.WIMMessageSetRange, wimgapi.WIMMessageSetPos, wimgapi.WIMMessageStepIt:
				fmt.Printf("[apply] progress: %d/%d (%.1f%%)\n", decoded.Current, decoded.Total, decoded.Percent)
			case wimgapi.WIMMessageDone:
				doneSeen = true
				if decoded.Total > 0 {
					fmt.Printf("[apply] DONE: %d/%d (100.0%%), processed items=%d, callbacks=%d\n", decoded.Total, decoded.Total, processedItems, events)
				} else {
					fmt.Printf("[apply] DONE: processed items=%d, callbacks=%d\n", processedItems, events)
				}
			case wimgapi.WIMMessageError, wimgapi.WIMMessageWarning:
				fmt.Printf("[apply] %s\n", decoded.Summary)
			default:
				nonProgressEvents++
			}
			return false
		},
	})
	if err != nil {
		return err
	}
	if events == 0 {
		return fmt.Errorf("apply produced no callback events")
	}
	if !doneSeen {
		fmt.Printf("[apply] completed: processed items=%d, callbacks=%d, progress-events=%d, non-progress-events=%d\n", processedItems, events, progressEvents, nonProgressEvents)
	}
	return nil
}

func compareTrees(sourceDir, targetDir string) error {
	src, err := collectTree(sourceDir)
	if err != nil {
		return err
	}
	dst, err := collectTree(targetDir)
	if err != nil {
		return err
	}

	if len(src) != len(dst) {
		return fmt.Errorf("entry count mismatch: src=%d dst=%d", len(src), len(dst))
	}

	for p, a := range src {
		b, ok := dst[p]
		if !ok {
			return fmt.Errorf("missing entry in target: %s", p)
		}
		if a != b {
			return fmt.Errorf("entry mismatch for %s: src=%+v dst=%+v", p, a, b)
		}
	}
	return nil
}

func collectTree(root string) (map[string]treeEntry, error) {
	m := make(map[string]treeEntry)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		key := filepath.ToSlash(rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		if d.IsDir() {
			m[key] = treeEntry{IsDir: true}
			return nil
		}
		sum, err := fileSHA256(path)
		if err != nil {
			return err
		}
		m[key] = treeEntry{IsDir: false, Size: info.Size(), SHA256: sum}
		return nil
	})
	return m, err
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func privilegeErr(err error) bool {
	var werr *wimgapi.Error
	if errors.As(err, &werr) && werr.Code == 1314 {
		return true
	}
	return false
}
