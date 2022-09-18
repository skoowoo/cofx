package crontrigger

import (
	"time"

	"github.com/robfig/cron/v3"
)

type CronTrigger struct {
	c *cron.Cron
}

func New() *CronTrigger {
	ct := &CronTrigger{
		c: cron.New(cron.WithSeconds()),
	}
	return ct
}

func (ct *CronTrigger) Add(format string, ch chan<- time.Time) (interface{}, error) {
	entityid, err := ct.c.AddFunc(format, func() {
		ch <- time.Now()
	})
	if err != nil {
		return nil, err
	}
	return entityid, nil
}

func (ct *CronTrigger) Remove(v interface{}) error {
	entityid := v.(cron.EntryID)
	ct.c.Remove(entityid)
	return nil
}

func (ct *CronTrigger) Start() {
	ct.c.Start()
}

func (ct *CronTrigger) Stop() {
	ct.c.Stop()
}
