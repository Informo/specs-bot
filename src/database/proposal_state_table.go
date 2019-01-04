package database

import (
	"database/sql"
	"strings"
)

const sep = ","

// Schema of the table.
const proposalStateSchema = `
-- Store proposal states
CREATE TABLE IF NOT EXISTS proposal_state (
	-- Numeric identifier of the proposal, i.e. the issue/PR's numeric ID
	number INTEGER PRIMARY KEY,
	-- Comma-separated list of labels, in the latest state of the proposal we know about.
	labels TEXT NOT NULL
);
`

const upsertStateSQL = `
	INSERT INTO proposal_state (number, labels) VALUES ($1, $2)
	ON CONFLICT (number) DO UPDATE SET labels = $2
`

const selectStateSQL = `
	SELECT labels FROM proposal_state WHERE number = $1
`

type proposalStateStatements struct {
	upsertStateStmt *sql.Stmt
	selectStateStmt *sql.Stmt
}

// Create the table if it doesn't exist and prepare the SQL statements.
func (ps *proposalStateStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(proposalStateSchema)
	if err != nil {
		return
	}
	if ps.upsertStateStmt, err = db.Prepare(upsertStateSQL); err != nil {
		return
	}
	if ps.selectStateStmt, err = db.Prepare(selectStateSQL); err != nil {
		return
	}
	return
}

// upsertState updates the state of a proposal, or inserts it if there's no
// saved state for this proposal.
// Returns an error if we couldn't talk to the database.
func (ps *proposalStateStatements) upsertState(number int64, labels []string) error {
	_, err := ps.upsertStateStmt.Exec(number, strings.Join(labels, sep))
	return err
}

// selectState retrieves the state of a proposal. Returns an empty slice if no
// state has been saved for this proposal.
// Returns an error if we couldn't talk to the database.
func (ps *proposalStateStatements) selectState(number int64) ([]string, error) {
	var s string

	if err := ps.selectStateStmt.QueryRow(number).Scan(&s); err == sql.ErrNoRows {
		return make([]string, 0), nil
	} else if err != nil {
		return nil, err
	}

	return strings.Split(s, sep), nil
}
