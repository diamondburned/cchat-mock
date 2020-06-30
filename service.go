// Package mock contains a mock cchat backend.
package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
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
	_ cchat.Icon             = (*Session)(nil)
	_ cchat.Session          = (*Session)(nil)
	_ cchat.ServerList       = (*Session)(nil)
	_ cchat.SessionSaver     = (*Session)(nil)
	_ cchat.Commander        = (*Session)(nil)
	_ cchat.CommandCompleter = (*Session)(nil)
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

func (s *Session) Icon(ctx context.Context, iconer cchat.IconContainer) (func(), error) {
	// Simulate IO while ignoring the context.
	simulateAustralianInternet()

	iconer.SetIcon(avatarURL)
	return func() {}, nil
}

func (s *Session) Save() (map[string]string, error) {
	return map[string]string{
		"username": s.username,
	}, nil
}

func (s *Session) RunCommand(cmds []string) (io.ReadCloser, error) {
	var r, w = io.Pipe()

	switch cmd := arg(cmds, 0); cmd {
	case "ls":
		go func() {
			fmt.Fprintln(w, "Commands: ls, random")
			w.Close()
		}()

	case "random":
		// callback used to generate stuff and stream into readcloser
		var generator func() string
		// number of times to generate the word
		var times = 1

		switch arg(cmds, 1) {
		case "paragraph":
			generator = randomdata.Paragraph
		case "noun":
			generator = randomdata.Noun
		case "silly_name":
			generator = randomdata.SillyName
		default:
			return nil, errors.New("Usage: random <paragraph|noun|silly_name> [repeat]")
		}

		if n := arg(cmds, 2); n != "" {
			i, err := strconv.Atoi(n)
			if err != nil {
				return nil, errors.Wrap(err, "Failed to parse repeat number")
			}
			times = i
		}

		go func() {
			defer w.Close()

			for i := 0; i < times; i++ {
				// Yes, we're simulating this even in something as trivial as a
				// command prompt.
				if err := simulateAustralianInternet(); err != nil {
					fmt.Fprintln(w, "Error:", err)
					continue
				}

				fmt.Fprintln(w, generator())
			}
		}()

	default:
		return nil, fmt.Errorf("Unknown command: %s", cmd)
	}

	return r, nil
}

func (s *Session) CompleteCommand(words []string, i int) []string {
	switch {
	case strings.HasPrefix("ls", words[i]):
		return []string{"ls"}

	case strings.HasPrefix("random", words[i]):
		return []string{
			"random paragraph",
			"random noun",
			"random silly_name",
		}

	case lookbackCheck(words, i, "random", "paragraph"):
		return []string{"paragraph"}

	case lookbackCheck(words, i, "random", "noun"):
		return []string{"noun"}

	case lookbackCheck(words, i, "random", "silly_name"):
		return []string{"silly_name"}

	default:
		return nil
	}
}

func arg(sl []string, i int) string {
	if i >= len(sl) {
		return ""
	}
	return sl[i]
}
