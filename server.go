package dominocount

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mitchellh/go-homedir"
)

// func NewServer(address string, output io.Writer, store sqliteStore) (server, error) {
func NewServer(store sqliteStore, options ...serverOption) (server, error) {
	assets, err := fs.Sub(css, "static")
	if err != nil {
		return server{}, err
	}

	s := server{
		Server:     &http.Server{Addr: defaultAddress},
		output:     os.Stdout,
		store:      &store,
		fileServer: http.FileServer(http.FS(assets)),
	}

	for _, opt := range options {
		err := opt(&s)
		if err != nil {
			return server{}, err
		}
	}
	return s, nil
}

func ServerWithAddress(address string) serverOption {
	return func(s *server) error {
		if address == "" {
			return errors.New("address cannot be empty")
		}
		s.Addr = address
		return nil
	}
}

func ServerWithOutput(output io.Writer) serverOption {
	return func(s *server) error {
		if output == nil {
			return errors.New("output cannot be nil")
		}
		s.output = output
		return nil
	}
}

// RunServer configures and starts a dominocount server on localhost port 8080
func RunServer(output io.Writer) {
	storeDir := os.Getenv(dbVolume)
	if storeDir == "" {
		homeDir, err := homedir.Dir()
		if err != nil {
			fmt.Fprintln(output, err)
		}
		storeDir = homeDir
	}

	store, err := OpenSQLiteStore(storeDir + "/.dominoCount.db")
	if err != nil {
		fmt.Fprintln(output, err)
	}

	address := os.Getenv("ADDRESS")
	if address == "" {
		fmt.Fprintln(output, "no address provided, defaulting to :8080")
		address = "localhost:8080"
	}

	server, err := NewServer(store, ServerWithAddress(address), ServerWithOutput(output))
	if err != nil {
		fmt.Fprintln(output, err)
	}

	server.Run()

}
func (s *server) Run() {
	fmt.Fprintln(s.output, "starting http server")

	s.Handler = s.Routes()

	err := s.ListenAndServe()
	if err != http.ErrServerClosed {
		fmt.Fprintln(s.output, err)
		return
	}
}
func (s *server) Routes() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", s.HandleIndex())
	router.HandleFunc("/match/create", s.HandleMatchForm())
	router.HandleFunc("/match/", s.HandleMatch())
	router.HandleFunc("/match/{id}", s.HandleMatch())
	router.Handle("/static/{file}", http.StripPrefix("/static", s.fileServer))

	return router
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
			return
		}

		if r.Method == http.MethodPost {
			s.handleCreateMatch(w, r)
			return
		}

		if r.Method == http.MethodPatch {
			//method PATCH
			s.handlePatchMatch(w, r)
			return
		}

		http.Error(w, "method not supported", http.StatusBadRequest)

	}
}

func (s *server) HandleMatchForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, formMatchTemplate, nil)
	}
}

func (s *server) handleGetMatch(w http.ResponseWriter, r *http.Request) {
	id, err := queryStringParseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := s.store.GetMatchByID(id)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if m.Id == 0 {
		http.Error(w, "Match Not Found", http.StatusNotFound)
		return
	}

	render(w, r, matchTemplate, m)
}

func (s server) handleCreateMatch(w http.ResponseWriter, r *http.Request) {
	team1Name := r.PostFormValue("team1_name")
	team2Name := r.PostFormValue("team2_name")

	m := NewMatch(MatchWithTeam1Name(team1Name), MatchWithTeam2Name(team2Name))
	err := s.store.CreateMatch(&m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render(w, r, formMatchTemplate, nil)
		return
	}
	matchURL := fmt.Sprintf("%d", m.Id)
	http.Redirect(w, r, matchURL, http.StatusSeeOther)
}

func (s server) handlePatchMatch(w http.ResponseWriter, r *http.Request) {
	id, err := queryStringParseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	score1, err := formParseScore(r, "team1_points")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	score2, err := formParseScore(r, "team2_points")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := s.store.AddPointsByID(id, score1, score2)
	if err != nil {
		_, ok := err.(*GameOverError)
		if !ok {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		m, err = s.store.GetMatchByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	render(w, r, matchTableTemplate, m)
}

func queryStringParseID(r *http.Request) (int64, error) {
	matchId := mux.Vars(r)["id"]
	if matchId == "" {
		return 0, errors.New("no match ID provided")
	}

	id, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		return 0, errors.New("not able to parse match ID")
	}
	return id, nil

}

func formParseScore(r *http.Request, id string) (int, error) {
	scoreString := r.PostFormValue(id)
	if scoreString == "" {
		return 0, errors.New("no points provided")
	}

	score, err := strconv.ParseInt(scoreString, 10, 32)
	if err != nil {
		return 0, errors.New("not able to parse score")
	}
	return int(score), nil

}

type server struct {
	*http.Server
	output     io.Writer
	store      Storage
	fileServer http.Handler
}

type serverOption func(*server) error

//go:embed templates
var resources embed.FS

//go:embed static *.css
var css embed.FS

var tmpl = template.Must(template.ParseFS(resources, templatesDir))

func render(w http.ResponseWriter, r *http.Request, templateName string, data any) {
	err := tmpl.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

const (
	templatesDir       = "templates/*"
	indexTemplate      = "index.html"
	formMatchTemplate  = "matchForm.html"
	matchTemplate      = "match.html"
	matchTableTemplate = "matchTable.html"

	dbVolume = "SQLITE_VOLUME"

	defaultAddress = "localhost:8080"
)
