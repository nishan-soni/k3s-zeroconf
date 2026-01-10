package agent

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"log/slog"

	"github.com/grandcat/zeroconf"
	"github.com/nishan-soni/k3_zeroconf/internal/common"
)

func registermDNS(name string, port int, pairingPort int, address string) (func(), error) {
	var (
		server *zeroconf.Server
		err    error
		advertisedAddress string
	)

	txt := []string{"pairing_port=" + strconv.Itoa(pairingPort)}

	if address == "" {
		server, err = zeroconf.Register(name, common.ServiceType, common.LocalDomain, port, txt, nil)
		advertisedAddress, _ = common.GetOutboundIP()
	} else {
		ipAddr := net.ParseIP(address)
		if ipAddr == nil {
			return nil, fmt.Errorf("invalid IP address: %s", address)
		}

		advertisedAddress = address

		hostName, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		server, err = zeroconf.RegisterProxy(
			name,
			common.ServiceType,
			common.LocalDomain,
			port,
			hostName,
			[]string{address},
			txt,
			nil,
		)
	}
	if err != nil {
		return nil, err
	}

	slog.Info("Registered device to mDNS.", "service-name", name, "service-type", common.ServiceType, "advertised-address", advertisedAddress)

	return server.Shutdown, nil
}
