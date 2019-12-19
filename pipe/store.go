package pipe

import (
	"fmt"
	"io/ioutil"
	"path"

	"os"
	"sync"

	"github.com/minus5/go-uof-sdk"
)

func InnerFileStore(root string) InnerStage {
	return Stage(func(in <-chan *uof.Message, out chan<- *uof.Message, errc chan<- error) {
		var wg sync.WaitGroup
		for m := range in {
			out <- m
			wg.Add(1)
			go func(m *uof.Message) {
				fn := root + "/" + filename(m)
				if err := save(fn, m.Marshal()); err != nil {
					errc <- uof.Notice("file save", err)
				}
				wg.Done()
			}(m)
		}
		wg.Wait()
	})
}

func FileStore(root string) ConsumerStage {
	return func(in <-chan *uof.Message) error {
		for m := range in {
			fn := root + "/" + filename(m)
			if err := save(fn, m.Marshal()); err != nil {
				return err
			}
		}
		return nil
	}
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
			return fmt.Sprintf("/state/%s/players/%08d/%13d", m.Lang, m.Player.ID, m.RequestedAt)
		case uof.MessageTypeMarkets:
			if len(m.Markets) > 1 {
				return fmt.Sprintf("/state/%s/markets/%s/%13d", m.Lang, m.Lang, m.RequestedAt)
			}
			s := m.Markets[0]
			return fmt.Sprintf("/state/%s/markets/%08d-%08d/%13d", m.Lang, s.ID, s.VariantID, m.RequestedAt)
		case uof.MessageTypeFixture:
			return fmt.Sprintf("/state/%s/fixtures/%08d/%13d", m.Lang, m.EventID, m.RequestedAt)
		}
	case uof.MessageKindSystem:
		return fmt.Sprintf("log/system/%13d-%s/%13d", m.ReceivedAt, m.Type, m.ReceivedAt)
	}
	return fmt.Sprintf("/other/%13d-%s", m.ReceivedAt, m.Type)
}

func save(filename string, buf []byte) error {
	dir, _ := path.Split(filename)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf, 0644)
}
