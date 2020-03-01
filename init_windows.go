// +build windows
package main

import (
  "os"
  "golang.org/x/sys/windows"
)

// set ENABLE_VIRTUAL_TERMINAL_PROCESSING on windows
// this makes cmd.exe interpret ANSI escape codes
func init() {
  stdout := windows.Handle(os.Stdout.Fd())
  var originalMode uint32

  windows.GetConsoleMode(stdout, &originalMode)
  windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}
