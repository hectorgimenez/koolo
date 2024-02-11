package memory

import (
	"fmt"
	"github.com/hectorgimenez/koolo/internal/helper/winproc"
	"github.com/winlabs/gowin32"
	"golang.org/x/sys/windows"
	"os"
	"strings"
	"syscall"
)

func InjectDLLs() error {
	dir, _ := os.Getwd()
	dllPath := dir + "\\" + "KooloDll.dll"

	lpAddress, _, err := winproc.VirtualAllocEx.Call(
		uintptr(handle),
		0,
		uintptr(len(dllPath))+1,
		uintptr(windows.MEM_COMMIT|windows.MEM_RESERVE),
		uintptr(windows.PAGE_READWRITE),
	)
	if err != nil && err != syscall.Errno(0) {
		return err
	}

	dst := uintptr(0)
	data := []byte(dllPath + "\x00")
	err = windows.WriteProcessMemory(handle, lpAddress, &data[0], uintptr(len(dllPath))+1, &dst)
	if err != nil {
		return err
	}

	threadHandle, _, err := winproc.CreateRemoteThread.Call(
		uintptr(handle),
		0,
		0,
		winproc.LoadLibraryA.Addr(),
		lpAddress,
		0,
		0,
	)
	if threadHandle == 0 {
		return err
	}

	return nil
}

func UnloadDll(pid uint32) error {
	modules, err := gowin32.GetProcessModules(pid)
	if err != nil {
		return fmt.Errorf("error getting process modules: %w", err)
	}

	for _, module := range modules {
		if strings.EqualFold(module.ModuleName, "KooloDll.dll") {
			threadHandle, _, err := winproc.CreateRemoteThread.Call(
				uintptr(handle),
				0,
				0,
				winproc.FreeLibrary.Addr(),
				uintptr(module.ModuleHandle),
				0,
				0,
			)
			if threadHandle == 0 {
				return err
			}
		}
	}

	return nil
}
