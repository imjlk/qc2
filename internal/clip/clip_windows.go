//go:build windows

package clip

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	openClipboardProc    = user32.NewProc("OpenClipboard")
	closeClipboardProc   = user32.NewProc("CloseClipboard")
	emptyClipboardProc   = user32.NewProc("EmptyClipboard")
	setClipboardDataProc = user32.NewProc("SetClipboardData")
	getClipboardDataProc = user32.NewProc("GetClipboardData")

	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	globalAllocProc  = kernel32.NewProc("GlobalAlloc")
	globalFreeProc   = kernel32.NewProc("GlobalFree")
	globalLockProc   = kernel32.NewProc("GlobalLock")
	globalUnlockProc = kernel32.NewProc("GlobalUnlock")
)

func (c SystemClipboard) copy(ctx context.Context, text string) error {
	return copyWindows(ctx, text)
}

// copyWindows writes UTF-16 text. clip.exe would decode stdin with the console
// code page and corrupt non-ASCII paths.
func copyWindows(ctx context.Context, text string) error {
	data, err := encodeClipboardUTF16(text)
	if err != nil {
		return fmt.Errorf("copy to clipboard: %w", err)
	}

	// OpenClipboard and CloseClipboard must run on the same OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := openClipboard(ctx); err != nil {
		return fmt.Errorf("copy to clipboard: %w", err)
	}
	defer closeClipboardProc.Call()

	if err := callProc(emptyClipboardProc); err != nil {
		return fmt.Errorf("copy to clipboard: %w", err)
	}

	size := uintptr(len(data)) * unsafe.Sizeof(data[0])
	handle, _, allocErr := globalAllocProc.Call(gmemMoveable, size)
	if handle == 0 {
		return fmt.Errorf("copy to clipboard: %w", allocErr)
	}
	owned := true
	defer func() {
		if owned {
			globalFreeProc.Call(handle)
		}
	}()

	ptr, _, lockErr := globalLockProc.Call(handle)
	if ptr == 0 {
		return fmt.Errorf("copy to clipboard: %w", lockErr)
	}
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(data))
	copy(dst, data)
	if err := unlockClipboardMemory(handle); err != nil {
		return fmt.Errorf("copy to clipboard: %w", err)
	}

	if err := callProc(setClipboardDataProc, cfUnicodeText, handle); err != nil {
		return fmt.Errorf("copy to clipboard: %w", err)
	}
	// The system owns the memory after a successful SetClipboardData.
	owned = false
	return nil
}

func openClipboard(ctx context.Context) error {
	var last error
	for attempt := 0; attempt < 10; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		r, _, err := openClipboardProc.Call(0)
		if r != 0 {
			return nil
		}
		last = err
		if attempt == 9 {
			break
		}
		timer := time.NewTimer(50 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	if last == nil {
		last = syscall.EINVAL
	}
	return last
}

func callProc(proc *syscall.LazyProc, args ...uintptr) error {
	r, _, err := proc.Call(args...)
	if r == 0 {
		if err == nil {
			return syscall.EINVAL
		}
		return err
	}
	return nil
}

func unlockClipboardMemory(handle uintptr) error {
	r, _, err := globalUnlockProc.Call(handle)
	if r != 0 {
		return nil
	}
	// GlobalUnlock returns zero when the lock count reaches zero and reports
	// Errno(0). That is success.
	if errno, ok := err.(syscall.Errno); ok && errno == 0 {
		return nil
	}
	if err == nil {
		return nil
	}
	return err
}
