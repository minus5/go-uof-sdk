package pipe

import (
	"github.com/minus5/svckit/log"
	"github.com/minus5/uof"
)

func ToMessage(in <-chan uof.QueueMsg) <-chan *uof.Message {
	out := make(chan *uof.Message, 16)
	go func() {
		defer close(out)
		for qm := range in {
			m, err := uof.NewQueueMessage(qm.RoutingKey, qm.Timestamp, qm.Body)
			if err != nil {
				log.Error(err)
				continue
			}
			out <- m
		}
	}()
	return out
}
