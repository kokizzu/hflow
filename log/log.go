// Package log provides a basic logging, modelled on stdlib's log package, with configurable verbosity
package log

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

var (
	verbosity int = 100 // Default to logging everything until a specific verbosity is set
)

func init() {
	log.SetPrefix("hflow ")
}

// SetVerbosity specifies the maximum verbosity of logs written
func SetVerbosity(v int) {
	Printf(0, "log verbosity set to [%v]", v)
	verbosity = v
}

// Printf writes s to the log formatted with args if the configured verbosity is <= to v
func Printf(v int, s string, args ...interface{}) {
	if v <= verbosity {
		log.Printf("lv="+strconv.Itoa(v)+" "+s+"\n", args...)
	}
}

// Fatalf calls Printf then exits the program with a return code of 1
func Fatalf(v int, s string, args ...interface{}) {
	Printf(v, s, args...)
	os.Exit(1)
}

// Panicf calls Printf then panics with the same information used for Printf
func Panicf(v int, s string, args ...interface{}) {
	Printf(v, s, args...)
	panic(fmt.Sprintf(s, args...))
}
