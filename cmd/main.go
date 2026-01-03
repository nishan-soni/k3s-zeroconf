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

	agentName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	rootCmd := &cobra.Command{
		Use:   "k3z",
		Short: "todo",
	}

	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "todo",
		Run: func(cmd *cobra.Command, args []string) {
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
			if err := master.AddNode(args[0], "test", "test"); err != nil {
				fmt.Printf("Failed to connect %s: %s\n", args[0], err.Error())
			}
		},
	}

	rootCmd.AddCommand(agentCmd, getNodesCmd, connectNodeCmd)
	rootCmd.Execute()

}
