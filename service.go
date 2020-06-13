// Package mock contains a mock cchat backend.
package mock

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat/services"
	"github.com/diamondburned/cchat/text"
	"github.com/pkg/errors"
)

func init() {
	services.RegisterService(&Service{})
}

// ErrInvalidSession is returned if SessionRestore is given a bad session.
var ErrInvalidSession = errors.New("invalid session")

type Service struct{}

var (
	_ cchat.Service         = (*Service)(nil)
	_ cchat.Configurator    = (*Service)(nil)
	_ cchat.SessionRestorer = (*Service)(nil)
)

func (s Service) Name() text.Rich {
	return text.Rich{Content: "Mock"}
}

func (s Service) RestoreSession(storage map[string]string) (cchat.Session, error) {
	if err := simulateAustralianInternet(); err != nil {
		return nil, errors.Wrap(err, "Restore failed")
	}

	username, ok := storage["username"]
	if !ok {
		return nil, ErrInvalidSession
	}

	return newSession(username), nil
}

func (s Service) Authenticate() cchat.Authenticator {
	return Authenticator{}
}

type Authenticator struct{}

var _ cchat.Authenticator = (*Authenticator)(nil)

func (Authenticator) AuthenticateForm() []cchat.AuthenticateEntry {
	return []cchat.AuthenticateEntry{
		{
			Name: "Username",
		},
		{
			Name:   "Password (ignored)",
			Secret: true,
		},
		{
			Name:      "Paragraph (ignored)",
			Multiline: true,
		},
	}
}

func (Authenticator) Authenticate(form []string) (cchat.Session, error) {
	// SLOW IO TIME.
	if err := simulateAustralianInternet(); err != nil {
		return nil, errors.Wrap(err, "Authentication failed")
	}

	return newSession(form[0]), nil
}

var (
	// channel.go @ simulateAustralianInternet
	internetCanFail = true
	// 500ms ~ 3s
	internetMinLatency = 500
	internetMaxLatency = 3000
)

func (s Service) Configuration() (map[string]string, error) {
	return map[string]string{
		"internet.canFail":    strconv.FormatBool(internetCanFail),
		"internet.minLatency": strconv.Itoa(internetMinLatency),
		"internet.maxLatency": strconv.Itoa(internetMaxLatency),
	}, nil
}

func (s Service) SetConfiguration(config map[string]string) error {
	for _, err := range []error{
		// shit code, would not recommend. It's only an ok-ish idea here because
		// unmarshalConfig() returns ErrInvalidConfigAtField.
		unmarshalConfig(config, "internet.canFail", &internetCanFail),
		unmarshalConfig(config, "internet.minLatency", &internetMinLatency),
		unmarshalConfig(config, "internet.maxLatency", &internetMaxLatency),
	} {
		if err != nil {
			return err
		}
	}
	return nil
}

func unmarshalConfig(config map[string]string, key string, value interface{}) error {
	if err := json.Unmarshal([]byte(config[key]), value); err != nil {
		return &cchat.ErrInvalidConfigAtField{
			Key: key,
			Err: err,
		}
	}
	return nil
}

type Session struct {
	username string
	servers  []cchat.Server
	lastid   uint32 // used for generation
}

var (
	_ cchat.Icon         = (*Session)(nil)
	_ cchat.Session      = (*Session)(nil)
	_ cchat.SessionSaver = (*Session)(nil)
)

func newSession(username string) *Session {
	ses := &Session{username: username}
	ses.servers = GenerateServers(ses)
	return ses
}

func (s *Session) ID() string {
	return s.username
}

func (s *Session) Name() text.Rich {
	return text.Rich{Content: s.username}
}

func (s *Session) Disconnect() error {
	// Nothing to do here, but emulate errors.
	return simulateAustralianInternet()
}

func (s *Session) Servers(container cchat.ServersContainer) error {
	// Simulate slight IO.
	<-time.After(time.Second)

	container.SetServers(s.servers)
	return nil
}

func (s *Session) Icon(iconer cchat.IconContainer) error {
	// Simulate IO.
	simulateAustralianInternet()

	iconer.SetIcon(avatarURL)
	return nil
}

func (s *Session) Save() (map[string]string, error) {
	return map[string]string{
		"username": s.username,
	}, nil
}
