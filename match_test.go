package dominocount_test

import (
	"dominocount"
	"testing"
)

func TestAddScoreIncreasesTeamScore(t *testing.T) {
	t.Parallel()
	m := dominocount.NewMatch()
	want := 20
	m.AddPoints(dominocount.Team1, want)

	got := m.Score(dominocount.Team1)
	if got != want {
		t.Errorf("want team 1 score to be %d, got %d", want, got)
	}
}

func TestGameOverAt200Points(t *testing.T) {
	t.Parallel()
	m := dominocount.NewMatch()
	m.AddPoints(dominocount.Team1, 200)

	gameOver := m.GameOver()
	if gameOver != true {
		t.Errorf("want game over at 200 points")
	}

	m = dominocount.NewMatch()
	m.AddPoints(dominocount.Team2, 200)

	gameOver = m.GameOver()
	if gameOver != true {
		t.Errorf("want game over at 200 points")
	}
}

func TestGameIsNotOverBelow200Points(t *testing.T) {
	t.Parallel()
	m := dominocount.NewMatch()
	m.AddPoints(dominocount.Team1, 199)

	gameOver := m.GameOver()
	if gameOver != false {
		t.Errorf("game shouldn't be over below 200 points")
	}

	m = dominocount.NewMatch()

	gameOver = m.GameOver()
	if gameOver != false {
		t.Errorf("game shouldn't be over below 200 point")
	}
}

func TestCannotScoreAfterGameOver(t *testing.T) {
	t.Parallel()

	m := dominocount.NewMatch()
	want := 200
	m.AddPoints(dominocount.Team1, want)
	m.AddPoints(dominocount.Team1, 10)

	got := m.Score(dominocount.Team1)
	if got != want {
		t.Errorf("score shouldn't change after game over")
	}

}

func TestCannotScoreNegativeNumbers(t *testing.T) {
	t.Parallel()
	m := dominocount.NewMatch()
	want := 0
	m.AddPoints(dominocount.Team1, -20)

	got := m.Score(dominocount.Team1)
	if got != want {
		t.Errorf("want team 2 score to be %d, got %d", want, got)
	}
}
