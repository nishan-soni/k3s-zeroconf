package agent

// Starts the pairing process for the agent by registering the device to mDNS
// and starting an http server for masters to add it to the cluster.
func Run(deviceName string, mDNSPort int, extraK3sflags []string) error {
	pairingInfoCh, pairingPort, err := startPairingServer()
	if err != nil {
		return err
	}

	stopmDNSServer, err := registermDNS(deviceName, mDNSPort, pairingPort)
	defer stopmDNSServer()
	if err != nil {
		return err
	}

	pairingInfo := <-pairingInfoCh

	startupFlags := []string{"--server", pairingInfo.MasterAddress, "--token", pairingInfo.JoinToken, "--lb-server-port", "55332"}
	flags := append(startupFlags, extraK3sflags...)
	err = handoffToK3s(flags)
	if err != nil {
		return err
	}
	return nil
}
