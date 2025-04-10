package bot

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type Cron struct {
	scheduler *cron.Cron
	bot       *BotService
}

func NewCron(bot *BotService) *Cron {
	return &Cron{
		scheduler: cron.New(),
		bot:       bot,
	}
}

func (c *Cron) RegisterTask(name string, schedule string, taskFunc func() error) {
	if schedule == "" {
		schedule = DefaultSchedule
	}
	id, err := c.scheduler.AddFunc(schedule, func() {
		t0 := time.Now()
		log.Printf("task=%s started", name)
		if err := taskFunc(); err != nil {
			log.Printf("task=%s failed: %v", name, err)
		} else {
			log.Printf("task=%s completed, duration=%v", name, time.Since(t0))
		}
	})
	if err != nil {
		log.Printf("failed to register task %s: %v", name, err)
		return
	}
	entry := c.scheduler.Entry(id)
	log.Printf("task=%s registered, next run at %v", name, entry.Next)
}

func (c *Cron) Start() {
	c.scheduler.Start()
}

func (c *Cron) StarsLimitTask() error {
	stars := getStarsCount()
	if stars > c.bot.limitByBack+10 {
		newLimit := stars + 10
		c.bot.SetLimitByBack(newLimit)
	}
	return nil
}
