package agent

// Starts the pairing process for the agent by registering the device to mDNS
// and starting an http server for masters to add it to the cluster.
func Run(deviceName string, mDNSPort int, extraK3sflags []string, agentAddress string) error {
	pairingInfoCh, pairingPort, err := startPairingServer()
	if err != nil {
		return err
	}

	stopmDNSServer, err := registermDNS(deviceName, mDNSPort, pairingPort, agentAddress)
	if err != nil {
		return err
	}
	defer stopmDNSServer()

	pairingInfo := <-pairingInfoCh

	startupFlags := []string{"--server", pairingInfo.MasterAddress, "--token", pairingInfo.JoinToken}
	flags := append(startupFlags, extraK3sflags...)
	err = handoffToK3s(flags)
	if err != nil {
		return err
	}
	return nil
}
