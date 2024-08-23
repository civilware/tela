package shards

type keys struct{}

// Common storage keys
var Key keys

// Endpoint DB storage key
func (keys) Endpoint() []byte {
	return []byte("endpoint")
}

// Network DB storage key
func (keys) Network() []byte {
	return []byte("network")
}
