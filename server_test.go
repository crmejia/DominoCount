package dominocount_test

import (
	"dominocount"
	"github.com/phayes/freeport"

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
	server, err := dominocount.NewServer(address, os.Stdout)
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
	_, err := dominocount.NewServer("", os.Stdout)
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
	server, err := dominocount.NewServer(address, os.Stdout)
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
