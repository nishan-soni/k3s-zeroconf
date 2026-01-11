package common

const ServiceType string = "_k3-zeroconf._tcp"
const LocalDomain string = "local."
const ServerTokenPath string = "/var/lib/rancher/k3s/server/agent-token"
const MDNSServerPort int = 55435
const K3sServerPort int = 6443
const PairingServerPort int = 45873
const JoinEndpoint string = "/join"

var K3sAllowPorts []string = []string{"6443/tcp", "10250/tcp"}

var K3sAllowIps []string = []string{"10.42.0.0/16", "10.43.0.0/16"}
