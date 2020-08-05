package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	dir    = "migrations/"
	ext    = ".sql"
	format = "20060102150405"
)

func createCmd(timestamp int64, name string) {
	var base string
	t := time.Unix(timestamp, 0)
	version := t.Format(format)
	base = fmt.Sprintf("%v%v_%v.", dir, version, name)

	os.MkdirAll(dir, os.ModePerm)
	createFile(base + "up" + ext)
}

func createFile(fname string) {
	if _, err := os.Create(fname); err != nil {
		log.fatalErr(err)
	}
}

func upCmd(m *migrate.Migrate, limit int) {
	if limit >= 0 {
		if err := m.Steps(limit); err != nil {
			if err != migrate.ErrNoChange {
				log.fatalErr(err)
			} else {
				log.Println(err)
			}
		}
	} else {
		if err := m.Up(); err != nil {
			if err != migrate.ErrNoChange {
				log.fatalErr(err)
			} else {
				log.Println(err)
			}
		}
	}
}
