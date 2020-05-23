// Package mock contains a mock cchat backend.
package mock

import (
	"encoding/json"
	"strconv"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat/services"
)

func init() {
	services.RegisterService(&Service{})
}

type Service struct{}

var (
	_ cchat.Service      = (*Service)(nil)
	_ cchat.Configurator = (*Service)(nil)
)

func (s Service) Name() string {
	return "Mock"
}

func (s Service) AuthenticateForm() []cchat.AuthenticateEntry {
	return []cchat.AuthenticateEntry{{
		Name: "Username",
	}}
}

func (s Service) Authenticate(form []string) (cchat.Session, error) {
	ses := &Session{username: form[0]}
	ses.servers = GenerateServers(ses)
	return ses, nil
}

var (
	// channel.go @ emulateAustralianInternet
	internetCanFail = true
	// 500ms ~ 3s
	internetMinLatency = 500
	internetMaxLatency = 2500
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
			Key: key, Err: err,
		}
	}
	return nil
}

type Session struct {
	username string
	servers  []cchat.Server
	lastid   uint32
}

var _ cchat.Session = (*Session)(nil)

func (s *Session) Name() (string, error) {
	return s.username, nil
}

func (s *Session) Servers(container cchat.ServersContainer) error {
	container.SetServers(s.servers)
	return nil
}
