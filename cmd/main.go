package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nishan-soni/k3_zeroconf/internal/agent"
	"github.com/nishan-soni/k3_zeroconf/internal/master"
	"github.com/spf13/cobra"
)

func main() {

	// CLI args parsing to check if we're running as a node or a master

	agentName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	mDNSServerPort := 5543

	rootCmd := &cobra.Command{
		Use:   "k3z",
		Short: "tdo",
	}

	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "todo",
		Run: func(cmd *cobra.Command, args []string) {
			agent.Run(agentName, mDNSServerPort, []string{})
		},
	}

	getNodesCmd := &cobra.Command{
		Use:   "list",
		Short: "todo",
		Run: func(cmd *cobra.Command, args []string) {
			if err := master.ListNodes(time.Second); err != nil {
				log.Fatalln("Failed to list nodes: %w", err.Error())
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
