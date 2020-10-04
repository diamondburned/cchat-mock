package shared

import (
	"errors"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/diamondburned/cchat"
)

func init() { rand.Seed(time.Now().UnixNano()) }

// ErrInvalidSession is returned if SessionRestore is given a bad session.
var ErrInvalidSession = errors.New("invalid session")

type State struct {
	SessionID string
	Username  string

	lastID uint32 // used for generation
}

var _ cchat.SessionSaver = (*State)(nil)

func NewState(username, sessionID string) *State {
	var state = &State{Username: username, SessionID: sessionID}
	if sessionID == "" {
		state.SessionID = strconv.FormatUint(rand.Uint64(), 10)
	}

	return state
}

func RestoreState(store map[string]string) (*State, error) {
	sID, ok := store["sessionID"]
	if !ok {
		return nil, ErrInvalidSession
	}

	un, ok := store["username"]
	if !ok {
		return nil, ErrInvalidSession
	}

	return NewState(un, sID), nil
}

func (s *State) NextID() uint32 {
	return atomic.AddUint32(&s.lastID, 1)
}

// ResetID resets the atomic ID counter.
func (s *State) ResetID() {
	atomic.StoreUint32(&s.lastID, 0)
}

func (s *State) SaveSession() map[string]string {
	return map[string]string{
		"sessionID": s.SessionID,
		"username":  s.Username,
	}
}
