package dominocount

type match struct {
	score1, score2 int
}

func NewMatch() match {
	return match{
		score1: 0,
		score2: 0,
	}
}

func (m *match) AddPoints(t team, points int) {
	if m.GameOver() {
		return
	}
	if t == Team1 {
		m.score1 += points
		return
	}
	m.score2 += points
}

func (m match) Score(t team) int {
	if t == Team1 {
		return m.score1
	}
	return m.score2
}

func (m match) GameOver() bool {
	if m.score1 >= 200 || m.score2 >= 200 {
		return true
	}
	return false
}

type team string

const (
	Team1 team = "team1"
	Team2 team = "team2"
)
