package mock

import (
	"math/rand"

	"github.com/Pallinder/go-randomdata"
	"github.com/diamondburned/cchat"
)

type Server struct {
	session  *Service
	name     string
	children []cchat.Server
}

var (
	_ cchat.Server     = (*Server)(nil)
	_ cchat.ServerList = (*Server)(nil)
)

func (sv *Server) Name() (string, error) {
	return sv.name, nil
}

func (sv *Server) Servers(container cchat.ServersContainer) error {
	container.SetServers(sv.children)
	return nil
}

func GenerateServers(s *Service) []cchat.Server {
	return generateServers(s, rand.Intn(45))
}

func generateServers(s *Service, amount int) []cchat.Server {
	var channels = make([]cchat.Server, amount)
	for i := range channels {
		channels[i] = &Server{
			session:  s,
			name:     randomdata.Noun(),
			children: generateChannels(s, rand.Intn(12)),
		}
	}
	return channels
}
