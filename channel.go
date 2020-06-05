package mock

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/segments"
	"github.com/diamondburned/cchat/text"
)

func init() {}

type Channel struct {
	id       uint32
	name     string
	username text.Rich

	done   chan struct{}
	send   chan cchat.SendableMessage // ideally this should be another type
	lastID uint32
}

var (
	_ cchat.Server                     = (*Channel)(nil)
	_ cchat.ServerMessage              = (*Channel)(nil)
	_ cchat.ServerMessageSender        = (*Channel)(nil)
	_ cchat.ServerMessageSendCompleter = (*Channel)(nil)
	_ cchat.ServerNickname             = (*Channel)(nil)
)

func (ch *Channel) ID() string {
	return strconv.Itoa(int(ch.id))
}

func (ch *Channel) Name(labeler cchat.LabelContainer) error {
	labeler.SetLabel(text.Rich{Content: ch.name})
	return nil
}

func (ch *Channel) Nickname(labeler cchat.LabelContainer) error {
	labeler.SetLabel(ch.username)
	return nil
}

func (ch *Channel) JoinServer(container cchat.MessagesContainer) error {
	var lastAuthor text.Rich

	var nextID = func() uint32 {
		id := ch.lastID
		ch.lastID++
		return id
	}
	var randomMsg = func() Message {
		msg := randomMessage(nextID())
		lastAuthor = msg.author
		return msg
	}

	// Write the backlog.
	for i := 0; i < 30; i++ {
		container.CreateMessage(randomMsg())
	}

	ch.done = make(chan struct{})
	ch.send = make(chan cchat.SendableMessage)

	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()

		editTick := time.NewTicker(10 * time.Second)
		defer editTick.Stop()

		deleteTick := time.NewTicker(15 * time.Second)
		defer deleteTick.Stop()

		for {
			select {
			case msg := <-ch.send:
				container.CreateMessage(echoMessage(msg, nextID(), ch.username))
			case <-ticker.C:
				container.CreateMessage(randomMsg())
			case <-editTick.C:
				container.UpdateMessage(newRandomMessage(ch.lastID, lastAuthor))
			case <-deleteTick.C:
				container.DeleteMessage(newEmptyMessage(ch.lastID, lastAuthor))
			case <-ch.done:
				return
			}
		}
	}()

	return nil
}

func (ch *Channel) LeaveServer() error {
	ch.done <- struct{}{}
	ch.send = nil
	return nil
}

func (ch *Channel) SendMessage(msg cchat.SendableMessage) error {
	if emulateAustralianInternet() {
		return errors.New("Failed to send message: Australian Internet unsupported.")
	}

	go func() {
		// Make no guarantee that a message may arrive immediately when the
		// function exits.
		<-time.After(time.Second)
		ch.send <- msg
	}()

	return nil
}

func (ch *Channel) CompleteMessage(words []string, i int) []string {
	switch {
	case strings.HasPrefix("complete", words[i]):
		words[i] = "complete"
	case strings.HasPrefix("me", words[i]) && i > 0 && words[i-1] == "complete":
		words[i] = "me"
	default:
		return nil
	}
	return words
}

func generateChannels(s *Session, amount int) []cchat.Server {
	var channels = make([]cchat.Server, amount)
	for i := range channels {
		channels[i] = &Channel{
			id:   atomic.AddUint32(&s.lastid, 1),
			name: "#" + randomdata.Noun(),
			username: text.Rich{
				Content: s.username,
				// hot pink-ish colored
				Segments: []text.Segment{segments.NewColored(s.username, 0xE88AF8)},
			},
		}
	}
	return channels
}

// emulate network latency
func emulateAustralianInternet() (lost bool) {
	var ms = rand.Intn(internetMaxLatency-internetMinLatency) + internetMinLatency
	<-time.After(time.Duration(ms) * time.Millisecond)

	// because australia, drop packet 20% of the time if internetCanFail is
	// true.
	return internetCanFail && rand.Intn(100) < 20
}
