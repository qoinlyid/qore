//go:build windows

package command

import (
	"syscall"
	"unsafe"
)

func enableVirtualTerminalProcessing() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	getStdHandle := kernel32.NewProc("GetStdHandle")

	const STD_OUTPUT_HANDLE = uint32(-11 & 0xFFFFFFFF)
	const ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004

	hOut, _, _ := getStdHandle.Call(uintptr(STD_OUTPUT_HANDLE))
	var mode uint32
	getConsoleMode.Call(hOut, uintptr(unsafe.Pointer(&mode)))
	mode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING
	setConsoleMode.Call(hOut, uintptr(mode))
}
