// bot/cron.go
package bot

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type CronService struct {
	cron *Cron
}

func NewCronService(bs *BotService, gh *GithubService) *CronService {
	return &CronService{
		cron: NewCron(bs, gh),
	}
}

func (cs *CronService) RegisterTasks() {
	cs.cron.RegisterTask(
		"update.stars.limit",
		DefaultSchedule,
		cs.cron.StarsLimitTask,
	)
}

func (cs *CronService) Start() {
	cs.cron.Start()
}

type Cron struct {
	scheduler *cron.Cron
	bot       *BotService
	gh        *GithubService
}

func NewCron(bs *BotService, gh *GithubService) *Cron {
	return &Cron{
		scheduler: cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
		bot:       bs,
		gh:        gh,
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
		log.Printf("task=%s started", name)
		if err := taskFunc(context.Background()); err != nil {
			log.Printf("task=%s failed: %v", name, err)
		} else {
			log.Printf("task=%s completed in %v", name, time.Since(t0))
		}
	})
	if err != nil {
		log.Printf("failed to register task %q: %v", name, err)
		return
	}
	log.Printf("task=%s registered, next run at %v", name, c.scheduler.Entry(id).Next)
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
