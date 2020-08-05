package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
)

// set main log
var log = &Log{}

func main() {
	helpPtr := flag.Bool("help", false, "")
	verbosePtr := flag.Bool("verbose", false, "")
	prefetchPtr := flag.Uint("prefetch", 10, "")
	lockTimeoutPtr := flag.Uint("lock-timeout", 15, "")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr,
			`Usage: migrate OPTIONS COMMAND [arg...]
       migrate [ -help ]
Options:
  -prefetch N      Number of migrations to load in advance before executing (default 10)
  -lock-timeout N  Allow N seconds to acquire database lock (default 15)
  -verbose         Print verbose logging
  -help            Print usage
Commands:
  create [-ext E] [-dir D] [-seq] [-digits N] [-format] NAME
			   Create a set of timestamped up/down migrations titled NAME, in directory D with extension E.
			   Use -seq option to generate sequential up/down migrations with N digits.
			   Use -format option to specify a Go time format string.
  up [N]       Apply all or N up migrations
`)
	}

	flag.Parse()

	// initialize logger
	log.verbose = *verbosePtr

	// show help
	if *helpPtr {
		flag.Usage()
		os.Exit(0)
	}

	source := "file://migrations"

	content, err := ioutil.ReadFile("config/database.json")
	if err != nil {
		log.fatalErr(err)
	}

	var config map[string]interface{}
	json.Unmarshal(content, &config)

	port, ok := config["port"].(float64)
	if !ok {
		log.fatal("port must be number")
	}

	databaseURL := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s",
		config["user"],
		config["password"],
		config["host"],
		int(port),
		config["database"])

	log.Println(databaseURL)

	// initialize migrate
	// don't catch migraterErr here and let each command decide
	// how it wants to handle the error
	migrater, migraterErr := migrate.New(source, databaseURL)
	defer func() {
		if migraterErr == nil {
			migrater.Close()
		}
	}()
	if migraterErr == nil {
		migrater.Log = log
		migrater.PrefetchMigrations = *prefetchPtr
		migrater.LockTimeout = time.Duration(int64(*lockTimeoutPtr)) * time.Second

		// handle Ctrl+c
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT)
		go func() {
			for range signals {
				log.Println("Stopping after this running migration ...")
				migrater.GracefulStop <- true
				return
			}
		}()
	}

	startTime := time.Now()

	switch flag.Arg(0) {
	case "create":
		args := flag.Args()[1:]

		createFlagSet := flag.NewFlagSet("create", flag.ExitOnError)
		createFlagSet.Parse(args)

		if createFlagSet.NArg() == 0 {
			log.fatal("error: please specify name")
		}
		name := createFlagSet.Arg(0)

		timestamp := startTime.Unix()

		createCmd(timestamp, name)

	case "up":
		if migraterErr != nil {
			log.fatalErr(migraterErr)
		}

		limit := -1
		if flag.Arg(1) != "" {
			n, err := strconv.ParseUint(flag.Arg(1), 10, 64)
			if err != nil {
				log.fatal("error: can't read limit argument N")
			}
			limit = int(n)
		}

		upCmd(migrater, limit)

		if log.verbose {
			log.Println("Finished after", time.Now().Sub(startTime))
		}

	default:
		flag.Usage()
		os.Exit(0)
	}
}
