package pipe

import (
	"fmt"

	"os"
	"sync"

	"github.com/minus5/svckit/file"
	"github.com/minus5/svckit/log"
	"github.com/minus5/uof"
)

func FileStore(root string, in <-chan *uof.Message) <-chan *uof.Message {
	if err := emptyDir(root); err != nil {
		log.Error(err)
	}
	out := make(chan *uof.Message, 16)
	go func() {
		var wg sync.WaitGroup
		defer close(out)

		for m := range in {
			out <- m
			wg.Add(1)
			go func(m *uof.Message) {
				fn := root + "/" + filename(m)
				err := file.JSON(fn, m)
				if err != nil {
					log.S("filename", fn).Error(err)
				}
				wg.Done()
			}(m)
		}
		wg.Wait()
	}()
	return out
}

// filename returns unique filename for the message
func filename(m *uof.Message) string {
	switch m.Type {
	case uof.MessageTypePlayer:
		return fmt.Sprintf("/players/%08d-%s", m.Player.ID, m.Lang)
	case uof.MessageTypeMarkets:
		if len(m.Markets) > 1 {
			return fmt.Sprintf("/markets/%s", m.Lang.Code())
		}
		s := m.Markets[0]
		return fmt.Sprintf("/markets/%08d-%08d-%s", s.ID, s.VariantID, m.Lang)
	case uof.MessageTypeFixture:
		return fmt.Sprintf("/fixtures/%08d-%s", m.EventID, m.Lang)
	case uof.MessageTypeOddsChange:
		return fmt.Sprintf("/events/%d/%13d", m.EventID, m.ReceivedAt)
	default:
		if m.Lang == uof.LangNone && m.EventID != 0 {
			return fmt.Sprintf("/events/%d/%13d-%d", m.EventID, m.ReceivedAt, m.Type)
		}
		if m.Scope == uof.MessageScopeSystem {
			return fmt.Sprintf("/system/%d/%13d-%d", m.EventID, m.ReceivedAt, m.Type)
		}
		return fmt.Sprintf("/other/%13d-%d-%s", m.ReceivedAt, m.Type, m.Lang)
	}
}

func emptyDir(root string) error {
	if err := os.RemoveAll(root); err != nil {
		return err
	}
	return os.MkdirAll(root, os.ModePerm)
}
