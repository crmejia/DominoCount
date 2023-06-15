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
