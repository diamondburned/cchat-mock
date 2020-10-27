package service

import (
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/diamondburned/cchat-mock/internal/session"
	"github.com/diamondburned/cchat/text"
	"github.com/pkg/errors"
)

type Authenticator struct{}

var _ cchat.Authenticator = (*Authenticator)(nil)

func (Authenticator) Name() text.Rich {
	return text.Plain("Slow Authentication")
}

func (Authenticator) Description() text.Rich {
	return text.Plain("")
}

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

func (Authenticator) Authenticate(form []string) (cchat.Session, cchat.AuthenticateError) {
	// SLOW IO TIME.
	err := internet.SimulateAustralian()
	if err == nil {
		return session.New(form[0], ""), nil
	}

	return nil, cchat.WrapAuthenticateError(errors.Wrap(err, "Authentication failed"))
}

type FastAuthenticator struct{}

var _ cchat.Authenticator = (*FastAuthenticator)(nil)

func (FastAuthenticator) Name() text.Rich {
	return text.Plain("Fast Authenticator")
}

func (FastAuthenticator) Description() text.Rich {
	return text.Plain("Internet fails and slow-downs disabled.")
}

func (FastAuthenticator) AuthenticateForm() []cchat.AuthenticateEntry {
	return []cchat.AuthenticateEntry{
		{
			Name: "Username (fast)",
		},
	}
}

func (FastAuthenticator) Authenticate(form []string) (cchat.Session, cchat.AuthenticateError) {
	return session.New(form[0], ""), nil
}
