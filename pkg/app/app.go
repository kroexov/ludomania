package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	botService "gradebot/pkg/bot"
	"gradebot/pkg/db"
	"gradebot/pkg/embedlog"

	"github.com/go-pg/pg/v10"
	"github.com/go-telegram/bot"
)

type Config struct {
	Database *pg.Options
	Server   struct {
		Host    string
		Port    int
		IsDevel bool
	}
	Bot struct {
		Token string
	}
}

type App struct {
	embedlog.Logger
	appName string
	cfg     Config
	db      db.DB
	b       *bot.Bot
	dbc     *pg.DB
	isDevel bool

	bs *botService.BotService
}

func New(appName string, verbose bool, cfg Config, db db.DB, dbc *pg.DB) *App {
	a := &App{
		appName: appName,
		cfg:     cfg,
		db:      db,
		dbc:     dbc,
		isDevel: cfg.Server.IsDevel,
	}

	a.SetStdLoggers(verbose)

	a.bs = botService.NewBotService(a.Logger, a.db)

	opts := []bot.Option{bot.WithDefaultHandler(a.bs.DefaultHandler)}
	b, err := bot.New(cfg.Bot.Token, opts...)
	if err != nil {
		panic(err)
	}
	a.b = b

	return a
}

// Run is a function that runs application.
func (a *App) Run() error {
	a.bs.RegisterBotHandlers(a.b)

	// for local usage
	if a.isDevel {
		go a.b.Start(context.TODO())
		return nil
	}

	// for server usage
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	_, err := a.b.SetWebhook(ctx, &bot.SetWebhookParams{
		URL: fmt.Sprintf("https://%s/isl/", a.cfg.Server.Host),
	})
	if err != nil {
		panic(err)
	}
	go a.b.StartWebhook(ctx)
	return http.ListenAndServe(fmt.Sprintf(":%d", a.cfg.Server.Port), a.b.WebhookHandler())
}

// Shutdown is a function that gracefully stops HTTP server.
func (a *App) Shutdown(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if _, err := a.b.Close(ctx); err != nil {
		a.Errorf("shutting down bot err=%q", err)
	}
}
