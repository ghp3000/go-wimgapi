//go:build windows

package go_wimgapi_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestRoundtripProgram(t *testing.T) {
	if os.Getenv("WIMGAPI_INTEGRATION") != "1" {
		t.Skip("set WIMGAPI_INTEGRATION=1 to run roundtrip integration test")
	}

	cmd := exec.Command("go", "run", "./examples/roundtrip-test")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	text := string(out)
	if err != nil {
		if strings.Contains(text, "code=1314") {
			t.Skip("insufficient privilege for WIM capture/apply; run elevated")
		}
		t.Fatalf("roundtrip failed: %v\n%s", err, text)
	}
	if !strings.Contains(text, "roundtrip test passed") {
		t.Fatalf("unexpected output:\n%s", text)
	}
}
