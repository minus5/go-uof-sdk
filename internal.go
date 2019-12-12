package uof

// Connection is used to describe connection status
type Connection struct {
	Status    ConnectionStatus `json:"status"`
	Timestamp int              `json:"timestamp,omitempty"`
}

// ConnectionStatus returns status of the connection
type ConnectionStatus int8

// Connection status values
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
