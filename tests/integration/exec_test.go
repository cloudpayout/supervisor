package integration

import "os/exec"

func execCommand(bin string, env []string, args ...string) *exec.Cmd {
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	return cmd
}
