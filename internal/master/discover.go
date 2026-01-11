package master

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/nishan-soni/k3s-zeroconf/internal/common"
)

type discoveredNode struct {
	address     string
	id          string
	pairingPort string
}

// List all nodes announced in mDNS.
func ListNodes(timeout time.Duration) error {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	services := make(chan *zeroconf.ServiceEntry)
	err = resolver.Browse(ctx, common.ServiceType, common.LocalDomain, services)
	if err != nil {
		return err
	}

	fmt.Printf("Discovering Nodes (%s)\n", timeout)

	var results []discoveredNode
	for entry := range services {
		discoveredNode, err := parseServiceEntry(entry)
		if err != nil {
			slog.Error("Failed to parse service.", "service", entry.Instance, slog.Any("error", err))
		}
		results = append(results, discoveredNode)
	}

	printNodesTable(results)
	return nil
}

// Parses an mDNS service entry into a discoveredNode.
func parseServiceEntry(serviceEntry *zeroconf.ServiceEntry) (node discoveredNode, err error) {
	if len(serviceEntry.AddrIPv4) == 0 {
		err = fmt.Errorf("node does not have an address.")
	}

	node.address = serviceEntry.AddrIPv4[0].String()
	node.id = serviceEntry.Instance

	if len(serviceEntry.Text) > 0 {
		_, after, found := strings.Cut(serviceEntry.Text[0], "=")
		if found {
			node.pairingPort = after
		} else {
			err = fmt.Errorf("node does not have a port.")
		}
	} else {
		err = fmt.Errorf("node does not have any text.")
	}

	return
}

func printNodesTable(discoveredNodes []discoveredNode) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(writer, "NAME\tCONNECT IP\tCONNECT PORT")

	for _, node := range discoveredNodes {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", node.id, node.address, node.pairingPort)
	}
	writer.Flush()
}

// Check if a node exists in mDNS.
func lookupNode(nodeName string) (discoveredNode, error) {
	resolver, _ := zeroconf.NewResolver(nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry)
	err := resolver.Lookup(ctx, nodeName, common.ServiceType, common.LocalDomain, entries)
	if err != nil {
		return discoveredNode{}, err
	}

	entry := <-entries
	if entry == nil {
		return discoveredNode{}, fmt.Errorf("node not found in mDNS.")
	}

	discovered, err := parseServiceEntry(entry)
	if err != nil {
		return discoveredNode{}, err
	}
	return discovered, nil
}

// Gets the master's server token from the path.
func getServerToken(tokenPath string) (string, error) {
	tokenFile, err := os.Open(tokenPath)
	if err != nil {
		return "", err
	}
	defer tokenFile.Close()

	scanner := bufio.NewScanner(tokenFile)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", fmt.Errorf("server token is empty.")
}

// Adds the node to the k3s cluster by sending an http request to the agent.
func AddNode(nodeName string, serverAddress string, serverTokenPath string) error {
	node, err := lookupNode(nodeName)
	if err != nil {
		return err
	}

	serverToken, err := getServerToken(serverTokenPath)
	if err != nil {
		return err
	}
	info := common.PairingInfo{
		MasterAddress: fmt.Sprintf("https://%s:%d", serverAddress, common.K3sServerPort),
		JoinToken:     serverToken,
	}

	payload, err := json.Marshal(info)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("http://%s:%s/%s", node.address, node.pairingPort, common.JoinEndpoint)
	_, err = http.Post(endpoint, "application/json", bytes.NewBuffer(payload))

	if err != nil {
		return err
	}

	fmt.Printf("Added %s to the cluster!\n", nodeName)

	return nil
}
