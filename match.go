package dominocount

type match struct {
	Score1, Score2 int
	Team1, Team2   string
	Id             int64
}
type matchOption func(*match) error

func NewMatch(opts ...matchOption) match {
	m := match{
		Team1:  string(Team1),
		Team2:  string(Team2),
		Score1: 0,
		Score2: 0,
	}

	for _, opt := range opts {
		//ignoring errors as current name option generates no errors
		opt(&m)
	}
	return m
}

func MatchWithTeam1Name(name string) matchOption {
	return func(m *match) error {
		if name != "" {
			m.Team1 = name
		}
		return nil
	}
}

func MatchWithTeam2Name(name string) matchOption {
	return func(m *match) error {
		if name != "" {
			m.Team2 = name
		}
		return nil
	}
}

func (m *match) AddPoints(t team, points int) {
	if m.GameOver() {
		return
	}
	if points < 0 {
		return
	}
	if t == Team1 {
		m.Score1 += points
		return
	}
	m.Score2 += points
}

func (m match) Score(t team) int {
	if t == Team1 {
		return m.Score1
	}
	return m.Score2
}

func (m match) GameOver() bool {
	if m.Score1 >= 200 || m.Score2 >= 200 {
		return true
	}
	return false
}

type team string

const (
	Team1 team = "Team1"
	Team2 team = "Team2"
)
