package bot

import (
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

func timerStarsCheck() {
	c := cron.New()

	type cronJob struct {
		schedule string
		fn       func() error
	}

	jobs := map[string]cronJob{
		"update.stars.limit": {
			schedule: "*/1 * * * *",
			fn: func() error {
				stars := getStarsCount()
				log.Printf("Получено количество звёзд: %d", stars)
				if stars > limitByBack+10 {
					limitByBack += 10
					log.Printf("Новый лимит: %d", limitByBack)
				}
				return nil
			},
		},
	}

	for name, job := range jobs {
		if job.schedule == "" {
			log.Printf("task=%v отключена конфигурацией", name)
			continue
		}

		name := name

		id, err := c.AddFunc(job.schedule, func() {
			t0 := time.Now()
			log.Printf("task=%v start running", name)
			if err := job.fn(); err != nil {
				log.Printf("task=%v run failed err=%v", name, err)
			} else {
				log.Printf("task=%v completed, duration=%v", name, time.Since(t0))
			}
		})
		if err != nil {
			log.Printf("task=%v failed to cron.AddFunc(), err=%v", name, err)
		} else {
			log.Printf("task=%v next run=%v", name, c.Entry(id).Next)
		}
	}

	c.Start()
}
