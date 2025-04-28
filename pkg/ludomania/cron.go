// ludomania/cron.go
package ludomania

import (
	"context"
	"github.com/go-pg/pg/v10"
	"gradebot/pkg/db"
	"gradebot/pkg/embedlog"
	"time"

	"github.com/robfig/cron/v3"
)

const (
	DefaultSchedule  = "*/1 * * * *"
	AdsLimitSchedule = "0 2 * * *"
)

type CronService struct {
	cron *Cron
}

func NewCronService(dbo db.DB, logger embedlog.Logger, bs *BotService, gh *GithubService) *CronService {
	return &CronService{
		cron: NewCron(dbo, logger, bs, gh),
	}
}

func (cs *CronService) RegisterTasks() {
	cs.cron.RegisterTask(
		"update.stars.limit",
		DefaultSchedule,
		cs.cron.StarsLimitTask,
	)
	cs.cron.RegisterTask(
		"update.ads.limit",
		AdsLimitSchedule,
		cs.cron.UpdateAdsLimits,
	)
}

func (cs *CronService) Start() {
	cs.cron.Start()
}

type Cron struct {
	embedlog.Logger
	scheduler *cron.Cron
	bot       *BotService
	gh        *GithubService

	db db.DB
	cr db.CommonRepo
}

func NewCron(dbo db.DB, logger embedlog.Logger, bs *BotService, gh *GithubService) *Cron {
	return &Cron{
		scheduler: cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
		Logger:    logger,
		bot:       bs,
		gh:        gh,
		db:        dbo,
		cr:        db.NewCommonRepo(dbo),
	}
}

func (c *Cron) RegisterTask(
	name, schedule string,
	taskFunc func(ctx context.Context) error,
) {
	if schedule == "" {
		schedule = DefaultSchedule
	}
	id, err := c.scheduler.AddFunc(schedule, func() {
		t0 := time.Now()
		c.Printf("task=%s started", name)
		if err := taskFunc(context.Background()); err != nil {
			c.Errorf("task=%s failed: %v", name, err)
		} else {
			c.Printf("task=%s completed in %v", name, time.Since(t0))
		}
	})
	if err != nil {
		c.Errorf("failed to register task %q: %v", name, err)
		return
	}
	c.Printf("task=%s registered, next run at %v", name, c.scheduler.Entry(id).Next)
}

func (c *Cron) Start() {
	c.scheduler.Start()
}

func (c *Cron) StarsLimitTask(ctx context.Context) error {
	stars, err := c.gh.GetStarsCount(ctx)
	if err != nil {
		return err
	}
	if stars > 0 {
		c.bot.SetLimitByBack(stars + 10)
	}
	return nil
}

func (c *Cron) UpdateAdsLimits(ctx context.Context) error {
	err := c.db.RunInLock(ctx, "adsWatched", func(tx *pg.Tx) error {
		_, err := tx.Exec(`update ludomans set "adsWatched" = 0;`)
		return err
	})
	if err != nil {
		c.Errorf("%v", err)
	}
	return err
}
