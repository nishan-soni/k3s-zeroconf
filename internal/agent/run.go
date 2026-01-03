package agent

// Starts the pairing process for the agent by registering the device to mDNS
// and starting an http server for masters to add it to the cluster.
func Run(deviceName string, mDNSPort int, k3sflags []string) error {
	pairingInfoCh, pairingPort, err := startPairingServer()
	if err != nil {
		return err
	}

	stopmDNSServer, err := registermDNS(deviceName, mDNSPort, pairingPort)
	defer stopmDNSServer()
	if err != nil {
		return err
	}

	<-pairingInfoCh

	err = handoffToK3s(k3sflags)
	if err != nil {
		return err
	}
	return nil
}
