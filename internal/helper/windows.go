package helper

import (
	"golang.org/x/sys/windows"
	"os"
	"syscall"
)

func HasAdminPermission() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")

	return err == nil
}

func ShowDialog(title, message string) {
	t, _ := syscall.UTF16PtrFromString(title)
	txt, _ := syscall.UTF16PtrFromString(message)

	windows.MessageBox(0, txt, t, 0)
}
