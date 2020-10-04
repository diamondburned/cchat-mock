package channel

import (
	"strconv"

	"github.com/Pallinder/go-randomdata"
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/shared"
	"github.com/diamondburned/cchat-mock/segments"
	"github.com/diamondburned/cchat/text"
	"github.com/diamondburned/cchat/utils/empty"
)

type Channel struct {
	empty.Server
	id   uint32
	name string
	user Username

	messenger *Messenger
}

var _ cchat.Server = (*Channel)(nil)

func NewChannels(state *shared.State, n int) []*Channel {
	var channels = make([]*Channel, n)
	for i := range channels {
		channels[i] = NewChannel(state)
	}
	return channels
}

func AsCChatServers(channels []*Channel) []cchat.Server {
	var cchs = make([]cchat.Server, len(channels))
	for i, ch := range channels {
		cchs[i] = ch
	}
	return cchs
}

// NewChannel creates a new random channel.
func NewChannel(state *shared.State) *Channel {
	return &Channel{
		id:   state.NextID(),
		name: "#" + randomdata.Noun(),
		user: Username{
			Content: state.Username,
			Segments: []text.Segment{
				// hot pink-ish colored
				segments.NewColoredSegment(state.Username, 0xE88AF8),
			},
		},
	}
}

func (ch *Channel) ID() string {
	return strconv.Itoa(int(ch.id))
}

func (ch *Channel) Name() text.Rich {
	return text.Plain(ch.name)
}

func (ch *Channel) AsNicknamer() cchat.Nicknamer {
	return ch.user
}

func (ch *Channel) AsMessenger() cchat.Messenger {
	return ch.messenger
}
