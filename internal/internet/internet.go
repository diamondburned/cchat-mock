package internet

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

var (
	// channel.go @ simulateAustralianInternet
	CanFail = true
	// 500ms ~ 3s
	MinLatency = 500
	MaxLatency = 3000
)

// ErrTimedOut is returned when the simulated IO decides to fail.
var ErrTimedOut = errors.New("Australian Internet unsupported.")

// SimulateAustralian simulates network latency with errors.
func SimulateAustralian() error {
	return SimulateAustralianCtx(context.Background())
}

// SimulateAustralianCtx simulates network latency with errors.
func SimulateAustralianCtx(ctx context.Context) (err error) {
	var ms = randClamp(MinLatency, MaxLatency)

	select {
	case <-time.After(time.Duration(ms) * time.Millisecond):
		// noop
	case <-ctx.Done():
		return ctx.Err()
	}

	// because australia, drop packet 20% of the time if internetCanFail is
	// true.
	if CanFail && rand.Intn(100) < 20 {
		return ErrTimedOut
	}

	return nil
}

func randClamp(min, max int) int {
	return rand.Intn(max-min) + min
}
