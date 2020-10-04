package typing

import (
	"time"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/diamondburned/cchat-mock/internal/message"
)

type Typer struct {
	message.Author
	time time.Time
}

var _ cchat.Typer = (*Typer)(nil)

func NewTyper(a message.Author) *Typer {
	return &Typer{Author: a, time: time.Now()}
}

func RandomTyper() *Typer {
	return &Typer{
		Author: message.RandomAuthor(),
		time:   time.Now(),
	}
}

func (t *Typer) Time() time.Time {
	return t.time
}

type Subscriber struct {
	self     message.Author
	incoming chan message.Author
}

func NewSubscriber(self message.Author) Subscriber {
	return Subscriber{
		self:     self,
		incoming: make(chan message.Author),
	}
}

func (ts Subscriber) TriggerTyping(author message.Author) {
	ts.incoming <- author
}

func (ts Subscriber) TypingSubscribe(ti cchat.TypingContainer) (func(), error) {
	var stopch = make(chan struct{})

	go func() {
		var ticker = time.NewTicker(8 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopch:
				return
			case <-ticker.C:
				ti.AddTyper(RandomTyper())
			case author := <-ts.incoming:
				ti.AddTyper(NewTyper(author))
			}
		}
	}()

	return func() { close(stopch) }, nil
}

// Typing sleeps and returns possibly an error.
func (ts Subscriber) Typing() error {
	if err := internet.SimulateAustralian(); err != nil {
		return err
	}
	ts.TypingNow()
	return nil
}

// TypingNow sends a typing event immediately.
func (ts Subscriber) TypingNow() {
	ts.TriggerTyping(ts.self)
}

// TypingTimeout returns 5 seconds.
func (ts Subscriber) TypingTimeout() time.Duration {
	return 5 * time.Second
}
