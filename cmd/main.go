package main

import (
	"fmt"
	"os"
	"time"
	"log/slog"

	"github.com/nishan-soni/k3_zeroconf/internal/agent"
	"github.com/nishan-soni/k3_zeroconf/internal/common"
	"github.com/nishan-soni/k3_zeroconf/internal/master"
	"github.com/spf13/cobra"
)

func main() {

	var agentAlias string
	var advertiseAddress string

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	rootCmd := &cobra.Command{
		Use:   "k3z",
		Short: "todo",
	}

	agentCmd := &cobra.Command{
		Use:   "agent [flags] -- [k3s flags]",
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
			if err := agent.Run(agentName, common.MDNSServerPort, args, advertiseAddress); err != nil {
				slog.Error("Agent failed", slog.Any("err", err))
			}
		},
	}

	discoverCmd := &cobra.Command{
		Use:   "discover",
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
			masterIP := advertiseAddress
			if masterIP == "" {
				address, err := common.GetOutboundIP()
				if err != nil {
					fmt.Printf("Failed to locate master's server IP.")
				}
				masterIP = address
			}
			for _, nodeId := range args {
				if err := master.AddNode(nodeId, masterIP, common.ServerTokenPath); err != nil {
					fmt.Printf("Failed to connect %s: %s\n", nodeId, err.Error())
				}
			}
		},
	}

	agentCmd.Flags().StringVar(&agentAlias, "alias", "", "Name of the agent (defaults to device name)")
	agentCmd.Flags().StringVar(&advertiseAddress, "advertise-address", "", "Optional advertise ip address for the agent's server. (Defaults to local)")
	connectNodeCmd.Flags().StringVar(&advertiseAddress, "advertise-address", "", "Optional ip address of the master server. (Defaults to local)")

	rootCmd.AddCommand(agentCmd, discoverCmd, connectNodeCmd)
	rootCmd.Execute()

}
