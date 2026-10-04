//go:build windows

package clip

import (
	"context"
	"runtime"
	"testing"
	"unicode/utf16"
	"unsafe"
)

func TestCopyWindowsRoundTrip(t *testing.T) {
	text := "qc2-한글 path"
	if err := copyWindows(context.Background(), text); err != nil {
		t.Fatalf("copyWindows() error = %v", err)
	}

	got, err := readWindowsClipboard()
	if err != nil {
		t.Fatalf("readWindowsClipboard() error = %v", err)
	}
	if got != text {
		t.Fatalf("clipboard = %q, want %q", got, text)
	}
}

func readWindowsClipboard() (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := openClipboard(context.Background()); err != nil {
		return "", err
	}
	defer closeClipboardProc.Call()

	handle, _, err := getClipboardDataProc.Call(cfUnicodeText)
	if handle == 0 {
		return "", err
	}
	ptr, _, err := globalLockProc.Call(handle)
	if ptr == 0 {
		return "", err
	}
	defer globalUnlockProc.Call(handle)

	const maxUnits = 1 << 20
	raw := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), maxUnits)
	n := 0
	for n < len(raw) && raw[n] != 0 {
		n++
	}
	return string(utf16.Decode(raw[:n])), nil
}
