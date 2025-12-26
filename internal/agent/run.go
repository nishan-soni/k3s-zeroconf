package agent

func Run(hostName string, mDNSPort int) {

	pairingInfoCh, pairingPort := startPairingServer()
	stopmDNSServer := registermDNS(hostName, mDNSPort, pairingPort)

	<-pairingInfoCh

	stopmDNSServer()



}
