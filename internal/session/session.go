package session

import (
	"math/rand"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/diamondburned/cchat-mock/internal/message"
	"github.com/diamondburned/cchat-mock/internal/server"
	"github.com/diamondburned/cchat-mock/internal/shared"
	"github.com/diamondburned/cchat/text"
	"github.com/diamondburned/cchat/utils/empty"
)

type Session struct {
	empty.Session
	State      *shared.State
	ServerList []cchat.Server
}

var _ cchat.Session = (*Session)(nil)

func New(username, sessionID string) *Session {
	return FromState(shared.NewState(username, sessionID))
}

func FromState(state *shared.State) *Session {
	return &Session{
		State: state,
		ServerList: server.AsCChatServers(
			server.NewServers(state, rand.Intn(35)+10),
		),
	}
}

func (s *Session) ID() string {
	return s.State.SessionID
}

func (s *Session) Name() text.Rich {
	return text.Plain(s.State.Username)
}

func (s *Session) Disconnect() error {
	s.State.SessionID = ""
	s.State.ResetID()

	return internet.SimulateAustralian()
}

func (s *Session) Servers(container cchat.ServersContainer) error {
	if err := internet.SimulateAustralian(); err != nil {
		return err
	}

	container.SetServers(s.ServerList)
	return nil
}

func (s *Session) AsIconer() cchat.Iconer {
	return shared.NewStaticIcon(message.AvatarURL)
}

func (s *Session) AsCommander() cchat.Commander {
	return &Commander{}
}

func (s *Session) AsSessionSaver() cchat.SessionSaver {
	return s.State
}
