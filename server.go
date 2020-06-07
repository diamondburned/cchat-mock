package mock

import (
	"math/rand"
	"strconv"
	"sync/atomic"

	"github.com/Pallinder/go-randomdata"
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat/text"
)

type Server struct {
	session  *Session
	id       uint32
	name     string
	children []cchat.Server
}

var (
	_ cchat.Server     = (*Server)(nil)
	_ cchat.ServerList = (*Server)(nil)
)

func (sv *Server) ID() string {
	return strconv.Itoa(int(sv.id))
}

func (sv *Server) Name(labeler cchat.LabelContainer) error {
	labeler.SetLabel(text.Rich{Content: sv.name})
	return nil
}

func (sv *Server) Servers(container cchat.ServersContainer) error {
	container.SetServers(sv.children)
	return nil
}

func GenerateServers(s *Session) []cchat.Server {
	return generateServers(s, rand.Intn(45)+2)
}

func generateServers(s *Session, amount int) []cchat.Server {
	var servers = make([]cchat.Server, amount)
	for i := range servers {
		servers[i] = &Server{
			session:  s,
			id:       atomic.AddUint32(&s.lastid, 1),
			name:     randomdata.Noun(),
			children: generateChannels(s, rand.Intn(12)+2),
		}
	}
	return servers
}
