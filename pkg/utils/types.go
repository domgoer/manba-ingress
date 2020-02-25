package utils

import "fmt"

// Endpoint describes a kubernetes endpoint, same as a server addr in Manba.
type Endpoint struct {
	// Address IP address of the endpoint
	Address string `json:"address"`
	// Port number of the TCP port
	Port string `json:"port"`
}

// String tansfer to <ip>:<port>
func (e *Endpoint) String() string {
	return fmt.Sprintf("%s:%s", e.Address, e.Port)
}
