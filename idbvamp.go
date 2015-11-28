package main

import (
	"log"
	"strings"
	"time"
	"git.office.bytemine.net/schuller/idbvamp/bacula"
	"git.office.bytemine.net/schuller/idbclient"
	"git.office.bytemine.net/schuller/wooly"

)

func main() {
	db, err := bacula.NewDB("mysql", "root:@tcp(127.0.0.1:3306)/bacula?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	idb := idbclient.NewIdb("idb-dev.office.bytemine.net", true)

	cs := make(chan bacula.Client)
	ms := make(chan idbclient.Machine)
	go clients(db, cs)
	go jobs(db, ms, cs)

	sendMachines(idb, ms)
}

func clients(db *bacula.DB, cs chan bacula.Client) {
	clients, err := db.Clients()
	if err != nil {
		close(cs)
	}

	for _, v := range clients {
		cs <- v
	}
	close(cs)
}

func jobs(db *bacula.DB, ms chan idbclient.Machine, cs chan bacula.Client) {
	for c := range(cs) {
		incJobs, err := db.LevelJobs("I", c)

		if err != nil {
			continue
		}

		diffJobs, err := db.LevelJobs("D", c)

		if err != nil {
			continue
		}

		fullJobs, err := db.LevelJobs("F", c)

		if err != nil {
			continue
		}

		var m idbclient.Machine
		m.Fqdn = strings.TrimRight(c.Name, "-fd")

		m.BackupBrand = idbclient.BackupBrandBacula

		if len(fullJobs) > 0 {
			m.BackupLastFullRun = wooly.New(fullJobs[len(fullJobs)-1].RealEndTime)
		}

		if len(incJobs) > 0 {
			m.BackupLastIncRun = wooly.New(incJobs[len(incJobs)-1].RealEndTime)
		}

		if len(diffJobs) > 0 {
			m.BackupLastDiffRun = wooly.New(diffJobs[len(diffJobs)-1].RealEndTime)
		}

		ms <- m

		log.Println(m)
	}

	close(ms)
}

func sendMachines(idb *idbclient.Idb, ms chan idbclient.Machine) {
	for m := range ms {
		_, err := idb.UpdateMachine(&m, true)
		log.Println(err)
	}
}