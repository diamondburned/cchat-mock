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
)

type Channel struct {
	session *Session
	id      uint32
	name    string
	done    chan struct{}
	send    chan Message // ideally this should be another type
	lastID  uint32
}

var (
	_ cchat.Server                     = (*Channel)(nil)
	_ cchat.ServerMessage              = (*Channel)(nil)
	_ cchat.ServerMessageSender        = (*Channel)(nil)
	_ cchat.ServerMessageSendCompleter = (*Channel)(nil)
)

func (ch *Channel) ID() string {
	return strconv.Itoa(int(ch.id))
}

func (ch *Channel) Name() (string, error) {
	return ch.name, nil
}

func (ch *Channel) JoinServer(container cchat.MessagesContainer) error {
	nextid := func() uint32 {
		return atomic.AddUint32(&ch.lastID, 1)
	}

	// Write the backlog.
	for i := 0; i < 30; i++ {
		container.CreateMessage(randomMessage(nextid()))
	}

	ch.done = make(chan struct{})
	go func() {
		ticker := time.Tick(10 * time.Second)
		for {
			select {
			case <-ticker:
				container.CreateMessage(randomMessage(nextid()))
			case msg := <-ch.send:
				container.CreateMessage(msg)
			case <-ch.done:
				return
			}
		}
	}()

	return nil
}

func (ch *Channel) LeaveServer() error {
	ch.done <- struct{}{}
	return nil
}

func (ch *Channel) SendMessage(msg cchat.SendableMessage) error {
	if emulateAustralianInternet() {
		return errors.New("Failed to send message: Australian Internet unsupported.")
	}

	ch.send <- echoMessage(msg, atomic.AddUint32(&ch.lastID, 1), ch.session.username)
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
			session: s,
			id:      atomic.AddUint32(&s.lastid, 1),
			name:    "#" + randomdata.Noun(),
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
