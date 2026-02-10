// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

//go:build !windows

package copilot

import (
	"fmt"
	"os"
	"runtime"
)

type consoleHandles struct {
	conin, conout *os.File
}

func attachConsole() (*consoleHandles, error) {
	return nil, fmt.Errorf("attachConsole is only supported on Windows (current OS: %s)", runtime.GOOS)
}

func (h *consoleHandles) restore() {}
