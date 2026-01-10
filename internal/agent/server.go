package agent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/nishan-soni/k3_zeroconf/internal/common"
)

func makeAddNodeHandler(pairingInfoCh chan<- common.PairingInfo) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		var joinRequest common.PairingInfo

		decoder := json.NewDecoder(req.Body)
		decoder.DisallowUnknownFields()

		err := decoder.Decode(&joinRequest)

		if err != nil {
			http.Error(writer, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		} else {
			slog.Info("Received join request.", "request", joinRequest)
			writer.WriteHeader(http.StatusOK)
			pairingInfoCh <- joinRequest
		}
	}
}

// Starts an HTTPS server allowing for master nodes to send their join token and ip address
// so the agent can join the cluster.
func startPairingServer() (<-chan common.PairingInfo, int, error) {
	pairingInfoCh := make(chan common.PairingInfo)

	mux := http.NewServeMux()
	mux.HandleFunc("POST " + common.JoinEndpoint, makeAddNodeHandler(pairingInfoCh))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", common.PairingServerPort))
	if err != nil {
		return nil, 0, err
	}

	port := listener.Addr().(*net.TCPAddr).Port

	slog.Info("Starting pairing server to listen for join requests from the k3s server.", "address", listener.Addr()) 
	go func() {
		if err := http.Serve(listener, mux); err != nil {
			if err != http.ErrServerClosed {
				slog.Error("HTTP server error.", slog.Any("err", err))
			}
		}
	}()

	return pairingInfoCh, port, nil
}
