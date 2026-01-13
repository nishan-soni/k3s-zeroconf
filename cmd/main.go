package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/nishan-soni/k3s-zeroconf/internal/agent"
	"github.com/nishan-soni/k3s-zeroconf/internal/common"
	"github.com/nishan-soni/k3s-zeroconf/internal/master"
	"github.com/spf13/cobra"
)

func main() {

	loggerOptions := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, loggerOptions))
	slog.SetDefault(logger)

	var agentAlias string
	var advertiseAddress string

	rootCmd := &cobra.Command{
		Use:   "k3z",
		Short: "Zero-config node discovery for k3s clusters.",
	}

	agentCmd := &cobra.Command{
		Use:   "agent [flags] -- [k3s flags]",
		Short: "Register this node for discovery",
		Long: `Runs the k3z agent on this node, making it discoverable via mDNS/zeroconf on the local network. 
The agent will listen for connection requests from the k3z server and can be joined to a k3s cluster. 
You can provide an alias for this node and pass additional k3s agent flags after '--'.
Ex. k3z agent --connect-ip=1 -- --node-ip=2`,
		Run: func(cmd *cobra.Command, args []string) {
			agentName := agentAlias
			if agentName == "" {
				hostname, err := os.Hostname()
				if err != nil {
					slog.Error("Failed to identify device name and no alias was provided.")
					return
				}
				agentName = hostname
			}
			a := agent.NewAgent()
			if err := a.Run(agentName, common.MDNSServerPort, args, advertiseAddress); err != nil {
				slog.Error("k3z agent failed", slog.Any("err", err))
			}
		},
	}

	discoverCmd := &cobra.Command{
		Use:   "discover",
		Short: "List available k3z agents to add to the server.",
		Long: `Scans the local network for available k3z agents using mDNS/zeroconf and lists all discovered nodes. 
	This helps the server operator see which nodes are available to be added to the k3s cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			timeout := time.Second
			fmt.Printf("Discovering Nodes (%s)\n", timeout)
			buf := &bytes.Buffer{}
			if err := master.ListNodes(buf, timeout); err != nil {
				slog.Error("Failed to discover nodes", slog.Any("error", err))
				return
			}
			os.Stdout.Write(buf.Bytes())
		},
	}

	addNodeCmd := &cobra.Command{
		Use:   "add",
		Short: "Add one or more discovered nodes to the k3s cluster.",
		Long: `Connects one or more discovered agent nodes to the k3s cluster by sending them the necessary join information. 
You can specify multiple node names to add them in a single command. 
Optionally, you can provide the server's IP address if it differs from the local system's IP.`,
		Run: func(cmd *cobra.Command, args []string) {
			masterIP := advertiseAddress
			if masterIP == "" {
				address, err := common.GetOutboundIP()
				if err != nil {
					slog.Error("Failed to get k3s server IP.")
					return
				}
				masterIP = address
			}
			for _, nodeId := range args {
				if err := master.AddNode(nodeId, masterIP, common.ServerTokenPath); err != nil {
					slog.Error("Failed to add node", "node", nodeId, slog.Any("error", err))
					return
				}
				fmt.Printf("Added %s to the cluster!\n", nodeId)
			}
		},
	}

	agentCmd.Flags().StringVar(&agentAlias, "mdns-alias", "", "Alias for the agent when registered to mDNS (defaults to device name)")
	agentCmd.Flags().StringVar(&advertiseAddress, "connect-ip", "", "Optional ip address for the agent's zeroconf connection server. Defaults to local ip.")
	addNodeCmd.Flags().StringVar(&advertiseAddress, "server-address", "", "Optional ip address that the server is running on if its different from systems local ip.")

	rootCmd.AddCommand(agentCmd, discoverCmd, addNodeCmd)
	rootCmd.Execute()

}
