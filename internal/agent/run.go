package agent

import (
	"log/slog"

	"github.com/nishan-soni/k3s-zeroconf/internal/common"
)

type Agent struct {
	StartPairingServer func() (<-chan common.PairingInfo, int, error)
	RegisterMDNS       func(name string, port int, pairingPort int, address string) (func(), error)
	HandoffToK3s       func(flags []string) error
}

func NewAgent() *Agent {
	return &Agent{
		StartPairingServer: startPairingServer,
		RegisterMDNS:       registermDNS,
		HandoffToK3s:       handoffToK3s,
	}
}

// Starts the pairing process for the agent by registering the device to mDNS
// and starting an http server for masters to add it to the cluster.
func (a *Agent) Run(deviceName string, mDNSServerPort int, extraK3sflags []string, agentAddress string) error {
	pairingInfoCh, pairingPort, err := a.StartPairingServer()
	if err != nil {
		return err
	}

	stopmDNSServer, err := a.RegisterMDNS(deviceName, mDNSServerPort, pairingPort, agentAddress)
	if err != nil {
		return err
	}
	defer stopmDNSServer()

	slog.Info("Waiting for requests.")

	pairingInfo := <-pairingInfoCh
	startupFlags := []string{"--server", pairingInfo.MasterAddress, "--token", pairingInfo.JoinToken}
	flags := append(startupFlags, extraK3sflags...)

	if err = a.HandoffToK3s(flags); err != nil {
		return err
	}
	return nil
}
