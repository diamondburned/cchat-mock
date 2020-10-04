package service

import (
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/diamondburned/cchat-mock/internal/session"
	"github.com/pkg/errors"
)

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
	if err := internet.SimulateAustralian(); err != nil {
		return nil, errors.Wrap(err, "Authentication failed")
	}

	return session.New(form[0], ""), nil
}
