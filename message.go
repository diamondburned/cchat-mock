package mock

import (
	"strconv"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat/text"
)

type Message struct {
	id      uint32
	time    time.Time
	author  string
	content string
	nonce   string
}

var (
	_ cchat.MessageCreate = (*Message)(nil)
	_ cchat.MessageUpdate = (*Message)(nil)
	_ cchat.MessageDelete = (*Message)(nil)
	_ cchat.MessageNonce  = (*Message)(nil)
)

func echoMessage(sendable cchat.SendableMessage, id uint32, author string) Message {
	var echo = Message{
		id:      id,
		time:    time.Now(),
		author:  author,
		content: sendable.Content(),
	}
	if noncer, ok := sendable.(cchat.MessageNonce); ok {
		echo.nonce = noncer.Nonce()
	}
	return echo
}

func randomMessage(id uint32) Message {
	var now = time.Now()
	return Message{
		id:      id,
		time:    now,
		author:  randomdata.SillyName(),
		content: randomdata.Paragraph(),
		nonce:   now.Format(time.RFC3339Nano),
	}
}

func (m Message) ID() string {
	return strconv.Itoa(int(m.id))
}

func (m Message) Time() time.Time {
	return m.time
}

func (m Message) Author() text.Rich {
	return text.Rich{Content: m.author}
}

func (m Message) Content() text.Rich {
	return text.Rich{Content: m.content}
}

func (m Message) Nonce() string {
	return m.nonce
}
