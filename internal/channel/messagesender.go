package channel

import (
	"time"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/pkg/errors"
)

type MessageSender struct {
	msgr *Messenger
	ch   chan cchat.SendableMessage
}

var _ cchat.Sender = (*MessageSender)(nil)

func NewMessageSender(msgr *Messenger) MessageSender {
	return MessageSender{
		msgr: msgr,
		ch:   make(chan cchat.SendableMessage),
	}
}

// CanAttach returns false.
func (msgs MessageSender) CanAttach() bool { return false }

func (msgs MessageSender) Send(msg cchat.SendableMessage) error {
	if err := internet.SimulateAustralian(); err != nil {
		return errors.Wrap(err, "Failed to send message")
	}

	go func() {
		// Make no guarantee that a message may arrive immediately when the
		// function exits.
		<-time.After(time.Second)
		msgs.ch <- msg
	}()

	return nil
}

func (msgs MessageSender) AsCompleter() cchat.Completer {
	return &MessageCompleter{msgr: msgs.msgr}
}
