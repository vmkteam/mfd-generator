package api

import (
	"embed"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/vmkteam/mfd-generator/api/dartclient"

	"github.com/spf13/cobra"
	"github.com/vmkteam/zenrpc/v2"
)

const (
	addrFlag = "port"
	pathFlag = "path"
	corsFlag = "cors"

	publicNS  = "public"
	projectNS = "project"
	xmlNS     = "xml"
	xmlVtNS   = "xmlvt"
	xmlLangNS = "xmllang"
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

	flags.StringP(addrFlag, "a", ":8880", "Set address to listen")

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

//go:embed web/*
var webuiFiles embed.FS

func (s *Server) Serve() error {
	apiroot := path.Join(s.path, "/api")
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
	// rpc.Register(xmlLangNS, NewXMLLangService(store))

	rpc.Use(ProjectMiddleware(store))

	router := http.NewServeMux()
	router.Handle("/", AddPrefix("web/", http.FileServer(http.FS(webuiFiles))))
	router.Handle(apiroot, rpc)
	router.Handle(docroot, http.StripPrefix(docroot, http.FileServer(http.Dir("tools/smd-box"))))
	router.Handle(docroot+"api_client.dart", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		resp, err := dartclient.NewClient(rpc.SMD()).Run()
		if err != nil {
			panic(err)
		}
		_, _ = rw.Write(resp)
	}))

	parts := strings.Split(s.addr, ":")
	if len(parts) == 2 && parts[0] == "" {
		s.addr = "localhost" + s.addr
	}

	log.Printf("starting server on http://%s\n", s.addr)

	return http.ListenAndServe(s.addr, router)
}

func AddPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := prefix + r.URL.Path
		rp := prefix + r.URL.RawPath
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = p
		r2.URL.RawPath = rp
		h.ServeHTTP(w, r2)
	})
}
