package dominocount

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mitchellh/go-homedir"
)

func NewServer(address string, output io.Writer, store sqliteStore) (server, error) {
	if address == "" {
		return server{}, errors.New("server address cannot be empty")
	}

	if output == nil {
		return server{}, errors.New("output cannot be nil")
	}

	assets, err := fs.Sub(css, "static")
	if err != nil {
		return server{}, err
	}

	server := server{
		Server:     &http.Server{Addr: address},
		output:     output,
		store:      &store,
		fileServer: http.FileServer(http.FS(assets)),
	}

	server.Handler = server.Routes()
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
		return
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
	return
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
	return
}

func (s server) HandleStatic(w http.ResponseWriter, r *http.Request) {

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

func (s *server) Routes() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/", s.HandleIndex())
	router.HandleFunc("/match/create", s.HandleMatchForm())
	router.HandleFunc("/match/", s.HandleMatch())
	router.HandleFunc("/match/{id}", s.HandleMatch())
	router.Handle("/static/{file}", http.StripPrefix("/static", s.fileServer))

	return router
}

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
)
