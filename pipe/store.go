package pipe

import (
	"fmt"

	"os"
	"sync"

	"github.com/minus5/svckit/file"
	"github.com/minus5/uof"
)

func FileStore(root string) stage {
	return Stage(func(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) {
		var wg sync.WaitGroup
		for m := range in {
			out <- m
			wg.Add(1)
			go func(m *uof.Message) {
				fn := root + "/" + filename(m)
				if err := file.Save(fn, m.Marshal()); err != nil {
					errc <- err
				}
				wg.Done()
			}(m)
		}
		wg.Wait()
	})
}

// filename returns unique filename for the message
func filename(m *uof.Message) string {
	switch m.Type.Kind() {
	case uof.MessageKindEvent:
		if m.Type == uof.MessageTypeOddsChange {
			return fmt.Sprintf("/log/events/%d/%13d", m.EventID, m.ReceivedAt)
		}
		return fmt.Sprintf("/log/events/%d/%13d-%s", m.EventID, m.ReceivedAt, m.Type)
	case uof.MessageKindLexicon:
		switch m.Type {
		case uof.MessageTypePlayer:
			return fmt.Sprintf("/state/%s/players/%08d", m.Lang, m.Player.ID)
		case uof.MessageTypeMarkets:
			if len(m.Markets) > 1 {
				return fmt.Sprintf("/state/%s/markets/%s", m.Lang, m.Lang)
			}
			s := m.Markets[0]
			return fmt.Sprintf("/state/%s/markets/%08d-%08d", m.Lang, s.ID, s.VariantID)
		case uof.MessageTypeFixture:
			return fmt.Sprintf("/state/%s/fixtures/%08d", m.Lang, m.EventID)
		}
	case uof.MessageKindSystem:
		return fmt.Sprintf("log/system/%13d-%s", m.ReceivedAt, m.Type)
	}
	return fmt.Sprintf("/other/%13d-%s", m.ReceivedAt, m.Type)
}

func emptyDir(root string) error {
	if err := os.RemoveAll(root); err != nil {
		return err
	}
	return os.MkdirAll(root, os.ModePerm)
}
