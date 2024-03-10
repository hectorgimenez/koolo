package game

import (
	"fmt"
	"os/exec"
	"strings"
)

func KillMultiClientHandleForPID(pid uint32) error {
	stdout, err := exec.Command("./tools/handle64.exe", "-accepteula", "-nobanner", "-a", "-v", "-p", fmt.Sprintf("%d", pid), "Check For Other Instances").Output()
	if err != nil {
		return fmt.Errorf("error running handle64.exe: %d", stdout)
	}

	stdoutLines := strings.Split(string(stdout), "\r\n")

	// Maybe the handle was previously closed, so we don't need to do anything
	if len(stdoutLines) > 1 {
		for _, line := range stdoutLines {
			if strings.Contains(line, "Check For Other Instances") {
				cols := strings.Split(line, ",") // 0: process, 1: pid, 2: type, 3: handle, 4: name

				stdout, err = exec.Command("./tools/handle64.exe", "-accepteula", "-nobanner", "-p", fmt.Sprintf("%d", pid), "-c", cols[3], "-y").Output()
				if err != nil {
					return fmt.Errorf("error running handle64.exe: %d", stdout)
				}
			}
		}
	}

	return nil
}
