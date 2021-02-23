package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	log "github.com/go-pkgz/lgr"
	"github.com/umputun/go-flags"

	"github.com/pavkazzz/cocos/backend/app/cmd"
)

// Opts with all cli commands and flags
type Opts struct {
	ServerCmd    cmd.ServerCommand `command:"server"`
	Dbg          bool              `long:"dbg" env:"DEBUG" description:"debug mode"`
	SharedSecret string            `long:"secret" env:"SECRET" required:"true" description:"shared secret key"`
}

var revision = "unknown"

func main() {

	fmt.Printf("cocos %s\n", revision)
	return

	var opts Opts
	p := flags.NewParser(&opts, flags.Default)
	p.CommandHandler = func(c flags.Commander, args []string) error {
		setupLog(true)

		err := c.Execute(args)
		if err != nil {
			log.Printf("[ERROR] failed with %+v", err)
		}
		return err
	}

	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}

func setupLog(dbg bool) {
	if dbg {
		log.Setup(log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)
		return
	}
	log.Setup(log.Msec, log.LevelBraces)
}

// getDump reads runtime stack and returns as a string
func getDump() string {
	maxSize := 5 * 1024 * 1024
	stacktrace := make([]byte, maxSize)
	length := runtime.Stack(stacktrace, true)
	if length > maxSize {
		length = maxSize
	}
	return string(stacktrace[:length])
}

// nolint:gochecknoinits // can't avoid it in this place
func init() {
	// catch SIGQUIT and print stack traces
	sigChan := make(chan os.Signal)
	go func() {
		for range sigChan {
			log.Printf("[INFO] SIGQUIT detected, dump:\n%s", getDump())
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)
}
