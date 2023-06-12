package dominocount

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/mitchellh/go-homedir"
)

func NewServer(address string, output io.Writer, store sqliteStore) (server, error) {
	if address == "" {
		return server{}, errors.New("server address cannot be empty")
	}

	if output == nil {
		return server{}, errors.New("output cannot be nil")
	}

	server := server{
		Server: &http.Server{Addr: address},
		output: output,
		store:  store,
	}

	server.Handler = server.routes()
	return server, nil
}

func RunServer(output io.Writer) {
	homeDir, err := homedir.Dir()
	if err != nil {
		fmt.Fprintln(output, err)
	}
	store, err := OpenSQLiteStore(homeDir + "/.dominoCount.db")
	if err != nil {
		fmt.Fprintln(output, err)
	}

	server, err := NewServer(":8080", output, store)
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
		team1Name := r.PostFormValue("team1name")
		team2Name := r.PostFormValue("team2name")
		m, err := s.store.CreateMatch(MatchWithTeam1Name(team1Name), MatchWithTeam2Name(team2Name))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render(w, r, createMatchTemplate, nil)
			return
		}
		matchURL := fmt.Sprintf("match/%d", m.Id)
		w.Header().Set("Location", matchURL)
		w.WriteHeader(http.StatusCreated)
		return
	}
}

func (s server) HandleMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			p := strings.Split(r.URL.Path, "/")
			if len(p) < 3 {
				http.Error(w, "no match ID provided", http.StatusBadRequest)
				return
			}

			matchId := p[2]
			id, err := strconv.ParseInt(matchId, 10, 64)
			if err != nil {
				http.Error(w, "not able to parse match ID", http.StatusBadRequest)
				return
			}

			m, err := s.store.GetMatch(id)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if m.Id == 0 {
				http.Error(w, "Match Not Found", http.StatusNotFound)
				return
			}

			render(w, r, matchTemplate, m)
			return
		}
		//http.MethodPost
	}
}

type server struct {
	*http.Server
	output io.Writer
	store  sqliteStore
}

func (s *server) routes() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/", s.HandleIndex())
	router.HandleFunc("/match/", s.HandleMatch())
	router.HandleFunc("/match/create/", s.HandleMatchCreate())

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
	createMatchTemplate = "createMatch.html"
	matchTemplate       = "match.html"
)
