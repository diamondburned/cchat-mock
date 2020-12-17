package message

import (
	"strings"
	"time"

	"github.com/diamondburned/aqs/incr"
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat/text"

	_ "github.com/diamondburned/aqs/data"
)

type Message struct {
	Header
	author  Author
	content string
}

var (
	_ cchat.MessageCreate = (*Message)(nil)
	_ cchat.MessageUpdate = (*Message)(nil)
	_ cchat.MessageDelete = (*Message)(nil)
)

func NewEmpty(id uint32, author Author) Message {
	return Message{
		Header: Header{id: id},
		author: author,
	}
}

func NewRandomFromMessage(old Message) Message {
	return NewRandom(old.id, old.author)
}

func NewRandom(id uint32, author Author) Message {
	return Message{
		Header:  Header{id: id, time: time.Now()},
		author:  author,
		content: incr.RandomQuote(author.char),
	}
}

func Echo(sendable cchat.SendableMessage, id uint32, author Author) Message {
	var echo = Message{
		Header:  Header{id: id, time: time.Now()},
		author:  author,
		content: sendable.Content(),
	}
	return echo
}

func Random(id uint32) Message {
	return RandomWithAuthor(id, RandomAuthor())
}

func RandomWithAuthor(id uint32, author Author) Message {
	return Message{
		Header:  Header{id: id, time: time.Now()},
		author:  author,
		content: incr.RandomQuote(author.char),
	}
}

func (m Message) Author() cchat.Author {
	return m.author
}

func (m Message) RealAuthor() Author {
	return m.author
}

// AuthorName returns the message author's username in string.
func (m Message) AuthorName() string {
	return m.author.name.Content
}

func (m Message) Content() text.Rich {
	return text.Rich{Content: m.content}
}

// Mentioned is true when the message content contains the author's name.
func (m Message) Mentioned() bool {
	// hack
	return strings.Contains(m.content, m.author.name.Content)
}

func (m *Message) SetContent(content string) {
	m.content = content
}
