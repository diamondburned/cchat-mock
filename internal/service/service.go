package service

import (
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/diamondburned/cchat-mock/internal/session"
	"github.com/diamondburned/cchat-mock/internal/shared"
	"github.com/diamondburned/cchat/text"
	"github.com/diamondburned/cchat/utils/empty"
	"github.com/pkg/errors"
)

type Service struct {
	empty.Service
}

var _ cchat.Service = (*Service)(nil)

func (s Service) Name() text.Rich {
	return text.Rich{Content: "Mock"}
}

func (s Service) RestoreSession(storage map[string]string) (cchat.Session, error) {
	if err := internet.SimulateAustralian(); err != nil {
		return nil, errors.Wrap(err, "Restore failed")
	}

	state, err := shared.RestoreState(storage)
	if err != nil {
		return nil, err
	}

	return session.FromState(state), nil
}

func (s Service) Authenticate() []cchat.Authenticator {
	return []cchat.Authenticator{
		Authenticator{},
		FastAuthenticator{},
	}
}

func (s Service) AsConfigurator() cchat.Configurator {
	return Configurator{}
}

func (s Service) AsSessionRestorer() cchat.SessionRestorer {
	return s
}
