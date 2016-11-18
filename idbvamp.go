package main

import (
	"github.com/idb-project/idbclient"
	"github.com/idb-project/idbclient/machine"
	"github.com/idb-project/idbvamp/bacula"
	"log"
	"strings"
	"time"
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

	idb, err := idbclient.NewIdb(c.Url, c.ApiToken, c.InsecureSkipVerify)
	if err != nil {
		log.Fatal(err)
	}

	if c.Debug {
		idb.Debug = true
	}

	cs := make(chan bacula.Client)
	ms := make(chan machine.Machine)

	// error channel for the goroutines
	errs := make(chan error)

	// log all errors
	go logErrors(errs)
	go clients(db, errs, cs)
	go jobs(db, errs, ms, cs)
	sendMachines(idb, errs, ms, c.Create)
}

// logErrors logs all errors received on a chan error.
func logErrors(errs chan error) {
	for v := range errs {
		log.Println(v)
	}
}

// clients retrieves all clients from the bacula database and writes them to channel cs.
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

// jobs reads clients from channel cs and retrieves their jobs from the database, writing machines filled
// with the job data to channel ms.
func jobs(db *bacula.DB, errs chan error, ms chan machine.Machine, cs chan bacula.Client) {
	for c := range cs {
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

		var m machine.Machine

		var timeFull time.Time
		var timeInc time.Time
		var timeDiff time.Time

		if len(fullJobs) > 0 {
			timeFull = fullJobs[len(fullJobs)-1].RealEndTime
		}

		if len(incJobs) > 0 {
			timeInc = incJobs[len(incJobs)-1].RealEndTime
		}

		if len(diffJobs) > 0 {
			timeDiff = diffJobs[len(diffJobs)-1].RealEndTime
		}

		var sizeFull int64
		for _, v := range fullJobs {
			sizeFull += v.Bytes
		}

		var sizeInc int64
		for _, v := range incJobs {
			sizeInc += v.Bytes
		}

		var sizeDiff int64
		for _, v := range diffJobs {
			sizeDiff += v.Bytes
		}

		m.Backup(strings.TrimRight(c.Name, "-fd"), machine.BackupBrandBacula, timeFull, timeInc, timeDiff, sizeFull, sizeInc, sizeDiff)

		ms <- m
	}

	close(ms)
}

// sendMachines reads from channel ms and updates every machine in the idb.
func sendMachines(idb *idbclient.Idb, errs chan error, ms chan machine.Machine, create bool) {
	for m := range ms {
		_, err := idb.UpdateMachine(&m, create)
		if err != nil {
			log.Printf("%+v\n", m)
			errs <- err
		}
	}
}
