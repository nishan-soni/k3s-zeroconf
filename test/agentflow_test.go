package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/nishan-soni/k3s-zeroconf/internal/agent"
	"github.com/nishan-soni/k3s-zeroconf/internal/common"
	"github.com/nishan-soni/k3s-zeroconf/internal/master"
)

var agentAddress, _ = common.GetOutboundIP()

const (
	deviceName    = "test_device"
	masterAddress = "123.123.123.123"
)

func TestAgentFlow(t *testing.T) {
	errCh := make(chan error, 1)

	orch := agent.NewAgent()
	var capturedFlags []string
	orch.HandoffToK3s = func(flags []string) error {
		capturedFlags = flags
		return nil
	}

	// Start the agent server
	go func() {
		errCh <- orch.Run(deviceName, common.MDNSServerPort, []string{"--test-flag=value"}, agentAddress)
	}()

	time.Sleep(time.Second) // Give some time for the agent to start

	// mDNS assertions
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

	// Make a request to add the node to the cluster
	pairingRequest := common.PairingInfo{
		MasterAddress: masterAddress,
		JoinToken:     "token",
	}
	payload, _ := json.Marshal(pairingRequest)
	endpoint := fmt.Sprintf("http://%s:%d/%s", agentAddress, common.PairingServerPort, common.JoinEndpoint)
	_, err = http.Post(endpoint, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Error making join request: %v", err)
	}

	// Wait for agent to finish
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for agent to finish")
	}

	// Assert correct k3s flags
	expectedFlags := []string{"--server", masterAddress, "--token", "token", "--test-flag=value"}
	if !reflect.DeepEqual(capturedFlags, expectedFlags) {
		t.Errorf("Expected flags %v, got %v", expectedFlags, capturedFlags)
	}
}
