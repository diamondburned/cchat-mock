package message

import (
	"strconv"
	"time"

	"github.com/diamondburned/cchat"
)

type Header struct {
	id   uint32
	time time.Time
}

var _ cchat.MessageHeader = (*Message)(nil)

func ParseID(id string) (uint32, error) {
	i, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return uint32(i), nil
}

func NewHeader(id uint32, t time.Time) Header {
	return Header{
		id:   id,
		time: t,
	}
}

func (m Header) ID() string {
	return strconv.Itoa(int(m.id))
}

func (m Header) RealID() uint32 {
	return m.id
}

func (m Header) Time() time.Time {
	return m.time
}
