package bacula

import (
	"time"
	"errors"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var ErrUnknownDriver = errors.New("Unknown database driver.")

type DB struct {
	*sql.DB
	levelJobsStmt *sql.Stmt
	clientsStmt *sql.Stmt
}

// NewDB returns a new bacula database connection.
// You should have parseTime=true as parameter in your dataSourceName, eg.
// bacula:@tcp(127.0.0.1:3306)/bacula?parseTime=true
func NewDB(driverName, dataSourceName string) (*DB, error) {
	x, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	db := &DB{DB: x}

	err = db.initStmts(driverName)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Close closes the internal prepared statements and the database connection.
func (db *DB) Close() error {
	err := db.levelJobsStmt.Close()
	if err != nil {
		return err
	}

	err = db.clientsStmt.Close()
	if err != nil {
		return err
	}

	return db.DB.Close()
}

func (db *DB) initStmts(driverName string) error {
	switch driverName {
		case "mysql":
			return db.initMysqlStmts()
		default:
			return ErrUnknownDriver
	}
}

func (db *DB) initMysqlStmts() error {
	var err error
	db.levelJobsStmt, err = db.DB.Prepare("SELECT Level,RealEndTime FROM Job WHERE ClientID=? AND Level=? ORDER BY RealEndTime")
	if err != nil {
		return err
	}

	db.clientsStmt, err = db.DB.Prepare("SELECT ClientID, Name from Client")
	if err != nil {
		return err
	}

	return nil
}

// Client is a entry of baculas Client table.
type Client struct {
	ClientID	uint
	Name	string
}

// Job is a entry of baculas Job table.
type Job struct {
	Level	string
	RealEndTime	time.Time
}

// Clients retrieves all clients.
// If an error occurs while scanning columns, it continues with the next column, storing the error to be returned
// (this means that an returned error must not belong to the last row processed).
func (db *DB) Clients() ([]Client, error) {
	var err error
	cs := make([]Client, 0)

	rows, err := db.clientsStmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Client
		err2 := rows.Scan(&c.ClientID, &c.Name)
		if err2 != nil {
			err = err2
			continue
		}
		cs = append(cs, c)
	}

	return cs, err
}

// LevelJobs retrieves all Jobs for a Level ("F"ull, "I"ncremental, "D"ifferential) of a Client.
// If an error occurs while scanning columns, it continues with the next column, storing the error to be returned
// (this means that an returned error must not belong to the last row processed).
func (db *DB) LevelJobs(level string, c Client) ([]Job, error) {
	var err error
	js := make([]Job, 0)

	rows, err := db.levelJobsStmt.Query(c.ClientID, level)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var j Job
		err2 := rows.Scan(&j.Level, &j.RealEndTime)
		if err2 != nil {
			err = err2
			continue
		}

		js = append(js, j)
	}

	return js, err
}