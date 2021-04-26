package api

import (
	"log"
	"net/http"
	"path"

	"github.com/semrush/zenrpc/v2"
	"github.com/spf13/cobra"
	"github.com/vmkteam/mfd-generator/api/dartclient"
)

const (
	addrFlag = "port"
	pathFlag = "path"
	corsFlag = "cors"

	publicNS  = "public"
	projectNS = "project"
	xmlNS     = "xml"
	xmlVtNS   = "xml-vt"
	xmlLangNS = "xml-lang"
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

	store := &Store{}

	rpc.Register(publicNS, NewPublicService())
	rpc.Register(projectNS, NewProjectService(store))
	rpc.Register(xmlNS, NewXMLService(store))
	rpc.Register(xmlVtNS, NewXMLVTService(store))
	rpc.Register(xmlLangNS, NewXMLLangService(store))

	rpc.Use(ProjectMiddleware(store))

	router := http.NewServeMux()
	router.Handle(apiroot, rpc)
	router.Handle(docroot, http.StripPrefix(docroot, http.FileServer(http.Dir("tools/smd-box"))))
	router.Handle(docroot+"api_client.dart", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		resp, err := dartclient.NewClient(rpc.SMD()).Run()
		if err != nil {
			panic(err)
		}
		rw.Write(resp)
	}))

	log.Printf("starting server on %s\n", s.addr)

	return http.ListenAndServe(s.addr, router)
}
