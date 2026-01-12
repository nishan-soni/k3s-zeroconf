package agent

import "log/slog"

// Starts the pairing process for the agent by registering the device to mDNS
// and starting an http server for masters to add it to the cluster.
func Run(deviceName string, mDNSServerPort int, extraK3sflags []string, agentAddress string) error {
	pairingInfoCh, pairingPort, err := startPairingServer()
	if err != nil {
		return err
	}

	stopmDNSServer, err := registermDNS(deviceName, mDNSServerPort, pairingPort, agentAddress)
	if err != nil {
		return err
	}
	defer stopmDNSServer()
	slog.Info("Waiting for requests.")

	pairingInfo := <-pairingInfoCh

	startupFlags := []string{"--server", pairingInfo.MasterAddress, "--token", pairingInfo.JoinToken}
	flags := append(startupFlags, extraK3sflags...)

	if err = handoffToK3s(flags); err != nil {
		return err
	}

	return nil
}
