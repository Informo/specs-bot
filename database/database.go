package database

import (
	"database/sql"

	"github.com/Informo/specs-bot/config"

	// Database drivers
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/sirupsen/logrus"
)

// Database represents the crawler's database.
type Database struct {
	db            *sql.DB
	proposalState proposalStateStatements
}

// NewDatabase creates a new instance of the Database structure by opening a
// PostgreSQL database accessible using a given connexion configuration string,
// and preparing the different statements used.
// Returns an error if there was an issue opening the database or preparing the
// different statements.
func NewDatabase(cfg *config.Config) (database *Database, err error) {
	database = new(Database)

	if database.db, err = sql.Open(cfg.Database.Driver, cfg.Database.DataSource); err != nil {
		return
	}
	if err = database.proposalState.prepare(database.db); err != nil {
		return
	}

	return
}

// UpdateProposalState updates the state of a proposal, or inserts it if there's
// no saved state for this proposal.
// Returns an error if we couldn't talk to the database.
func (d *Database) UpdateProposalState(number int64, labels []string) error {
	logrus.WithFields(logrus.Fields{
		"number": number,
		"labels": labels,
	}).Debug("Updating proposal state")
	return d.proposalState.upsertState(number, labels)
}

// GetProposalState retrieves the state of a proposal. Returns an empty slice if
// no state has been saved for this proposal.
// Returns an error if we couldn't talk to the database.
func (d *Database) GetProposalState(number int64) ([]string, error) {
	logrus.WithFields(logrus.Fields{
		"number": number,
	}).Debug("Retrieving proposal state")
	return d.proposalState.selectState(number)
}
