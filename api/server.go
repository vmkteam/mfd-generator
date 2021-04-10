package api

import (
	"log"
	"net/http"
	"path"

	"github.com/semrush/zenrpc/v2"
	"github.com/spf13/cobra"

	"github.com/vmkteam/mfd-generator/api/dartclient"
)

//go:generate zenrpc

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
	apiroot := path.Join(s.path, "/")
	docroot := path.Join(s.path, "/doc") + "/"

	rpc := zenrpc.NewServer(zenrpc.Options{
		ExposeSMD: true,
		AllowCORS: s.cors,
		TargetURL: apiroot,
	})

	rpc.Register("xml", NewXMLService())
	rpc.Register("api", NewMockService())

	router := http.NewServeMux()
	router.Handle(apiroot, rpc)
	router.Handle(docroot, http.StripPrefix(docroot, http.FileServer(http.Dir("tools/smd-box"))))
	router.Handle(path.Join(docroot, "/api_client.dart"), s.handleDart(rpc))

	log.Printf("starting server on %s\n", s.addr)

	return http.ListenAndServe(s.addr, router)
}

// handleDart is a handler for dart schema.
func (s *Server) handleDart(srv zenrpc.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bb, err := dartclient.NewClient(srv.SMD()).Run()
		if err != nil {
			log.Printf("failed to convert dart err=%q", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		_, err = w.Write(bb)
		if err != nil {
			log.Printf("failed to write dart err=%q", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
