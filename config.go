package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

const version = "0.0.5"
const exampleFilename = "idbvamp.json.example"

var configFile = flag.String("config", "/etc/bytemine/idbvamp.json", "config file")
var writeExample = flag.Bool("example", false, "write an example config to "+exampleFilename+" in the current dir.")
var showVersion = flag.Bool("version", false, "display version and exit")

func init() {
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *writeExample {
		c := example()
		c.write(exampleFilename)
		os.Exit(0)
	}
}

type config struct {
	// Create machines not existing in idb?
	Create bool

	// Data Source Name in the format like described in https://github.com/go-sql-driver/mysql#dsn-data-source-name . Example: root:@tcp(127.0.0.1:3306)/bacula?parseTime=true
	Dsn string

	// IDB URL, eg. https://idb.example.com
	Url string

	// IDB API Token
	ApiToken string

	// Allow invalid SSL certificate chains
	InsecureSkipVerify bool

	// Debugging
	Debug bool
}

func newConfig() *config {
	return &config{}
}

// Example returns an example config.
func example() *config {
	c := &config{}
	c.Create = false
	c.Dsn = "root:@tcp(127.0.0.1:3306)/bacula?parseTime=true"
	c.Url = "https://idb.example.com"
	c.ApiToken = "myVerySecretToken"
	c.InsecureSkipVerify = false
	c.Debug = false
	return c
}

// Load loads configuration values from a file.
func (c *config) load(file string) error {
	buf, err := ioutil.ReadFile(file)

	if err != nil {
		return err
	}

	err = json.Unmarshal(buf, c)
	if err != nil {
		return err
	}

	if c.Dsn == "" {
		return errors.New("Dsn can't be empty.")
	}

	if c.Url == "" {
		return errors.New("Url can't be empty")
	}

	if c.ApiToken == "" {
		return errors.New("ApiToken can't be empty")
	}

	return nil
}

// Write the JSON-marshaled config to a file.
func (c *config) write(file string) error {
	buf, err := json.MarshalIndent(c, "", "\t")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, buf, 0600)
	return err
}
