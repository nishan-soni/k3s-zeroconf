package agent

import (
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	"github.com/nishan-soni/k3_zeroconf/internal/common"
)

func detectFirewall() string {
	if _, err := exec.LookPath("firewall-cmd"); err == nil {
		return "firewalld"
	}
	if _, err := exec.LookPath("ufw"); err == nil {
		return "ufw"
	}
	return ""
}

// Opens the required ports for k3s to run properly
// Expects the program to be ran with elevated privileges
func openPorts(ips []string, ports []string) error {
	var (
		err    error
		output []byte
	)

	slog.Info("Opening the required ports for k3s to work. Additional ports may be required to be opened based on your setup. https://docs.k3s.io/installation/requirements?os=debian#inbound-rules-for-k3s-nodes", "opened-ports", ports, "trusted-ips", ips)

	firewall := detectFirewall()
	switch firewall {
	case "firewalld":
		for _, port := range ports {
			output, err = exec.Command("firewall-cmd", "--add-port="+port).CombinedOutput()
			if err != nil {
				return err
			}
		}

		for _, ip := range ips {
			output, err = exec.Command("firewall-cmd", "--zone=trusted", "--add-source="+ip).CombinedOutput()
			if err != nil {
				return err
			}
		}

		_, err = exec.Command("firewall-cmd", "--reload").CombinedOutput()
		if err != nil {
			return err
		}

	case "ufw":
		for _, port := range ports {
			output, err = exec.Command("ufw", "allow", port).CombinedOutput()
			if err != nil {
				return err
			}
		}

		for _, ip := range ips {
			output, err = exec.Command("ufw", "allow", "from", ip, "to", "any").CombinedOutput()
			if err != nil {
				return err
			}
		}
	default:
		return nil
	}

	slog.Info("opened ports", slog.Any("output", output))
	return nil
}

// Hands off the process to k3s which then attaches the agent to the cluster.
func handoffToK3s(flags []string) error {
	k3sPath, err := exec.LookPath("k3s")
	if err != nil {
		return err
	}

	openPorts(common.K3sAllowIps, common.K3sAllowPorts)

	args := append([]string{"k3s", "agent"}, flags...)
	env := os.Environ()

	slog.Info("Handing off to k3s.", "command", args, "env", env)
	err = syscall.Exec(k3sPath, args, env)
	if err != nil {
		return err
	}

	return nil

}
