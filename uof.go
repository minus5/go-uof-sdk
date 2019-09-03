package uof

import "time"

type QueueMsg struct {
	RoutingKey string
	Body       []byte
	Timestamp  time.Time
}
