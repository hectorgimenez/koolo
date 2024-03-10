package helper

import "os"

func HasAdminPermission() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")

	return err == nil
}
