package master

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"
	"net/http"

	"github.com/grandcat/zeroconf"
	"github.com/nishan-soni/k3_zeroconf/internal/common"
)

type DiscoveredNode struct {
	address string
	id      string
}

/*
Connection flow:

	node announces its self in mDNS

	master runs `mytool node list` to view all the nodes in mDNS
	master runs 'mytool node connect mynode-1' to connect to the node
		-> master sends an http request to mynode-1 with its ip and connection token
		-> mynode-1 receieves the http request and then runs the k3 agent command
		-> node is now connected to the cluster
*/

func ListNodes(timeout time.Duration) {
	resolver, err := zeroconf.NewResolver(nil)

	if err != nil {
		log.Fatalln("Failed to init zeroconf resolver:", err.Error())
	}

	var results []*zeroconf.ServiceEntry
	services := make(chan *zeroconf.ServiceEntry)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		err := resolver.Browse(ctx, common.ServiceType, common.LocalDomain, services)
		if err != nil {
			log.Fatalln("Zeroconf browsing failed:", err.Error())
		}
	}()

	fmt.Printf("Discovering Nodes (%s)\n", timeout)

	for entry := range services {
		results = append(results, entry)
	}

	printNodesTable(results)
}

func printNodesTable(discoveredNodes []*zeroconf.ServiceEntry) {
	// Prints the nodes to stdout as a table

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(writer, "NAME\tIP\tPORT")

	for _, node := range discoveredNodes {
		ip := "Unknown"
		if len(node.AddrIPv4) > 0 {
			ip = node.AddrIPv4[0].String()
		}

		fmt.Fprintf(writer, "%s\t%s\t%d\n", node.Instance, ip, node.Port)
	}

	writer.Flush()
}

func AddNode(node *DiscoveredNode) {

	// Sends an http request to the node to the cluster

	// Maybe it should check if it was actually added? Idk if it should be done here though

	// Errors that it raises:
	//	Node no longer available (because its already connected to smth else or it just died)
	//		This would prob be raised on a 404 error
	// 	Other connection failures
	//		Add a context timeout thing to check after 30s if the node was added or not
	//			Can check using the k3 cli command most likley

	endpoint := node.address
	_, err := http.Get(endpoint)

	if err != nil {
		log.Fatalln("Failed to add node:", err.Error())
	}


}
