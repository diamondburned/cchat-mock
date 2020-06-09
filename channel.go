package mock

import (
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/segments"
	"github.com/diamondburned/cchat/text"
	"github.com/pkg/errors"
)

// FetchBacklog is the number of messages to fake-fetch.
const FetchBacklog = 35
const maxBacklog = FetchBacklog * 2

// max number to add to before the next author, with rand.Intn(limit) + incr.
const sameAuthorLimit = 12

type Channel struct {
	id       uint32
	name     string
	username text.Rich

	send chan cchat.SendableMessage // ideally this should be another type
	edit chan Message               // id

	messageMutex sync.Mutex
	messages     []Message
	messageixs   map[uint32]int // indices

	// used for unique ID generation of messages
	incrID uint32
	// used for generating the same author multiple times before shuffling, goes
	// up to about 12 or so. check sameAuthorLimit.
	incrAuthor uint8

	busyWg sync.WaitGroup
}

var (
	_ cchat.Server                     = (*Channel)(nil)
	_ cchat.ServerMessage              = (*Channel)(nil)
	_ cchat.ServerMessageSender        = (*Channel)(nil)
	_ cchat.ServerMessageSendCompleter = (*Channel)(nil)
	_ cchat.ServerNickname             = (*Channel)(nil)
	_ cchat.ServerMessageEditor        = (*Channel)(nil)
	_ cchat.ServerMessageActioner      = (*Channel)(nil)
)

func (ch *Channel) ID() string {
	return strconv.Itoa(int(ch.id))
}

func (ch *Channel) Name() text.Rich {
	return text.Rich{Content: ch.name}
}

func (ch *Channel) Nickname(labeler cchat.LabelContainer) error {
	// Simulate IO.
	simulateAustralianInternet()

	labeler.SetLabel(ch.username)
	return nil
}

func (ch *Channel) JoinServer(container cchat.MessagesContainer) (func(), error) {
	// Is this a fresh channel? If yes, generate messages with some IO latency.
	if len(ch.messages) == 0 || ch.messageixs == nil {
		// Simulate IO.
		simulateAustralianInternet()

		// Initialize.
		ch.messageixs = make(map[uint32]int, FetchBacklog)
		ch.messages = make([]Message, 0, FetchBacklog)

		// Allocate 2 channels that we won't clean up, because we're lazy.
		ch.send = make(chan cchat.SendableMessage)
		ch.edit = make(chan Message)

		// Generate the backlog.
		for i := 0; i < FetchBacklog; i++ {
			ch.addMessage(randomMessage(ch.nextID()), container)
		}
	} else {
		// Else, flush the old backlog over.
		for i := range ch.messages {
			container.CreateMessage(ch.messages[i])
		}
	}

	// Initialize channels for use.
	doneCh := make(chan struct{})

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
				ch.addMessage(echoMessage(msg, ch.nextID(), ch.username), container)

			case msg := <-ch.edit:
				container.UpdateMessage(msg)

			case <-ticker.C:
				ch.addMessage(ch.randomMsg(), container)

			case <-editTick.C:
				var old = ch.randomOldMsg()
				ch.updateMessage(newRandomMessage(old.id, old.author), container)

			case <-deleteTick.C:
				var old = ch.randomOldMsg()
				ch.deleteMessage(MessageHeader{old.id, time.Now()}, container)

			case <-doneCh:
				return
			}
		}
	}()

	return func() { doneCh <- struct{}{} }, nil
}

func (ch *Channel) RawMessageContent(id string) (string, error) {
	i, err := parseID(id)
	if err != nil {
		return "", err
	}

	ch.messageMutex.Lock()
	defer ch.messageMutex.Unlock()

	ix, ok := ch.messageixs[i]
	if !ok {
		return "", errors.New("Message not found")
	}

	return ch.messages[ix].content, nil
}

func (ch *Channel) EditMessage(id, content string) error {
	i, err := parseID(id)
	if err != nil {
		return err
	}

	simulateAustralianInternet()

	ch.messageMutex.Lock()
	defer ch.messageMutex.Unlock()

	ix, ok := ch.messageixs[i]
	if !ok {
		return errors.New("Message not found.")
	}

	m := ch.messages[ix]
	m.content = content

	ch.messages[ix] = m
	ch.edit <- m

	return nil
}

func (ch *Channel) addMessage(msg Message, container cchat.MessagesContainer) {
	ch.messageMutex.Lock()

	// Clean up the backlog.
	if clean := len(ch.messages) - maxBacklog; clean > 0 {
		// Remove them from the map.
		for _, m := range ch.messages[:clean] {
			delete(ch.messageixs, m.id)
		}

		// Cut the message IDs away by shifting the slice.
		ch.messages = append(ch.messages[:0], ch.messages[clean:]...)
	}

	ch.messageixs[msg.id] = len(ch.messages)
	ch.messages = append(ch.messages, msg)

	ch.messageMutex.Unlock()

	container.CreateMessage(msg)
}

func (ch *Channel) updateMessage(msg Message, container cchat.MessagesContainer) {
	ch.messageMutex.Lock()

	i, ok := ch.messageixs[msg.id]
	if ok {
		ch.messages[i] = msg
	}

	ch.messageMutex.Unlock()

	if ok {
		container.UpdateMessage(msg)
	}
}

func (ch *Channel) deleteMessage(msg MessageHeader, container cchat.MessagesContainer) {
	ch.messageMutex.Lock()

	i, ok := ch.messageixs[msg.id]
	if ok {
		ch.messages = append(ch.messages[:i], ch.messages[i+1:]...)
		delete(ch.messageixs, msg.id)
	}

	ch.messageMutex.Unlock()

	if ok {
		container.DeleteMessage(msg)
	}
}

// randomMsgID returns a random recent message ID.
func (ch *Channel) randomOldMsg() Message {
	ch.messageMutex.Lock()
	defer ch.messageMutex.Unlock()

	// Pick a random number, clamped to 10 and len channel.
	n := rand.Intn(len(ch.messages)) % 10
	return ch.messages[n]
}

// randomMsg uses top of the state algorithms to return fair and balanced
// messages suitable for rigorous testing.
func (ch *Channel) randomMsg() (msg Message) {
	ch.messageMutex.Lock()
	defer ch.messageMutex.Unlock()

	// If we don't have any messages, then skip.
	if len(ch.messages) == 0 {
		return randomMessage(ch.nextID())
	}

	// Add a random number into incrAuthor and determine if that should be
	// enough to generate a new author.
	ch.incrAuthor += uint8(rand.Intn(5)) // 2~4 appearances

	// Should we generate a new author for the new message?
	if ch.incrAuthor > sameAuthorLimit {
		msg = randomMessage(ch.nextID())
		ch.incrAuthor = 0 // reset
	} else {
		last := ch.messages[len(ch.messages)-1]
		msg = randomMessageWithAuthor(ch.nextID(), last.author)
	}

	return
}

func (ch *Channel) nextID() (id uint32) {
	return atomic.AddUint32(&ch.incrID, 1)
}

func (ch *Channel) SendMessage(msg cchat.SendableMessage) error {
	if simulateAustralianInternet() {
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

const (
	DeleteAction   = "Delete"
	NoopAction     = "No-op"
	BestTrapAction = "What's the best trap?"
)

func (ch *Channel) MessageActions() []string {
	return []string{
		DeleteAction,
		NoopAction,
		BestTrapAction,
	}
}

// DoMessageAction will be blocked by IO. As goes for every other method that
// takes a container: the frontend should call this in a goroutine.
func (ch *Channel) DoMessageAction(c cchat.MessagesContainer, action, messageID string) error {
	switch action {
	case DeleteAction:
		i, err := strconv.Atoi(messageID)
		if err != nil {
			return errors.Wrap(err, "Invalid ID")
		}

		// Simulate IO.
		simulateAustralianInternet()
		ch.deleteMessage(MessageHeader{uint32(i), time.Now()}, c)

	case NoopAction:
		// do nothing.

	case BestTrapAction:
		return ch.EditMessage(messageID, "Astolfo.")

	default:
		return errors.New("Unknown action.")
	}

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

func randClamp(min, max int) int {
	return rand.Intn(max-min) + min
}

// simulate network latency
func simulateAustralianInternet() (lost bool) {
	var ms = randClamp(internetMinLatency, internetMaxLatency)
	<-time.After(time.Duration(ms) * time.Millisecond)

	// because australia, drop packet 20% of the time if internetCanFail is
	// true.
	return internetCanFail && rand.Intn(100) < 20
}
