package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"apisrv/pkg/app"
	"apisrv/pkg/db"

	"github.com/BurntSushi/toml"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/namsral/flag"
)

const appName = "apisrv"

var (
	fs                 = flag.NewFlagSetWithEnvPrefix(os.Args[0], "APISRV", 0)
	flConfigPath       = fs.String("config", "config.toml", "Path to config file")
	flVerbose          = fs.Bool("verbose", false, "enable debug output")
	flVerboseSql       = fs.Bool("verbose-sql", false, "enable all sql output")
	flGenerateTSClient = fs.Bool("ts_client", false, "generate TypeScript vt rpc client and exit")
	cfg                app.Config
	version            string
)

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.DefaultConfigFlagname = "config.flag"
	exitOnError(fs.Parse(os.Args[1:]))
	fixStdLog(*flVerbose)

	log.Printf("starting %v version=%v", appName, version)
	if _, err := toml.DecodeFile(*flConfigPath, &cfg); err != nil {
		exitOnError(err)
	}

	// enable sentry
	if cfg.Sentry.DSN != "" {
		exitOnError(sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.Sentry.DSN,
			Environment: cfg.Sentry.Environment,
			Release:     version,
		}))
	}

	// check db connection
	dbconn := pg.Connect(cfg.Database)
	dbc := db.New(dbconn)
	v, err := dbc.Version()
	exitOnError(err)
	log.Println(v)

	// log all sql queries
	if *flVerboseSql {
		sqlLogger := log.New(os.Stdout, "Q", log.LstdFlags)
		dbconn.AddQueryHook(db.NewQueryLogger(sqlLogger))
	}

	// create & run app
	application := app.New(appName, *flVerbose, cfg, dbc, dbconn)

	// enable vfs
	if cfg.Server.EnableVFS {
		err = application.RegisterVFS(cfg.VFS)
		exitOnError(err)
	}

	// generate TS client from cmd flags
	if *flGenerateTSClient {
		b, err := application.VTTypeScriptClient()
		exitOnError(err)
		fmt.Print(string(b))
		os.Exit(0)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Run
	go func() {
		if err := application.Run(); err != nil {
			exitOnError(err)
		}
	}()
	<-quit
	application.Shutdown(5 * time.Second)

}

// fixStdLog sets additional params to std logger (prefix D, filename & line).
func fixStdLog(verbose bool) {
	log.SetPrefix("D")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if verbose {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(io.Discard)
	}
}

// exitOnError calls log.Fatal if err wasn't nil.
func exitOnError(err error) {
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Fatal(err)
	}
}
