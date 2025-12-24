package main

import (
	"time"

	"github.com/nishan-soni/k3_zeroconf/internal/master"
)

func main() {
	// Validate that k3s is installed. If it isnt then log the the user should install it

	// CLI args parsing to check if we're running as a node or a master

	master.ListNodes(time.Second)
}
