package app

import (
	"context"
	"time"

	"apisrv/pkg/db"
	"apisrv/pkg/embedlog"
	"apisrv/pkg/vt"

	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	"github.com/vmkteam/rpcgen/v2"
	"github.com/vmkteam/rpcgen/v2/typescript"
	"github.com/vmkteam/vfs"
	"github.com/vmkteam/zenrpc/v2"
)

type Config struct {
	Database *pg.Options
	Server   struct {
		Host      string
		Port      int
		IsDevel   bool
		EnableVFS bool
	}
	Sentry struct {
		Environment string
		DSN         string
	}
	VFS vfs.Config
}

type App struct {
	embedlog.Logger
	appName string
	cfg     Config
	db      db.DB
	dbc     *pg.DB
	echo    *echo.Echo
	vtsrv   zenrpc.Server
}

func New(appName string, verbose bool, cfg Config, db db.DB, dbc *pg.DB) *App {
	a := &App{
		appName: appName,
		cfg:     cfg,
		db:      db,
		dbc:     dbc,
		echo:    echo.New(),
	}
	a.SetStdLoggers(verbose)
	a.echo.HideBanner = true
	a.echo.HidePort = true
	a.echo.IPExtractor = echo.ExtractIPFromRealIPHeader()
	a.vtsrv = vt.New(a.db, a.Logger, a.cfg.Server.IsDevel)

	return a
}

// Run is a function that runs application.
func (a *App) Run() error {
	a.registerMetrics()
	a.registerHandlers()
	a.registerDebugHandlers()
	a.registerAPIHandlers()
	a.registerVTApiHandlers()
	return a.runHTTPServer(a.cfg.Server.Host, a.cfg.Server.Port)
}

// VTTypeScriptClient returns TypeScript client for VT.
func (a *App) VTTypeScriptClient() ([]byte, error) {
	gen := rpcgen.FromSMD(a.vtsrv.SMD())
	tsSettings := typescript.Settings{ExcludedNamespace: []string{NSVFS}, WithClasses: true}
	return gen.TSCustomClient(tsSettings).Generate()
}

// Shutdown is a function that gracefully stops HTTP server.
func (a *App) Shutdown(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := a.echo.Shutdown(ctx); err != nil {
		a.Errorf("shutting down server err=%q", err)
	}
}
