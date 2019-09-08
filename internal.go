package uof

type Connection struct {
	Status    ConnectionStatus `json:"status,omitempty"`
	Timestamp int64            `json:"timestamp,omitempty"`
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
