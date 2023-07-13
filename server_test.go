package dominocount_test

import (
	"bytes"
	"dominocount"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/phayes/freeport"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// server tests
func TestNewServerErrorsOnEmptyAddress(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"

	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	_, err = dominocount.NewServer(store, dominocount.ServerWithAddress(""))
	if err == nil {
		t.Errorf("want error on empty server address")
	}
}

func TestNewServerErrorsOnEmptyOutput(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	_, err = dominocount.NewServer(store, dominocount.ServerWithOutput(nil))
	if err == nil {
		t.Errorf("want error on empty server address")
	}
}

// handler test
func TestIndexHandlerRendersNewGameButton(t *testing.T) {
	t.Parallel()

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}

	address := fmt.Sprintf("localhost:%d", freePort)

	tempDB := t.TempDir() + t.Name() + ".store"

	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	server, err := dominocount.NewServer(store, dominocount.ServerWithAddress(address))
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler := server.HandleIndex()
	handler(rec, req)

	result := rec.Result()
	if result.StatusCode != http.StatusOK {
		t.Errorf("want status 200 OK, got %d", result.StatusCode)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}

	want := "nuevo juego"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want index to contain %s, got:\n%s", want, got)
	}
}

func TestServer_HandleMatchFormRendersForm(t *testing.T) {
	t.Parallel()
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}

	address := fmt.Sprintf("localhost:%d", freePort)

	tempDB := t.TempDir() + t.Name() + ".store"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	server, err := dominocount.NewServer(store, dominocount.ServerWithAddress(address))

	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	//the handler doesn't user the match actually
	req := httptest.NewRequest(http.MethodGet, "/match/create", nil)

	handler := server.HandleMatchForm()
	handler(rec, req)
	//server.handleCreateMatch(rec, req)

	result := rec.Result()
	if result.StatusCode != http.StatusOK {
		t.Errorf("want status 200 OK, got %d", result.StatusCode)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}

	want := "Contar Nuevo Juego"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want index to contain %s, got:\n%s", want, got)
	}
}

func TestMatchHandlerCreatesMatch(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	server, err := dominocount.NewServer(store)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	form := strings.NewReader("team1_name=foo&team2_name=bar")
	req := httptest.NewRequest(http.MethodPost, "/match/create", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler := server.HandleMatch()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Errorf("expected status 303 SeeOther, got %d", res.StatusCode)
	}

	//todo extract id from url?
	location := strings.Split(res.Header.Get("location"), "/")
	if len(location) < 2 {
		t.Errorf("want url to contain ID")
	}
	id, err := strconv.ParseInt(location[2], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	m, err := store.GetMatchByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if m.Team1 != "foo" {
		t.Errorf("want team1 name to be foo, got %s", m.Team1)
	}
}

func TestMatchHandlerRendersMatchScore(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name() + ".store"

	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	m := dominocount.NewMatch(dominocount.MatchWithTeam1Name("foo"), dominocount.MatchWithTeam2Name("bar"))
	err = store.CreateMatch(&m)
	if err != nil {
		t.Fatal(err)
	}

	server, err := dominocount.NewServer(store)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	url := fmt.Sprintf("/match/%d", m.Id)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(m.Id, 10)})

	handler := server.HandleMatch()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, body %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "foo"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want index to contain %s\nGot:\n%s", want, got)
	}
}

func TestMatchHandlerUpdatesScore(t *testing.T) {
	t.Parallel()

	tempDB := t.TempDir() + t.Name() + ".db"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	m := dominocount.NewMatch(dominocount.MatchWithTeam1Name("foo"), dominocount.MatchWithTeam2Name("bar"))
	err = store.CreateMatch(&m)
	if err != nil {
		t.Fatal(err)
	}

	server, err := dominocount.NewServer(store)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	url := fmt.Sprintf("/match/%d", m.Id)
	form := strings.NewReader("team1_points=20&team2_points=0")
	req := httptest.NewRequest(http.MethodPatch, url, form)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(m.Id, 10)})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := server.HandleMatch()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, body %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "20"
	got := string(body)
	if !strings.Contains(got, want) {
		t.Errorf("want score to contain %s\nGot:\n%s", want, got)
	}
}

func TestStaticFilesAreBeingServed(t *testing.T) {
	t.Parallel()

	tempDB := t.TempDir() + t.Name() + ".db"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	server, err := dominocount.NewServer(store)
	if err != nil {
		t.Fatal(err)
	}

	testServer := httptest.NewServer(server.Routes())
	defer testServer.Close()

	res, err := http.Get(testServer.URL + "/static/style.css")
	if err != nil {
		t.Fatal()
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("want to status %d, got %d", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "tailwindcss"
	if !strings.Contains(string(body), want) {
		t.Errorf("expected stylesheet to contain tailwindcss")
	}
}

// Run server tests
func TestServer_RunStartsServer(t *testing.T) {
	t.Parallel()
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	address := fmt.Sprintf("localhost:%d", freePort)
	output := bytes.Buffer{}

	tempDB := t.TempDir() + t.Name() + ".store"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	server, err := dominocount.NewServer(store, dominocount.ServerWithAddress(address), dominocount.ServerWithOutput(&output))
	go server.Run()
	time.Sleep(100 * time.Millisecond)

	address = "http://" + address
	//resp, err := retryHttpGet(address)
	resp, err := http.Get(address)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Want Status %d, got: %d", http.StatusNotFound, resp.StatusCode)
	}

	want := "Contar nuevo juego"
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Not able to parse response")
	}
	if !strings.Contains(string(got), want) {
		t.Errorf("want response body to be:\n %s \ngot:\n %s", want, got)
	}
}

func TestRunServerSetsDbOnDifferentLocation(t *testing.T) {
	t.Parallel()
	dbLocation := t.TempDir()
	os.Setenv("SQLITE_VOLUME", dbLocation)

	output := &bytes.Buffer{}
	go dominocount.RunServer(output)
	time.Sleep(200 * time.Millisecond)

	dbLocation = dbLocation + "/.dominoCount.db"
	if _, err := os.Stat(dbLocation); os.IsNotExist(err) {
		t.Error("want to find db in non-default location")
	}
}
