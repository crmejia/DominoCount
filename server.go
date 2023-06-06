package dominocount

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

func NewServer(address string, output io.Writer) (server, error) {
	if address == "" {
		return server{}, errors.New("server address cannot be empty")
	}
	server := server{
		Server: &http.Server{Addr: address},
		output: output,
	}

	server.Handler = server.routes()
	return server, nil
}

func RunServer(output io.Writer) {
	server, err := NewServer(":8080", output)
	if err != nil {
		fmt.Fprintln(output, err)
	}

	fmt.Fprintln(output, "starting http server")
	err = server.ListenAndServe()
	if err != http.ErrServerClosed {
		fmt.Fprintln(output, err)
		return
	}
}

func (s server) HandleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, indexTemplate, nil)
	}
}

func (s server) HandleMatchCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			render(w, r, createMatchTemplate, nil)
			return
		}
		//http.MethodPost
	}
}

type server struct {
	*http.Server
	output io.Writer
}

func (s *server) routes() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/", s.HandleIndex())

	return router
}

//go:embed templates
var resources embed.FS
var tmpl = template.Must(template.ParseFS(resources, templatesDir))

func render(w http.ResponseWriter, r *http.Request, templateName string, data any) {
	err := tmpl.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

const (
	templatesDir        = "templates/*"
	indexTemplate       = "index.html"
	createMatchTemplate = "createGuide.html"
)
