package mock

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

var (
	// channel.go @ simulateAustralianInternet
	internetCanFail = true
	// 500ms ~ 3s
	internetMinLatency = 500
	internetMaxLatency = 3000
)

// ErrTimedOut is returned when the simulated IO decides to fail.
var ErrTimedOut = errors.New("Australian Internet unsupported.")

// simulate network latency
func simulateAustralianInternet() error {
	return simulateAustralianInternetCtx(context.Background())
}

func simulateAustralianInternetCtx(ctx context.Context) (err error) {
	var ms = randClamp(internetMinLatency, internetMaxLatency)

	select {
	case <-time.After(time.Duration(ms) * time.Millisecond):
		// noop
	case <-ctx.Done():
		return ctx.Err()
	}

	// because australia, drop packet 20% of the time if internetCanFail is
	// true.
	if internetCanFail && rand.Intn(100) < 20 {
		return ErrTimedOut
	}

	return nil
}
