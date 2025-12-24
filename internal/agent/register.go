package agent

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/grandcat/zeroconf"
	"github.com/nishan-soni/k3_zeroconf/internal/common"
)

func AnnounceAgent(name string, port int) {
	// Announces the node to mdns
	server, err := zeroconf.Register(name, common.ServiceType, common.LocalDomain, port, []string{}, nil)
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	fmt.Println("Shutting down.")
}
