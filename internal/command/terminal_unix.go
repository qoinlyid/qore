//go:build darwin || linux || freebsd || openbsd || netbsd

package command

// Unix-like already support ANSI.
func enableVirtualTerminalProcessing() {}
