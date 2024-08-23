package shards

type network struct{}

type values struct {
	Network network
}

// Common storage values
var Value values

// Valid DERO networks
var validNetworks = []string{
	"Mainnet",
	"Testnet",
	"Simulator",
}

// Is network a valid DERO network
func isNetworkValid(network string) bool {
	for _, n := range validNetworks {
		if network == n {
			return true
		}
	}

	return false
}

// Mainnet network value
func (network) Mainnet() string {
	return validNetworks[0]
}

// Testnet network value
func (network) Testnet() string {
	return validNetworks[1]
}

// Simulator network value
func (network) Simulator() string {
	return validNetworks[2]
}
