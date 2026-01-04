package agent

import (
	"os"
	"os/exec"
	"syscall"
)

// Hands off the process to k3s which then attaches the agent to the cluster.
func handoffToK3s(flags []string) error {
	k3sPath, err := exec.LookPath("k3s")
	if err != nil {
		return err
	}

	args := append([]string{"k3s", "agent"}, flags...)
	env := os.Environ()

	err = syscall.Exec(k3sPath, args, env)
	if err != nil {
		return err
	}

	return nil

}
