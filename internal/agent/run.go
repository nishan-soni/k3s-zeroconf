package agent

func Run(deviceName string, mDNSPort int, k3sflags []string) {

	pairingInfoCh, pairingPort := startPairingServer()
	stopmDNSServer := registermDNS(deviceName, mDNSPort, pairingPort)

	<-pairingInfoCh

	stopmDNSServer()

	handoffToK3s(k3sflags)



}
