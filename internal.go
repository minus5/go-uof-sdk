package uof

type Connection struct {
	Status     ConnectionStatus `json:"status"`
	Timestamp  int              `json:"timestamp,omitempty"`
	ServerName string           `json:"servername,omitempty"`
	LocalAddr  string           `json:"localaddr,omitempty"`
	Network    string           `json:"network,omitempty"`
	TLSVersion uint16           `json:"tlsversion,omitempty"`
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
