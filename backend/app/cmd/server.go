package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/pavkazzz/cocos/backend/app/api"
	"github.com/pavkazzz/cocos/backend/app/store/engine"
	"github.com/pavkazzz/cocos/backend/app/store/service"
)

// ServerCommand with command line flags and env
type ServerCommand struct {
	Store StoreGroup `group:"store" namespace:"store" env-namespace:"STORE"`

	Port           int    `long:"port" env:"APP_PORT" default:"8080" description:"port"`
	BackupLocation string `long:"backup" env:"BACKUP_PATH" default:"./var/backup" description:"backups location"`
	MaxBackupFiles int    `long:"max-back" env:"MAX_BACKUP_FILES" default:"10" description:"max backups to keep"`
	SharedSecret   string `long:"secret" env:"SECRET" required:"true" description:"shared secret key"`
}

// StoreGroup defines options group for store params
type StoreGroup struct {
	Bolt struct {
		Path    string        `long:"path" env:"PATH" default:"./var" description:"parent dir for bolt files"`
		Timeout time.Duration `long:"timeout" env:"TIMEOUT" default:"30s" description:"bolt timeout"`
	} `group:"bolt" namespace:"bolt" env-namespace:"BOLT"`
}

// serverApp holds all active objects
type serverApp struct {
	*ServerCommand
	restSrv *api.Rest
	// dataService *service.DataStore
	// authenticator *auth.Service
	terminated chan struct{}
}

// Execute is the entry point for "server" command, called by flag parser
func (s *ServerCommand) Execute(_ []string) error {
	log.Printf("[INFO] start server on port %d", s.Port)

	ctx, cancel := context.WithCancel(context.Background())
	go func() { // catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	app, err := s.newServerApp()
	if err != nil {
		log.Printf("[PANIC] failed to setup application, %+v", err)
		return err
	}
	if err = app.run(ctx); err != nil {
		log.Printf("[ERROR] cocos terminated with error %+v", err)
		return err
	}
	log.Printf("[INFO] cocos terminated")
	return nil
}

// newServerApp prepares application and return it with all active parts
// doesn't start anything
func (s *ServerCommand) newServerApp() (*serverApp, error) {

	if err := makeDirs(s.BackupLocation); err != nil {
		return nil, errors.Wrap(err, "failed to create backup store")
	}

	storeEngine, err := s.makeDataStore()
	if err != nil {
		return nil, errors.Wrap(err, "failed to make data store engine")
	}

	dataService := &service.DataStore{Engine: storeEngine}
	return &serverApp{
		restSrv: &api.Rest{
			Version:     "0.0.1",
			DataService: dataService,
		},
	}, nil
}

// makeDataStore creates store for all sites
func (s *ServerCommand) makeDataStore() (result engine.Interface, err error) {
	log.Printf("[INFO] make data store")

	if err = makeDirs(s.Store.Bolt.Path); err != nil {
		return nil, errors.Wrap(err, "failed to create bolt store")
	}
	sites := engine.BoltSite{FileName: fmt.Sprintf("backup/%s.db", s.Store.Bolt.Path)}
	return engine.NewBoltDB(bolt.Options{Timeout: s.Store.Bolt.Timeout}, sites)
}

// Run all application objects
func (a *serverApp) run(ctx context.Context) error {
	go func() {
		// shutdown on context cancellation
		<-ctx.Done()
		log.Print("[INFO] shutdown initiated")
		a.restSrv.Shutdown()
	}()

	// a.activateBackup(ctx) // runs in goroutine for each site

	a.restSrv.Run(a.Port)

	close(a.terminated)
	return nil
}

// activateBackup runs background backups for each site
// func (a *serverApp) activateBackup(ctx context.Context) {
// 	backup := migrator.AutoBackup{
// 		Exporter:       a.exporter,
// 		BackupLocation: a.BackupLocation,
// 		SiteID:         siteID,
// 		KeepMax:        a.MaxBackupFiles,
// 		Duration:       24 * time.Hour,
// 	}
// 	go backup.Do(ctx)
// }
