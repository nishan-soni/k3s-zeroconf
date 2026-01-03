package master

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/nishan-soni/k3_zeroconf/internal/common"
)

type DiscoveredNode struct {
	address     string
	id          string
	pairingPort string
}

var (
	ErrNodeNotFound = errors.New("node not found in mDNS.")
)

/*
Connection flow:

	node announces its self in mDNS

	master runs `mytool node list` to view all the nodes in mDNS
	master runs 'mytool node connect mynode-1' to connect to the node
		-> master sends an http request to mynode-1 with its ip and connection token
		-> mynode-1 receieves the http request and then runs the k3 agent command
		-> node is now connected to the cluster
*/

func ListNodes(timeout time.Duration) error {
	resolver, err := zeroconf.NewResolver(nil)

	if err != nil {
		return err
	}

	var results []*DiscoveredNode
	services := make(chan *zeroconf.ServiceEntry)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		err := resolver.Browse(ctx, common.ServiceType, common.LocalDomain, services)
		if err != nil {
			log.Fatalln("Failed to browse mDNS: %w", err.Error())
		}
	}()

	fmt.Printf("Discovering Nodes (%s)\n", timeout)

	for entry := range services {
		discoveredNode, err := parseServiceEntry(entry)
		if err != nil {
			log.Printf("Failed to parse service %s\n", entry.Instance)
		}
		results = append(results, &discoveredNode)
	}

	printNodesTable(results)

	return nil
}

func parseServiceEntry(serviceEntry *zeroconf.ServiceEntry) (DiscoveredNode, error) {

	if len(serviceEntry.AddrIPv4) == 0 {
		return DiscoveredNode{}, fmt.Errorf("node does not have an address.")
	}

	pairingPort := ""
	if len(serviceEntry.Text) > 0 {
		_, after, found := strings.Cut(serviceEntry.Text[0], "=")
		if found {
			pairingPort = after
		}
	} else {
		return DiscoveredNode{}, fmt.Errorf("node does not have a port.")
	}

	return DiscoveredNode{
		address:     serviceEntry.AddrIPv4[0].String(),
		id:          serviceEntry.Instance,
		pairingPort: pairingPort,
	}, nil
}

func printNodesTable(discoveredNodes []*DiscoveredNode) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(writer, "NAME\tIP\tSERVER PORT")

	for _, node := range discoveredNodes {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", node.id, node.address, node.pairingPort)
	}
	writer.Flush()
}

func lookupNode(nodeName string) (DiscoveredNode, error) {
	resolver, _ := zeroconf.NewResolver(nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	entries := make(chan *zeroconf.ServiceEntry)

	go func() {
		err := resolver.Lookup(ctx, nodeName, common.ServiceType, common.LocalDomain, entries)
		if err != nil {
			log.Fatalln("Failed to identify node:", err)
		}
	}()

	entry := <-entries
	if entry == nil {
		return DiscoveredNode{}, ErrNodeNotFound
	}

	discovered, err := parseServiceEntry(entry)
	if err != nil {
		return DiscoveredNode{}, fmt.Errorf("Failed to parse service entry: %w", err)
	}
	return discovered, nil

}

func getServerToken(tokenPath string) (string, error) {
	tokenFile, err := os.Open(tokenPath)
	if err != nil {
		return "", fmt.Errorf("Failed to get server token: %w", err)
	}

	defer tokenFile.Close()

	scanner := bufio.NewScanner(tokenFile)
	return scanner.Text(), nil
}

func AddNode(nodeName string, serverAddress string, serverTokenPath string) error {

	// Sends an http request to the node to the cluster

	// Maybe it should check if it was actually added? Idk if it should be done here though

	// Errors that it raises:
	//	Node no longer available (because its already connected to smth else or it just died)
	//		This would prob be raised on a 404 error
	// 	Other connection failures
	//		Add a context timeout thing to check after 30s if the node was added or not
	//			Can check using the k3 cli command most likley

	// Return these errors soon
	node, err := lookupNode(nodeName)

	if err != nil {
		return err
	}

	serverToken, _ := getServerToken(serverTokenPath)

	info := common.PairingInfo{
		MasterIP:  serverAddress,
		JoinToken: serverToken,
	}

	payload, err := json.Marshal(info)
	if err != nil {
		log.Fatalln("Failed to decode pairing info:", err.Error())
	}

	endpoint := "http://" + node.address + ":" + node.pairingPort + "/addnode"

	_, res_err := http.Post(endpoint, "applications/json", bytes.NewBuffer(payload))

	if res_err != nil {
		log.Fatalln("Failed to add node:", res_err.Error())
	}

	return nil
}
