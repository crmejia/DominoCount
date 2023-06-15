package dominocount

import (
	"embed"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mitchellh/go-homedir"
	"html/template"
	"io"
	"net/http"
	"strconv"
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

func (s server) HandleMatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handleGetMatch(w, r)
		}

		if r.Method == http.MethodPost {
			s.handleCreateMatch(w, r)
		}
	}
}

func (s *server) HandleMatchForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, formMatchTemplate, nil)
		return
	}
}
func (s *server) handleGetMatch(w http.ResponseWriter, r *http.Request) {
	matchId := mux.Vars(r)["id"]
	if matchId == "" {
		http.Error(w, "no match ID provided", http.StatusBadRequest)
		return
	}

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

func (s server) handleCreateMatch(w http.ResponseWriter, r *http.Request) {
	team1Name := r.PostFormValue("team1_name")
	team2Name := r.PostFormValue("team2_name")
	m, err := s.store.CreateMatch(MatchWithTeam1Name(team1Name), MatchWithTeam2Name(team2Name))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render(w, r, formMatchTemplate, nil)
		return
	}
	matchURL := fmt.Sprintf("match/%d", m.Id)
	w.Header().Set("Location", matchURL)
	w.WriteHeader(http.StatusCreated)
	return
}

type server struct {
	*http.Server
	output io.Writer
	store  sqliteStore
}

func (s *server) routes() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", s.HandleIndex())
	router.HandleFunc("/match/create", s.HandleMatchForm())
	router.HandleFunc("/match/", s.HandleMatch())
	router.HandleFunc("/match/{id}", s.HandleMatch())

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
	templatesDir      = "templates/*"
	indexTemplate     = "index.html"
	formMatchTemplate = "matchForm.html"
	matchTemplate     = "match.html"
)
