package api

import (
	"log"
	"net/http"
	"path"

	"github.com/spf13/cobra"
	"github.com/vmkteam/zenrpc"
)

const (
	addrFlag = "port"
	pathFlag = "path"
	corsFlag = "cors"
)

type Server struct {
	addr string
	path string
	cors bool
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) AddFlags(command *cobra.Command) {
	flags := command.Flags()
	flags.SortFlags = false

	flags.StringP(addrFlag, "a", ":8080", "Set address to listen")

	flags.String(pathFlag, "/", "Set path to handle")

	flags.Bool(corsFlag, false, "Allow CORS")
}

func (s *Server) ReadFlags(command *cobra.Command) (err error) {
	flags := command.Flags()

	s.addr, err = flags.GetString(addrFlag)
	if err != nil {
		return err
	}

	s.path, err = flags.GetString(pathFlag)
	if err != nil {
		return err
	}

	s.cors, err = flags.GetBool(corsFlag)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Serve() error {
	p := path.Join(pathFlag, "/")

	rpc := zenrpc.NewServer(zenrpc.Options{
		ExposeSMD: true,
		AllowCORS: s.cors,
		TargetURL: p,
	})

	http.Handle(p, rpc)

	log.Printf("starting server on %s\n", s.addr)

	return http.ListenAndServe(s.addr, nil)
}
