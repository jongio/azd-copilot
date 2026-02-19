// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

//go:build windows

package copilot

import (
	"os"
	"syscall"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procSetStdHandle = kernel32.NewProc("SetStdHandle")
	procGetStdHandle = kernel32.NewProc("GetStdHandle")
)

const (
	stdInputHandle  = uintptr(0xFFFFFFF6) // STD_INPUT_HANDLE  = -10
	stdOutputHandle = uintptr(0xFFFFFFF5) // STD_OUTPUT_HANDLE = -11
	stdErrorHandle  = uintptr(0xFFFFFFF4) // STD_ERROR_HANDLE  = -12
)

// consoleHandles holds CONIN$/CONOUT$ file handles and the original
// process standard handles so they can be restored after the child exits.
type consoleHandles struct {
	conin, conout            *os.File
	origIn, origOut, origErr uintptr
}

// attachConsole opens CONIN$/CONOUT$ and uses SetStdHandle to make them
// the process's standard handles. Child processes spawned via exec.Cmd
// will then inherit real console handles that Node.js recognises as a TTY
// (isTTY = true, columns/rows populated).
//
// Simply passing the CONOUT$ *os.File as cmd.Stdout is NOT enough —
// Node.js only treats fd 1 as a TTY when the underlying Windows handle
// was the process's standard output at spawn time.
func attachConsole() (*consoleHandles, error) {
	conout, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	conin, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		_ = conout.Close()
		return nil, err
	}

	h := &consoleHandles{conin: conin, conout: conout}

	// Save originals
	h.origIn, _, _ = procGetStdHandle.Call(stdInputHandle)
	h.origOut, _, _ = procGetStdHandle.Call(stdOutputHandle)
	h.origErr, _, _ = procGetStdHandle.Call(stdErrorHandle)

	// Redirect to console
	_, _, _ = procSetStdHandle.Call(stdInputHandle, uintptr(conin.Fd()))
	_, _, _ = procSetStdHandle.Call(stdOutputHandle, uintptr(conout.Fd()))
	_, _, _ = procSetStdHandle.Call(stdErrorHandle, uintptr(conout.Fd()))

	return h, nil
}

// restore puts the original standard handles back and closes the console files.
func (h *consoleHandles) restore() {
	_, _, _ = procSetStdHandle.Call(stdInputHandle, h.origIn)
	_, _, _ = procSetStdHandle.Call(stdOutputHandle, h.origOut)
	_, _, _ = procSetStdHandle.Call(stdErrorHandle, h.origErr)
	_ = h.conin.Close()
	_ = h.conout.Close()
}
