package game

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

func KillAllClientHandles() error {
	cmd := exec.Command("./tools/handle64.exe", "-accepteula", "-nobanner", "-a", "-v", "-p", "d2r.exe", "Check For Other Instances")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	stdout, err := cmd.Output()
	if err != nil && !strings.Contains(string(stdout), "No matching handles found.") {
		return fmt.Errorf("error running handle64.exe: %d", stdout)
	}

	stdoutLines := strings.Split(string(stdout), "\r\n")

	// Maybe the handle was previously closed, so we don't need to do anything
	if len(stdoutLines) > 1 {
		for _, line := range stdoutLines {
			if strings.Contains(line, "Check For Other Instances") {
				cols := strings.Split(line, ",") // 0: process, 1: pid, 2: type, 3: handle, 4: name

				cmd = exec.Command("./tools/handle64.exe", "-accepteula", "-nobanner", "-p", fmt.Sprintf("%s", cols[1]), "-c", cols[3], "-y")
				cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				stdout, err = cmd.Output()
				if err != nil {
					return fmt.Errorf("error running handle64.exe: %d", stdout)
				}
			}
		}
	}

	return nil
}
