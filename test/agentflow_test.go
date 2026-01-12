package test

import (
	"strconv"
	"testing"
	"time"

	"github.com/nishan-soni/k3s-zeroconf/internal/agent"
	"github.com/nishan-soni/k3s-zeroconf/internal/common"
	"github.com/nishan-soni/k3s-zeroconf/internal/master"
)

const deviceName string = "test_device"
const agentAddress string = "192.168.1.10"

func TestMDNSRegistration(t *testing.T) {
	go func() {
		agent.Run(deviceName, common.MDNSServerPort, []string{"--test-flag=value"}, agentAddress)
	}()

	time.Sleep(time.Second) // Give some time for the agent to start

	node, err := master.LookupNode(deviceName)

	if err != nil {
		t.Fatalf("Unable to look up the agent. %v", err)
	}

	if address := node.Address(); address != agentAddress {
		t.Errorf("Address mismatch: expected %s, got %s", agentAddress, address)
	}

	if name := node.ID(); name != deviceName {
		t.Errorf("Device name mismatch: expected %s, got %s", deviceName, name)
	}

	expectedPort := strconv.Itoa(common.PairingServerPort)

	if port := node.PairingPort(); port != expectedPort {
		t.Errorf("Port mismatch: expected %s, got %s", port, expectedPort)
	}

}
