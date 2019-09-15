package uof

// Alive messages are sent by each producer every 10 seconds. This is indicating
// that the given producer is operating normally and you are able to receive
// messages from it.
// Reference: https://docs.betradar.com/display/BD/UOF+-+Alive
type Alive struct {
	//	The producer that sent this alive message.
	Producer Producer `xml:"product,attr" json:"producer"`
	// Timestamp in milliseconds since epoch when this message was generated
	// according to generating system's clock.
	Timestamp int `xml:"timestamp,attr" json:"timestamp"`
	// If set to 0 this means the product is up again after downtime, and the
	// receiving client will have to issue recovery messages against the API to
	// start receiving any additional messages and get the current state.
	Subscribed int `xml:"subscribed,attr" json:"subscribed"`
}
