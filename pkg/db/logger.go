package db

import (
	"context"
	"log"
	"time"

	"github.com/go-pg/pg/v10"
)

type QueryLogger struct {
	logger *log.Logger
}

func (ql QueryLogger) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	if event.Stash == nil {
		event.Stash = make(map[interface{}]interface{})
	}

	event.Stash["startedAt"] = time.Now()
	return ctx, nil
}

func (ql QueryLogger) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	query, err := event.FormattedQuery()
	if err != nil {
		ql.logger.Printf("formatted query err=%s", err)
	}

	var since time.Duration
	if event.Stash != nil {
		if v, ok := event.Stash["startedAt"]; ok {
			if startAt, ok := v.(time.Time); ok {
				since = time.Since(startAt)
			}
		}
	}

	ql.logger.Printf("query=%s duration=%v", query, since)
	return nil
}

func NewQueryLogger(logger *log.Logger) QueryLogger {
	return QueryLogger{logger: logger}
}
