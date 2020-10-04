package server

import (
	"math/rand"
	"strconv"

	"github.com/Pallinder/go-randomdata"
	"github.com/diamondburned/cchat"
	"github.com/diamondburned/cchat-mock/internal/channel"
	"github.com/diamondburned/cchat-mock/internal/internet"
	"github.com/diamondburned/cchat-mock/internal/shared"
	"github.com/diamondburned/cchat/text"
	"github.com/diamondburned/cchat/utils/empty"
)

type Server struct {
	empty.Server
	state    *shared.State
	id       uint32
	name     string
	children ChannelList
}

var _ cchat.Server = (*Server)(nil)

func NewServers(state *shared.State, n int) []*Server {
	var servers = make([]*Server, n)
	for i := range servers {
		servers[i] = New(state)
	}
	return servers
}

// AsCChatServers casts a list of *Server to a list of cchat.Server.
func AsCChatServers(servers []*Server) []cchat.Server {
	var csvs = make([]cchat.Server, len(servers))
	for i, sv := range servers {
		csvs[i] = sv
	}
	return csvs
}

func New(state *shared.State) *Server {
	return &Server{
		state:    state,
		id:       state.NextID(),
		name:     randomdata.Noun(),
		children: RandomChannels(state, rand.Intn(12)+5),
	}
}

func (sv *Server) ID() string {
	return strconv.Itoa(int(sv.id))
}

func (sv *Server) Name() text.Rich {
	return text.Plain(sv.name)
}

func (sv *Server) AsLister() cchat.Lister {
	return sv.children
}

type ChannelList []cchat.Server

func RandomChannels(state *shared.State, n int) ChannelList {
	return channel.AsCChatServers(channel.NewChannels(state, n))
}

func (chl ChannelList) Servers(container cchat.ServersContainer) error {
	// IO time.
	if err := internet.SimulateAustralian(); err != nil {
		return err
	}

	container.SetServers([]cchat.Server(chl))
	return nil
}
