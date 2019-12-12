package uof

// ProducersChange type
type ProducersChange []ProducerChange

// ProducerChange information
type ProducerChange struct {
	Producer   Producer       `json:"producer,omitempty"`
	Status     ProducerStatus `json:"status,omitempty"`
	RecoveryID int            `json:"recoveryID,omitempty"`
	Timestamp  int            `json:"timestamp,omitempty"`
}

// Add *ProducersChange append
func (p *ProducersChange) Add(producer Producer, timestamp int) {
	*p = append(*p, ProducerChange{Producer: producer, Timestamp: timestamp})
}

// ProducerStatus represents status of the producer
type ProducerStatus int8

// Reason for the Producer status change
const (
	ProducerStatusDown       ProducerStatus = -1 // Producer is down
	ProducerStatusActive     ProducerStatus = 1  // Producer is active
	ProducerStatusInRecovery ProducerStatus = 2  // Producer in recovery
)
