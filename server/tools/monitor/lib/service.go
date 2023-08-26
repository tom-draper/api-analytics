package monitor

import "os/exec"

func ServiceDown(service string) bool {
	cmd := exec.Command("systemctl", "check", service)
	_, err := cmd.CombinedOutput()
	return err != nil
}
