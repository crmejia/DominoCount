package dominocount

import (
	"database/sql"
	"errors"

	_ "modernc.org/sqlite"
)

func OpenSQLiteStore(dbPath string) (sqliteStore, error) {
	if dbPath == "" {
		return sqliteStore{}, errors.New("db path cannot be empty")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return sqliteStore{}, err
	}

	for _, stmt := range []string{pragmaWALEnabled, pragma500BusyTimeout, pragmaForeignKeysON} {
		_, err = db.Exec(stmt, nil)
		if err != nil {
			return sqliteStore{}, err
		}
	}

	_, err = db.Exec(createMatchTable)
	if err != nil {
		return sqliteStore{}, err
	}
	store := sqliteStore{db: db}
	return store, nil

}

func (s *sqliteStore) CreateMatch(opts ...matchOption) (*match, error) {
	m := NewMatch(opts...)
	stmt, err := s.db.Prepare(insertMatch)
	if err != nil {
		return nil, err
	}

	rs, err := stmt.Exec(m.Team1, m.Team2)
	if err != nil {
		return nil, err
	}
	lastInsertID, err := rs.LastInsertId()
	if err != nil {
		return nil, err
	}
	m.Id = lastInsertID
	return &m, nil
}

func (s *sqliteStore) UpdateMatch(m *match) error {
	stmt, err := s.db.Prepare(updateMatch)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(m.Team1, m.Team2, m.score1, m.score2, m.Id)
	if err != nil {
		return err
	}
	return nil
}

func (s *sqliteStore) GetMatch(id int64) (match, error) {
	rows, err := s.db.Query(getMatch, id)
	if err != nil {
		return match{}, err
	}
	m := match{}

	for rows.Next() {
		var (
			team1Name, team2Name   string
			team1Score, team2Score int
		)
		err = rows.Scan(&team1Name, &team2Name, &team1Score, &team2Score)
		if err != nil {
			return match{}, err
		}
		m.Id = id
		m.Team1 = team1Name
		m.Team2 = team2Name
		m.score1 = team1Score
		m.score2 = team2Score
	}
	if err = rows.Err(); err != nil {
		return match{}, err
	}
	return m, nil
}

type sqliteStore struct {
	db *sql.DB
}

const pragmaWALEnabled = `PRAGMA journal_mode = WAL;`
const pragma500BusyTimeout = `PRAGMA busy_timeout = 5000;`
const pragmaForeignKeysON = `PRAGMA foreign_keys = on;`

const createMatchTable = `
CREATE TABLE IF NOT EXISTS match(
ID INTEGER NOT NULL PRIMARY KEY,
team1name TEXT  NOT NULL DEFAULT 'Team1',
team2name TEXT  NOT NULL DEFAULT 'Team2',
team1Score INTEGER NOT NULL DEFAULT 0,
team2Score INTEGER NOT NULL DEFAULT 0
);`

const insertMatch = `INSERT INTO match(team1name, team2name) VALUES (?, ?);`
const updateMatch = `UPDATE match SET team1name = ?, team2name = ?, team1score = ?, team2score = ? WHERE ID = ?;`
const getMatch = `SELECT team1name, team2name, team1score, team2score FROM  match WHERE ID = ?;`
