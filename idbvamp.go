package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	idbclient "github.com/idb-project/go-idb/client"
	"github.com/idb-project/go-idb/client/api"
	"github.com/idb-project/go-idb/models"
	"github.com/idb-project/idbvamp/bacula"
)

const backupBrandBacula = 1
const timeFormat = "2006-01-02 15:04:05"

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

	// Use a http.Client to enable ignoring of ssl verification errors
	httpclient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: c.InsecureSkipVerify}}}

	idburl, err := url.Parse(c.Url)
	if err != nil {
		log.Fatal(err)
	}

	address := fmt.Sprintf("%v:%v", idburl.Hostname(), idburl.Port())

	trans := client.NewWithClient(address, "/", []string{"http"}, httpclient)

	// Set up authentication in transport
	apiKeyHeaderAuth := client.APIKeyAuth("X-IDB-API-Token", "header", c.ApiToken)
	trans.DefaultAuthentication = apiKeyHeaderAuth

	// Set debug
	trans.Debug = c.Debug

	idb := idbclient.New(trans, strfmt.Default)

	cs := make(chan bacula.Client)
	ms := make(chan models.Machine)

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
func jobs(db *bacula.DB, errs chan error, ms chan models.Machine, cs chan bacula.Client) {
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

		var m models.Machine

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

		m.Fqdn = strings.TrimRight(c.Name, "-fd")
		m.BackupBrand = backupBrandBacula
		if !timeFull.IsZero() {
			m.BackupLastFullRun = timeFull.Format(timeFormat)
			m.BackupLastFullSize = sizeFull
		}

		if !timeInc.IsZero() {
			m.BackupLastIncRun = timeInc.Format(timeFormat)
			m.BackupLastIncSize = sizeInc
		}

		if !timeDiff.IsZero() {
			m.BackupLastDiffRun = timeDiff.Format(timeFormat)
			m.BackupLastDiffSize = sizeDiff
		}

		ms <- m
	}

	close(ms)
}

// sendMachines reads from channel ms and updates every machine in the idb.
func sendMachines(idb *idbclient.Idb, errs chan error, ms chan models.Machine, create bool) {
	for m := range ms {
		params := api.NewGetAPIV3MachinesRfqdnParams()
		params.SetRfqdn(m.Fqdn)
		_, err := idb.API.GetAPIV3MachinesRfqdn(params)

		switch {
		// machine not found, create
		case create && err != nil:
			params := api.NewPostAPIV3MachinesParams()
			params.SetMachine(&m)

			_, err := idb.API.PostAPIV3Machines(params)
			if err != nil {
				log.Fatal(err)
				errs <- err
				continue
			}
		// machine not found, but shouldn't create. skip to next machine
		case !create && err != nil:
			continue
		// machine found, update
		case create && err == nil || !create && err == nil:
			params := api.NewPutAPIV3MachinesRfqdnParams()
			params.SetMachine(&m)

			_, err := idb.API.PutAPIV3MachinesRfqdn(params)
			if err != nil {
				log.Fatal(err)
				errs <- err
				continue
			}
		}
	}
}
