package channel

import (
	"strconv"
	"time"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/diamondburned/cchat-mock/internal/message"
	"github.com/pkg/errors"
)

type MessageActioner struct {
	msgr *Messenger
}

var _ cchat.Actioner = (*MessageActioner)(nil)

func NewMessageActioner(msgr *Messenger) MessageActioner {
	return MessageActioner{
		msgr: msgr,
	}
}

const (
	DeleteAction        = "Delete"
	NoopAction          = "No-op"
	BestTrapAction      = "Who's the best trap?"
	TriggerTypingAction = "Trigger Typing"
)

func (msga MessageActioner) Actions(id string) []string {
	return []string{
		DeleteAction,
		NoopAction,
		BestTrapAction,
		TriggerTypingAction,
	}
}

// DoAction will be blocked by IO. As goes for every other method that
// takes a container: the frontend should call this in a goroutine.
func (msga MessageActioner) DoAction(action, messageID string) error {
	switch action {
	case DeleteAction, TriggerTypingAction:
		i, err := strconv.Atoi(messageID)
		if err != nil {
			return errors.Wrap(err, "Invalid ID")
		}

		// Simulate IO.
		if err := internet.SimulateAustralian(); err != nil {
			return err
		}

		switch action {
		case DeleteAction:
			msga.msgr.del <- message.NewHeader(uint32(i), time.Now())
		case TriggerTypingAction:
			msga.msgr.typ.TriggerTyping(msga.msgr.messages[uint32(i)].RealAuthor())
		}

	case NoopAction:
		// do nothing.

	case BestTrapAction:
		return msga.msgr.EditMessage(messageID, "Astolfo.")

	default:
		return errors.New("Unknown action.")
	}

	return nil
}
