package channel

import (
	"context"

	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/diamondburned/cchat/text"
)

// Username is the type for a username/nickname.
type Username text.Rich

func (u Username) String() string {
	return text.Rich(u).String()
}

func (u Username) Rich() text.Rich {
	return text.Rich(u)
}

// Nickname sets the labeler to the nickname. It simulates heavy IO. This
// function stops as cancel is called in JoinServer, as Nickname is specially
// for that.
func (u Username) Nickname(ctx context.Context, labeler cchat.LabelContainer) (func(), error) {
	if err := internet.SimulateAustralianCtx(ctx); err != nil {
		return nil, err
	}

	labeler.SetLabel(text.Rich(u))
	return func() {}, nil
}
