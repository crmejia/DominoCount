package dominocount_test

import (
	"dominocount"
	"testing"
)

func TestSQLiteStore_MatchRoundtripCreateUpdateGet(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + "matchRoundtrip.store"
	sqLiteStore, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	m := dominocount.NewMatch()
	err = sqLiteStore.CreateMatch(&m)
	if err != nil {
		t.Fatal(err)
	}
	want := "test"
	m.Team1 = want
	//m.Id = m.Id
	err = sqLiteStore.UpdateMatch(&m)
	if err != nil {
		t.Fatal(err)
	}

	got, err := sqLiteStore.GetMatchByID(m.Id)
	if err != nil {
		t.Fatal(err)
	}

	if want != got.Team1 {
		t.Errorf("want rountrip(create,update,get) test to return %s, got %s", want, got.Team1)
	}
}

func TestSQLiteStore_MatchAddPoints(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name()
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	m := dominocount.NewMatch()
	err = store.CreateMatch(&m)
	if err != nil {
		t.Fatal(err)
	}

	want := 20
	_, err = store.AddPointsByID(m.Id, 20, 0)

	if err != nil {
		t.Fatal(err)
	}

	got, err := store.GetMatchByID(m.Id)
	if err != nil {
		t.Fatal(err)
	}

	if want != got.Score1 {
		t.Errorf("want match add %d points test to return %d, got %d", want, want, got.Score1)
	}
}
func TestSQLiteStore_MatchAddPointsErrorsAfter200(t *testing.T) {
	t.Parallel()
	tempDB := t.TempDir() + t.Name()
	store, err := dominocount.OpenSQLiteStore(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	m := dominocount.NewMatch()
	err = store.CreateMatch(&m)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.AddPointsByID(m.Id, 199, 0)
	if err != nil {
		t.Error("expect no error adding 199")
	}

	_, err = store.AddPointsByID(m.Id, 10, 0)
	if err != nil {
		t.Error("expect no error adding 10")
	}

	_, err = store.AddPointsByID(m.Id, 10, 0)
	if err == nil {
		t.Error("want error adding once the max score has been reached")
	}

	_, ok := err.(*dominocount.GameOverError)
	if !ok {
		t.Errorf("want error to GameOverError got %s", err.Error())
	}

	want := 209
	got, err := store.GetMatchByID(m.Id)
	if err != nil {
		t.Fatal(err)
	}

	if want != got.Score1 {
		t.Errorf("want match to be game over and %d, got %d", want, got.Score1)
	}

}
