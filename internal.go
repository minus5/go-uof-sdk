package uof

import "crypto/tls"

type Connection struct {
	Status     ConnectionStatus `json:"status"`
	Timestamp  int              `json:"timestamp,omitempty"`
	ServerName string           `json:"servername,omitempty"`
	LocalAddr  string           `json:"localaddr,omitempty"`
	Network    string           `json:"network,omitempty"`
	TLSVersion uint16           `json:"tlsversion,omitempty"`
}

func (c Connection) TLSVersionToString() string {
	switch c.TLSVersion {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	}
	return "unknown"
}

type ConnectionStatus int8

const (
	ConnectionStatusUp ConnectionStatus = iota
	ConnectionStatusDown
)

func (cs ConnectionStatus) String() string {
	switch cs {
	case ConnectionStatusDown:
		return "down"
	case ConnectionStatusUp:
		return "up"
	default:
		return "?"
	}
}
