package shared

import (
	"context"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/pkg/errors"
)

// StaticIcon is a struct that implements cchat.Iconer. It never updates.
type StaticIcon struct {
	URL string
}

func NewStaticIcon(url string) StaticIcon {
	return StaticIcon{url}
}

func (icn StaticIcon) Icon(ctx context.Context, iconer cchat.IconContainer) (func(), error) {
	if err := internet.SimulateAustralian(); err != nil {
		return nil, errors.Wrap(err, "failed to query for icon")
	}

	iconer.SetIcon(icn.URL)
	return func() {}, nil
}
