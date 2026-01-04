package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nishan-soni/k3_zeroconf/internal/agent"
	"github.com/nishan-soni/k3_zeroconf/internal/common"
	"github.com/nishan-soni/k3_zeroconf/internal/master"
	"github.com/spf13/cobra"
)

func main() {

	// CLI args parsing to check if we're running as a node or a master

	var agentAlias string

	rootCmd := &cobra.Command{
		Use:   "k3z",
		Short: "todo",
	}

	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "todo",
		Run: func(cmd *cobra.Command, args []string) {
			agentName := agentAlias
			if agentName == "" {
				hostname, err := os.Hostname()
				if err != nil {
					fmt.Println("Failed to identify device name and no alias was provided.")
					return
				}
				agentName = hostname
			}
			if err := agent.Run(agentName, common.MDNSServerPort, []string{}); err != nil {
				fmt.Printf("Agent failed: %s\n", err)
			}
		},
	}

	getNodesCmd := &cobra.Command{
		Use:   "list",
		Short: "todo",
		Run: func(cmd *cobra.Command, args []string) {
			if err := master.ListNodes(time.Second); err != nil {
				fmt.Printf("Failed to list nodes: %s\n", err.Error())
			}
		},
	}

	connectNodeCmd := &cobra.Command{
		Use:   "connect",
		Short: "todo",
		Run: func(cmd *cobra.Command, args []string) {
			masterIP, err := common.GetOutboundIP()
			if err != nil {
				fmt.Printf("Failed to locate master's server IP.")
			}
			for _, nodeId := range args {
				if err := master.AddNode(nodeId, masterIP, common.ServerTokenPath); err != nil {
					fmt.Printf("Failed to connect %s: %s\n", nodeId, err.Error())
				}
			}
		},
	}

	agentCmd.Flags().StringVar(&agentAlias, "alias", "", "Name of the agent (defaults to device name)")

	rootCmd.AddCommand(agentCmd, getNodesCmd, connectNodeCmd)
	rootCmd.Execute()

}
