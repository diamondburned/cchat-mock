package mock

import (
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/aqs"
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/segments"
	"github.com/diamondburned/cchat/text"

	_ "github.com/diamondburned/aqs/data"
	"github.com/diamondburned/aqs/incr"
)

const avatarURL = "https://gist.github.com/diamondburned/945744c2b5ce0aa0581c9267a4e5cf24/raw/598069da673093aaca4cd4aa0ede1a0e324e9a3a/astolfo_selfie.png"

type MessageHeader struct {
	id   uint32
	time time.Time
}

var _ cchat.MessageHeader = (*Message)(nil)

func parseID(id string) (uint32, error) {
	i, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return uint32(i), nil
}

func (m MessageHeader) ID() string {
	return strconv.Itoa(int(m.id))
}

func (m MessageHeader) Time() time.Time {
	return m.time
}

type Message struct {
	MessageHeader
	author  Author
	content string
	nonce   string
}

var (
	_ cchat.MessageCreate    = (*Message)(nil)
	_ cchat.MessageUpdate    = (*Message)(nil)
	_ cchat.MessageDelete    = (*Message)(nil)
	_ cchat.MessageNonce     = (*Message)(nil)
	_ cchat.MessageMentioned = (*Message)(nil)
)

func newEmptyMessage(id uint32, author Author) Message {
	return Message{
		MessageHeader: MessageHeader{id: id},
		author:        author,
	}
}

func newRandomMessage(id uint32, author Author) Message {
	return Message{
		MessageHeader: MessageHeader{id: id, time: time.Now()},
		author:        author,
		content:       incr.RandomQuote(author.char),
	}
}

func echoMessage(sendable cchat.SendableMessage, id uint32, author Author) Message {
	var echo = Message{
		MessageHeader: MessageHeader{id: id, time: time.Now()},
		author:        author,
		content:       sendable.Content(),
	}
	if noncer, ok := sendable.(cchat.MessageNonce); ok {
		echo.nonce = noncer.Nonce()
	}
	return echo
}

func randomMessage(id uint32) Message {
	return randomMessageWithAuthor(id, randomAuthor())
}

func randomMessageWithAuthor(id uint32, author Author) Message {
	return Message{
		MessageHeader: MessageHeader{id: id, time: time.Now()},
		author:        author,
		content:       incr.RandomQuote(author.char),
	}
}

func (m Message) Author() cchat.MessageAuthor {
	return m.author
}

func (m Message) Content() text.Rich {
	return text.Rich{Content: m.content}
}

func (m Message) Nonce() string {
	return m.nonce
}

// Mentioned is true when the message content contains the author's name.
func (m Message) Mentioned() bool {
	// hack
	return strings.Contains(m.content, m.author.name.Content)
}

type Author struct {
	name text.Rich
	char aqs.Character
}

var (
	_ cchat.MessageAuthor       = (*Author)(nil)
	_ cchat.MessageAuthorAvatar = (*Author)(nil)
)

func newAuthor(name text.Rich) Author {
	return Author{name: name}
}

func randomAuthor() Author {
	var char = aqs.RandomCharacter()
	return Author{
		char: char,
		name: text.Rich{
			Content:  char.Name,
			Segments: []text.Segment{segments.NewColorful(char.Name, char.NameColor())},
		},
	}
}

func (a Author) ID() string {
	return a.name.Content
}

func (a Author) Name() text.Rich {
	return a.name
}

func (a Author) Avatar() string {
	if a.char.ImageURL != "" {
		return a.char.ImageURL
	}
	return avatarURL
}
