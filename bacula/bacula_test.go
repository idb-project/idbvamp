package bacula

/* These tests need a mysql bacula database with Client and Job tables running
on localhost. */

import (
	"testing"
)

func TestNewDB(t *testing.T) {
	_, err := NewDB("mysql", "root:@tcp(127.0.0.1:3306)/bacula")
	if err != nil {
		t.Error(err)
	}
}

func TestClients(t *testing.T) {
	db, err := NewDB("mysql", "root:@tcp(127.0.0.1:3306)/bacula")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	cs, err := db.Clients()
	if err != nil {
		t.Error(err)
	}

	t.Log(cs)
}

func TestLevelJobs(t *testing.T) {
	db, err := NewDB("mysql", "root:@tcp(127.0.0.1:3306)/bacula?parseTime=true")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	cs, err := db.Clients()
	if err != nil {
		t.Error(err)
	}

	js, err := db.LevelJobs("I", cs[0])
	if err != nil {
		t.Error(err)
	}

	t.Log(js)
}
