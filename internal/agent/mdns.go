package agent

import (
	"fmt"
	"log"
	"strconv"

	"github.com/grandcat/zeroconf"
	"github.com/nishan-soni/k3_zeroconf/internal/common"
)

func registermDNS(name string, port int, pairingPort int) func() {

	server, err := zeroconf.Register(name, common.ServiceType, common.LocalDomain, port, []string{"pairing_port=" + strconv.Itoa(pairingPort)}, nil)
	if err != nil {
		log.Fatalln("Failed to register agent to mDNS", err.Error())
	}

	fmt.Printf("%s is open for connection on port %d.\n", name, pairingPort)

	return server.Shutdown
}
