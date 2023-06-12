package dominocount_test

import (
	"dominocount"
	"github.com/phayes/freeport"
	"strconv"

	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestIndexHandlerRendersNewGameButton(t *testing.T) {
	t.Parallel()

	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}

	address := fmt.Sprintf("localhost:%d", freePort)

	tempDB := t.TempDir() + "indexhandler.store"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	server, err := dominocount.NewServer(address, os.Stdout, store)
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

func TestNewServerErrorsOnEmptyAddress(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "error.store"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	_, err = dominocount.NewServer("", os.Stdout, store)
	if err == nil {
		t.Errorf("want error on empty server address")
	}
}

func TestCreateMatchHandlerGetRendersForm(t *testing.T) {
	t.Parallel()
	freePort, err := freeport.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}

	address := fmt.Sprintf("localhost:%d", freePort)

	tempDB := t.TempDir() + "handleMatchCreateGet.store"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	server, err := dominocount.NewServer(address, os.Stdout, store)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/match/create", nil)

	handler := server.HandleMatchCreate()
	handler(rec, req)

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

func TestCreateMatchHandlerPostCreatesMatch(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "matchHandlerPost.store"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}

	server, err := dominocount.NewServer("localhost:8080", os.Stdout, store)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	form := strings.NewReader("team1name=foo&team2name=bar")
	req := httptest.NewRequest(http.MethodPost, "/match/create", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler := server.HandleMatchCreate()
	handler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected status 303 SeeOther, got %d", res.StatusCode)
	}

	//todo extract id from url?
	location := strings.Split(res.Header.Get("Location"), "/")
	if len(location) < 2 {
		t.Errorf("want url to contain ID")
	}
	id, err := strconv.ParseInt(location[1], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	m, err := store.GetMatch(id)
	if err != nil {
		t.Fatal(err)
	}
	if m.Team1 != "foo" {
		t.Errorf("want team1 name to be foo, got %s", m.Team1)
	}
}

func TestMatchHandlerRendersMatchScore(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "match.store"
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	m, err := store.CreateMatch(dominocount.MatchWithTeam1Name("foo"), dominocount.MatchWithTeam2Name("bar"))
	if err != nil {
		t.Fatal(err)
	}

	server, err := dominocount.NewServer("localhost:8080", os.Stdout, store)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	url := fmt.Sprintf("/match/%d", m.Id)
	req := httptest.NewRequest(http.MethodGet, url, nil)
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
