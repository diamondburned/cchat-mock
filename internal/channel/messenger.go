package channel

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/diamondburned/cchat-mock/internal/message"
	"github.com/diamondburned/cchat-mock/internal/typing"
	"github.com/diamondburned/cchat/utils/empty"
	"github.com/pkg/errors"
)

// FetchBacklog is the number of messages to fake-fetmsgr.
const FetchBacklog = 35
const maxBacklog = FetchBacklog * 2

// max number to add to before the next author, with rand.Intn(limit) + incr.
const sameAuthorLimit = 6

type Messenger struct {
	empty.Messenger
	channel *Channel

	send MessageSender
	edit chan message.Message // id
	del  chan message.Header
	typ  typing.Subscriber

	messageMutex sync.Mutex
	messages     map[uint32]message.Message
	messageids   []uint32 // indices

	// used for unique ID generation of messages
	incrID uint32
	// used for generating the same author multiple times before shuffling, goes
	// up to about 12 or so. check sameAuthorLimit.
	incrAuthor uint8
}

var _ cchat.Messenger = (*Messenger)(nil)

func (msgr *Messenger) JoinServer(ctx context.Context, ct cchat.MessagesContainer) (func(), error) {
	// Is this a fresh channel? If yes, generate messages with some IO latency.
	if len(msgr.messageids) == 0 || msgr.messages == nil {
		// Simulate IO and error.
		if err := internet.SimulateAustralianCtx(ctx); err != nil {
			return nil, err
		}

		// Initialize.
		msgr.messages = make(map[uint32]message.Message, FetchBacklog)
		msgr.messageids = make([]uint32, 0, FetchBacklog)

		// Allocate 3 channels that we won't clean up, because we're lazy.
		msgr.send = NewMessageSender(msgr)
		msgr.edit = make(chan message.Message)
		msgr.del = make(chan message.Header)
		msgr.typ = typing.NewSubscriber(message.NewAuthor(msgr.channel.user.Rich()))

		// Generate the backlog.
		for i := 0; i < FetchBacklog; i++ {
			msgr.addMessage(msgr.randomMsg(), ct)
		}
	} else {
		// Else, flush the old backlog over.
		for i := range msgr.messages {
			ct.CreateMessage(msgr.messages[i])
		}
	}

	// Initialize context for cancellation. The context passed in is used only
	// for initialization, so we'll use our own context for the loop.
	ctx, stop := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()

		editTick := time.NewTicker(10 * time.Second)
		defer editTick.Stop()

		// deleteTick := time.NewTicker(15 * time.Second)
		// defer deleteTick.Stop()

		for {
			select {
			case msg := <-msgr.send.ch:
				msgr.addMessage(message.Echo(
					msg,
					msgr.nextID(),
					message.NewAuthor(msgr.channel.user.Rich()),
				), ct)

			case msg := <-msgr.edit:
				ct.UpdateMessage(msg)

			case msh := <-msgr.del:
				msgr.deleteMessage(msh, ct)

			case <-ticker.C:
				msgr.addMessage(msgr.randomMsg(), ct)

			case <-editTick.C:
				var old = msgr.randomOldMsg()
				msgr.updateMessage(message.NewRandomFromMessage(old), ct)

			// case <-deleteTick.C:
			// 	var old = msgr.randomOldMsg()
			// 	msgr.deleteMessage(message.Header{old.id, time.Now()}, container)

			case <-ctx.Done():
				return
			}
		}
	}()

	return stop, nil
}

func (msgr *Messenger) nextID() (id uint32) {
	return atomic.AddUint32(&msgr.incrID, 1)
}

func (msgr *Messenger) AsEditor() cchat.Editor { return msgr }

// MessageEditable returns true if the message belongs to the author.
func (msgr *Messenger) MessageEditable(id string) bool {
	i, err := message.ParseID(id)
	if err != nil {
		return false
	}

	msgr.messageMutex.Lock()
	defer msgr.messageMutex.Unlock()

	m, ok := msgr.messages[i]
	if ok {
		// Editable if same author.
		return m.Author().Name().String() == msgr.channel.user.String()
	}

	return false
}

func (msgr *Messenger) RawMessageContent(id string) (string, error) {
	i, err := message.ParseID(id)
	if err != nil {
		return "", err
	}

	msgr.messageMutex.Lock()
	defer msgr.messageMutex.Unlock()

	m, ok := msgr.messages[i]
	if ok {
		return m.Content().String(), nil
	}

	return "", errors.New("Message not found")
}

func (msgr *Messenger) EditMessage(id, content string) error {
	i, err := message.ParseID(id)
	if err != nil {
		return err
	}

	if err := internet.SimulateAustralian(); err != nil {
		return err
	}

	msgr.messageMutex.Lock()
	defer msgr.messageMutex.Unlock()

	m, ok := msgr.messages[i]
	if ok {
		m.SetContent(content)
		msgr.messages[i] = m
		msgr.edit <- m

		return nil
	}

	return errors.New("Message not found.")
}

func (msgr *Messenger) addMessage(msg message.Message, container cchat.MessagesContainer) {
	msgr.messageMutex.Lock()

	// Clean up the backlog.
	if clean := len(msgr.messages) - maxBacklog; clean > 0 {
		// Remove them from the map.
		for _, id := range msgr.messageids[:clean] {
			delete(msgr.messages, id)
		}

		// Cut the message IDs away by shifting the slice.
		msgr.messageids = append(msgr.messageids[:0], msgr.messageids[clean:]...)
	}

	msgr.messages[msg.RealID()] = msg
	msgr.messageids = append(msgr.messageids, msg.RealID())

	msgr.messageMutex.Unlock()

	container.CreateMessage(msg)
}

func (msgr *Messenger) updateMessage(msg message.Message, container cchat.MessagesContainer) {
	msgr.messageMutex.Lock()

	_, ok := msgr.messages[msg.RealID()]
	if ok {
		msgr.messages[msg.RealID()] = msg
	}

	msgr.messageMutex.Unlock()

	if ok {
		container.UpdateMessage(msg)
	}
}

func (msgr *Messenger) deleteMessage(msg message.Header, container cchat.MessagesContainer) {
	msgr.messageMutex.Lock()

	// Delete from the map.
	delete(msgr.messages, msg.RealID())

	// Delete from the ordered slice.
	var ok bool
	for i, id := range msgr.messageids {
		if id == msg.RealID() {
			msgr.messageids = append(msgr.messageids[:i], msgr.messageids[i+1:]...)
			ok = true
			break
		}
	}

	msgr.messageMutex.Unlock()

	if ok {
		container.DeleteMessage(msg)
	}
}

// randomMsgID returns a random recent message ID.
func (msgr *Messenger) randomOldMsg() message.Message {
	msgr.messageMutex.Lock()
	defer msgr.messageMutex.Unlock()

	// Pick a random index from last, clamped to 10 and len channel.
	n := len(msgr.messageids) - 1 - rand.Intn(len(msgr.messageids))%10
	return msgr.messages[msgr.messageids[n]]
}

// randomMsg uses top of the state algorithms to return fair and balanced
// messages suitable for rigorous testing.
func (msgr *Messenger) randomMsg() (msg message.Message) {
	msgr.messageMutex.Lock()
	defer msgr.messageMutex.Unlock()

	// If we don't have any messages, then skip.
	if len(msgr.messages) == 0 {
		return message.Random(msgr.nextID())
	}

	// Add a random number into incrAuthor and determine if that should be
	// enough to generate a new author.
	msgr.incrAuthor += uint8(rand.Intn(sameAuthorLimit)) // 1~6 appearances

	var lastID = msgr.messageids[len(msgr.messageids)-1]
	var lastAu = msgr.messages[lastID].RealAuthor()

	// If the last author is not the current user, then we can use it.
	// Should we generate a new author for the new message? No if we're not over
	// the limits.
	if !lastAu.Equal(msg.RealAuthor()) && msgr.incrAuthor < sameAuthorLimit {
		msg = message.RandomWithAuthor(msgr.nextID(), lastAu)
	} else {
		msg = message.Random(msgr.nextID())
		msgr.incrAuthor = 0 // reset
	}

	return
}

func (msgr *Messenger) AsSender() cchat.Sender {
	return msgr.send
}

func (msgr *Messenger) AsActioner() cchat.Actioner {
	return &MessageActioner{msgr}
}

func (msgr *Messenger) AsTypingIndicator() cchat.TypingIndicator {
	return msgr.typ
}
