package main

import (
	"fmt"
	logpkg "log"
	"os"
)

// Log for golang-migrate
type Log struct {
	verbose bool
}

// Printf for golang-migrate
func (l *Log) Printf(format string, v ...interface{}) {
	if l.verbose {
		logpkg.Printf(format, v...)
	} else {
		fmt.Fprintf(os.Stderr, format, v...)
	}
}

// Println for golang-migrate
func (l *Log) Println(args ...interface{}) {
	if l.verbose {
		logpkg.Println(args...)
	} else {
		fmt.Fprintln(os.Stderr, args...)
	}
}

// Verbose for golang-migrate
func (l *Log) Verbose() bool {
	return l.verbose
}

func (l *Log) fatalf(format string, v ...interface{}) {
	l.Printf(format, v...)
	os.Exit(1)
}

func (l *Log) fatal(args ...interface{}) {
	l.Println(args...)
	os.Exit(1)
}

func (l *Log) fatalErr(err error) {
	l.fatal("error:", err)
}
