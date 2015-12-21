package main

import (
	"log"
	"strings"
	"git.office.bytemine.net/schuller/idbvamp/bacula"
	"git.office.bytemine.net/schuller/idbclient"
	"git.office.bytemine.net/schuller/wooly"

)

func main() {
	c := newConfig()
	err := c.load(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	db, err := bacula.NewDB("mysql", c.Dsn)
	if err != nil {
		log.Fatal(err)
	}

	idb := idbclient.NewIdb(c.Url, c.InsecureSkipVerify)
	if c.Debug {
		idb.Debug = true
	}

	cs := make(chan bacula.Client)
	ms := make(chan idbclient.Machine)
	errs := make(chan error)

	go logErrors(errs)
	go clients(db, errs, cs)
	go jobs(db, errs, ms, cs)
	sendMachines(idb, errs, ms)
}

func logErrors(errs chan error) {
	for v := range errs {
		log.Println(v)
	}
}

func clients(db *bacula.DB, errs chan error, cs chan bacula.Client) {
	clients, err := db.Clients()
	if err != nil {
		errs <- err
		close(cs)
	}

	for _, v := range clients {
		cs <- v
	}
	close(cs)
}

func jobs(db *bacula.DB, errs chan error, ms chan idbclient.Machine, cs chan bacula.Client) {
	for c := range(cs) {
		incJobs, err := db.LevelJobs("I", c)

		if err != nil {
			errs <- err
			continue
		}

		diffJobs, err := db.LevelJobs("D", c)

		if err != nil {
			errs <- err
			continue
		}

		fullJobs, err := db.LevelJobs("F", c)

		if err != nil {
			errs <- err
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
	}

	close(ms)
}

func sendMachines(idb *idbclient.Idb, errs chan error, ms chan idbclient.Machine) {
	for m := range ms {
		_, err := idb.UpdateMachine(&m, true)
		if err != nil {
			log.Printf("%+v\n", m)
			errs <- err
		}
	}
}