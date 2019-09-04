package uof

type QueueMsg struct {
	RoutingKey string
	Body       []byte
	Timestamp  int64
}
