package mock

import (
	"context"
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
const sameAuthorLimit = 6

type Channel struct {
	id       uint32
	name     string
	username text.Rich

	send chan cchat.SendableMessage // ideally this should be another type
	edit chan Message               // id
	del  chan MessageHeader
	typ  chan Author

	messageMutex sync.Mutex
	messages     map[uint32]Message
	messageids   []uint32 // indices

	// used for unique ID generation of messages
	incrID uint32
	// used for generating the same author multiple times before shuffling, goes
	// up to about 12 or so. check sameAuthorLimit.
	incrAuthor uint8
}

var (
	_ cchat.Server                       = (*Channel)(nil)
	_ cchat.ServerMessage                = (*Channel)(nil)
	_ cchat.ServerMessageSender          = (*Channel)(nil)
	_ cchat.ServerMessageSendCompleter   = (*Channel)(nil)
	_ cchat.ServerNickname               = (*Channel)(nil)
	_ cchat.ServerMessageEditor          = (*Channel)(nil)
	_ cchat.ServerMessageActioner        = (*Channel)(nil)
	_ cchat.ServerMessageTypingIndicator = (*Channel)(nil)
)

func (ch *Channel) ID() string {
	return strconv.Itoa(int(ch.id))
}

func (ch *Channel) Name() text.Rich {
	return text.Rich{Content: ch.name}
}

// Nickname sets the labeler to the nickname. It simulates heavy IO. This
// function stops as cancel is called in JoinServer, as Nickname is specially
// for that.
//
// The given context is cancelled.
func (ch *Channel) Nickname(ctx context.Context, labeler cchat.LabelContainer) (func(), error) {
	// Simulate IO with cancellation. Ignore the error if it's a simulated time
	// out, else return.
	if err := simulateAustralianInternetCtx(ctx); err != nil && err != ErrTimedOut {
		return nil, err
	}

	labeler.SetLabel(ch.username)
	return func() {}, nil
}

func (ch *Channel) JoinServer(ctx context.Context, ct cchat.MessagesContainer) (func(), error) {
	// Is this a fresh channel? If yes, generate messages with some IO latency.
	if len(ch.messageids) == 0 || ch.messages == nil {
		// Simulate IO and error.
		if err := simulateAustralianInternetCtx(ctx); err != nil {
			return nil, err
		}

		// Initialize.
		ch.messages = make(map[uint32]Message, FetchBacklog)
		ch.messageids = make([]uint32, 0, FetchBacklog)

		// Allocate 3 channels that we won't clean up, because we're lazy.
		ch.send = make(chan cchat.SendableMessage)
		ch.edit = make(chan Message)
		ch.del = make(chan MessageHeader)
		ch.typ = make(chan Author)

		// Generate the backlog.
		for i := 0; i < FetchBacklog; i++ {
			ch.addMessage(ch.randomMsg(), ct)
		}
	} else {
		// Else, flush the old backlog over.
		for i := range ch.messages {
			ct.CreateMessage(ch.messages[i])
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
			case msg := <-ch.send:
				ch.addMessage(echoMessage(msg, ch.nextID(), newAuthor(ch.username)), ct)

			case msg := <-ch.edit:
				ct.UpdateMessage(msg)

			case msh := <-ch.del:
				ch.deleteMessage(msh, ct)

			case <-ticker.C:
				ch.addMessage(ch.randomMsg(), ct)

			case <-editTick.C:
				var old = ch.randomOldMsg()
				ch.updateMessage(newRandomMessage(old.id, old.author), ct)

			// case <-deleteTick.C:
			// 	var old = ch.randomOldMsg()
			// 	ch.deleteMessage(MessageHeader{old.id, time.Now()}, container)

			case <-ctx.Done():
				return
			}
		}
	}()

	return stop, nil
}

// MessageEditable returns true if the message belongs to the author.
func (ch *Channel) MessageEditable(id string) bool {
	i, err := parseID(id)
	if err != nil {
		return false
	}

	ch.messageMutex.Lock()
	defer ch.messageMutex.Unlock()

	m, ok := ch.messages[i]
	if ok {
		// Editable if same author.
		return m.author.name.Content == ch.username.Content
	}

	return false
}

func (ch *Channel) RawMessageContent(id string) (string, error) {
	i, err := parseID(id)
	if err != nil {
		return "", err
	}

	ch.messageMutex.Lock()
	defer ch.messageMutex.Unlock()

	m, ok := ch.messages[i]
	if ok {
		return m.content, nil
	}

	return "", errors.New("Message not found")
}

func (ch *Channel) EditMessage(id, content string) error {
	i, err := parseID(id)
	if err != nil {
		return err
	}

	simulateAustralianInternet()

	ch.messageMutex.Lock()
	defer ch.messageMutex.Unlock()

	m, ok := ch.messages[i]
	if ok {
		m.content = content
		ch.messages[i] = m
		ch.edit <- m

		return nil
	}

	return errors.New("Message not found.")
}

func (ch *Channel) addMessage(msg Message, container cchat.MessagesContainer) {
	ch.messageMutex.Lock()

	// Clean up the backlog.
	if clean := len(ch.messages) - maxBacklog; clean > 0 {
		// Remove them from the map.
		for _, id := range ch.messageids[:clean] {
			delete(ch.messages, id)
		}

		// Cut the message IDs away by shifting the slice.
		ch.messageids = append(ch.messageids[:0], ch.messageids[clean:]...)
	}

	ch.messages[msg.id] = msg
	ch.messageids = append(ch.messageids, msg.id)

	ch.messageMutex.Unlock()

	container.CreateMessage(msg)
}

func (ch *Channel) updateMessage(msg Message, container cchat.MessagesContainer) {
	ch.messageMutex.Lock()

	_, ok := ch.messages[msg.id]
	if ok {
		ch.messages[msg.id] = msg
	}

	ch.messageMutex.Unlock()

	if ok {
		container.UpdateMessage(msg)
	}
}

func (ch *Channel) deleteMessage(msg MessageHeader, container cchat.MessagesContainer) {
	ch.messageMutex.Lock()

	// Delete from the map.
	delete(ch.messages, msg.id)

	// Delete from the ordered slice.
	var ok bool
	for i, id := range ch.messageids {
		if id == msg.id {
			ch.messageids = append(ch.messageids[:i], ch.messageids[i+1:]...)
			ok = true
			break
		}
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

	// Pick a random index from last, clamped to 10 and len channel.
	n := len(ch.messageids) - 1 - rand.Intn(len(ch.messageids))%10
	return ch.messages[ch.messageids[n]]
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
	ch.incrAuthor += uint8(rand.Intn(sameAuthorLimit)) // 1~6 appearances

	var lastID = ch.messageids[len(ch.messageids)-1]
	var lastAu = ch.messages[lastID].author

	// If the last author is not the current user, then we can use it.
	// Should we generate a new author for the new message? No if we're not over
	// the limits.
	if lastAu.name.Content != ch.username.Content && ch.incrAuthor < sameAuthorLimit {
		msg = randomMessageWithAuthor(ch.nextID(), lastAu)
	} else {
		msg = randomMessage(ch.nextID())
		ch.incrAuthor = 0 // reset
	}

	return
}

func (ch *Channel) nextID() (id uint32) {
	return atomic.AddUint32(&ch.incrID, 1)
}

func (ch *Channel) SendMessage(msg cchat.SendableMessage) error {
	if err := simulateAustralianInternet(); err != nil {
		return errors.Wrap(err, "Failed to send message")
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
	DeleteAction        = "Delete"
	NoopAction          = "No-op"
	BestTrapAction      = "Who's the best trap?"
	TriggerTypingAction = "Trigger Typing"
)

func (ch *Channel) MessageActions(id string) []string {
	return []string{
		DeleteAction,
		NoopAction,
		BestTrapAction,
		TriggerTypingAction,
	}
}

// DoMessageAction will be blocked by IO. As goes for every other method that
// takes a container: the frontend should call this in a goroutine.
func (ch *Channel) DoMessageAction(action, messageID string) error {
	switch action {
	case DeleteAction, TriggerTypingAction:
		i, err := strconv.Atoi(messageID)
		if err != nil {
			return errors.Wrap(err, "Invalid ID")
		}

		// Simulate IO.
		simulateAustralianInternet()

		switch action {
		case DeleteAction:
			ch.del <- MessageHeader{uint32(i), time.Now()}
		case TriggerTypingAction:
			ch.typ <- ch.messages[uint32(i)].author
		}

	case NoopAction:
		// do nothing.

	case BestTrapAction:
		return ch.EditMessage(messageID, "Astolfo.")

	default:
		return errors.New("Unknown action.")
	}

	return nil
}

func (ch *Channel) CompleteMessage(words []string, i int) (entries []cchat.CompletionEntry) {
	switch {
	case strings.HasPrefix("complete", words[i]):
		entries = makeCompletion(
			"complete",
			"complete me",
			"complete you",
			"complete everyone",
		)

	case lookbackCheck(words, i, "complete", "me"):
		entries = makeCompletion("me")

	case lookbackCheck(words, i, "complete", "you"):
		entries = makeCompletion("you")

	case lookbackCheck(words, i, "complete", "everyone"):
		entries = makeCompletion("everyone")

	case lookbackCheck(words, i, "best", "trap:"):
		entries = makeCompletion(
			"trap: Astolfo",
			"trap: Hackadoll No. 3",
			"trap: Totsuka",
			"trap: Felix Argyle",
		)

	default:
		var found = map[string]struct{}{}

		ch.messageMutex.Lock()
		defer ch.messageMutex.Unlock()

		// Look for members.
		for _, id := range ch.messageids {
			if msg := ch.messages[id]; strings.HasPrefix(msg.author.name.Content, words[i]) {
				if _, ok := found[msg.author.name.Content]; ok {
					continue
				}

				found[msg.author.name.Content] = struct{}{}

				entries = append(entries, cchat.CompletionEntry{
					Raw:     msg.author.name.Content,
					Text:    msg.author.name,
					IconURL: avatarURL,
				})
			}
		}
	}

	return
}

func makeCompletion(word ...string) []cchat.CompletionEntry {
	var entries = make([]cchat.CompletionEntry, len(word))
	for i, w := range word {
		entries[i].Raw = w
		entries[i].Text.Content = w
		entries[i].IconURL = avatarURL
	}
	return entries
}

// completion will only override `this'.
func lookbackCheck(words []string, i int, prev, this string) bool {
	return strings.HasPrefix(this, words[i]) && i > 0 && words[i-1] == prev
}

// Typing sleeps and returns possibly an error.
func (ch *Channel) Typing() error {
	return simulateAustralianInternet()
}

// TypingTimeout returns 5 seconds.
func (ch *Channel) TypingTimeout() time.Duration {
	return 5 * time.Second
}

type Typer struct {
	Author
	time time.Time
}

var _ cchat.Typer = (*Typer)(nil)

func newTyper(a Author) *Typer   { return &Typer{a, time.Now()} }
func randomTyper() *Typer        { return &Typer{randomAuthor(), time.Now()} }
func (t *Typer) Time() time.Time { return t.time }

func (ch *Channel) TypingSubscribe(ti cchat.TypingIndicator) (stop func(), err error) {
	var stopch = make(chan struct{})

	go func() {
		var ticker = time.NewTicker(8 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopch:
				return
			case <-ticker.C:
				ti.AddTyper(randomTyper())
			case author := <-ch.typ:
				ti.AddTyper(newTyper(author))
			}
		}
	}()

	return func() { close(stopch) }, nil
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
